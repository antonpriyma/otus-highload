package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/cert"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/utils"
	"github.com/rs/xid"
)

type Config struct {
	ConnectTimeout   time.Duration `mapstructure:"connect_timeout"`
	ReadWriteTimeout time.Duration `mapstructure:"read_write_timeout"`

	KeepAlive time.Duration `mapstructure:"keep_alive"`

	MaxIdleConns        int `mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int `mapstructure:"max_idle_conns_per_host"`

	LogParams LogParams `mapstructure:"log_params"`

	Proxy ProxyConfig `mapstructure:"proxy"`

	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`

	Cert cert.StorageConfig `mapstructure:"cert"`
}

type ProxyConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username" json:"-"`
	Password string `mapstructure:"password" json:"-"`
}

type LogParams struct {
	// params that will be replaced with <secret> in logs
	SecretURLQueryParams []string `mapstructure:"secret_url_query_params"`
	// Both Request and Response headers
	SecretHeaders   []string `mapstructure:"secret_headers"`
	LogConnectTrace bool     `mapstructure:"log_connect_trace"`
	Debug           bool     `mapstructure:"debug"`
	Headers         []string `mapstructure:"headers"`
}

func (c Config) fillCerts(tlsConfig *tls.Config) error {
	if c.Cert.IsEmpty() {
		return nil
	}

	certStorage, err := cert.NewStorage(c.Cert)
	if err != nil {
		return errors.Wrap(err, "failed to init cert storage")
	}

	tlsConfig.Certificates = []tls.Certificate{certStorage.Cert()}
	tlsConfig.RootCAs = certStorage.CA()

	return nil
}

const (
	secretPlaceholder = "secret"

	headerRequestID = "X-Request-ID"
)

var (
	defaultLogHeaders = []string{headerRequestID}
)

type simpleHTTPClient struct {
	Client http.Client

	Logger    log.Logger
	LogParams LogParams

	ProxyBasicAuth string
}

func NewSimpleHTTPClient(
	cfg Config,
	logger log.Logger,
) (Client, error) {
	dialer := net.Dialer{
		Timeout: cfg.ConnectTimeout,
	}

	tlsConfig := tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	transport := http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.KeepAlive,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tlsConfig,
	}

	err := cfg.fillCerts(&tlsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fill client certs")
	}

	if cfg.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	var proxyBasicAuth string
	if cfg.Proxy.URL != "" {
		proxyURL, err := url.Parse(cfg.Proxy.URL)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse proxy url %q", cfg.Proxy.URL)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		if cfg.Proxy.Username != "" && cfg.Proxy.Password != "" {
			proxyBasicAuth = getProxyBasicAuth(cfg.Proxy.Username, cfg.Proxy.Password)
		}
	}

	httpClient := http.Client{Transport: &transport, Timeout: cfg.ReadWriteTimeout}

	cfg.LogParams.Headers = utils.MergeStringSlices(cfg.LogParams.Headers, defaultLogHeaders)

	return &simpleHTTPClient{
		Client:         httpClient,
		Logger:         logger,
		LogParams:      cfg.LogParams,
		ProxyBasicAuth: proxyBasicAuth,
	}, nil
}

func getProxyBasicAuth(username, password string) string {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return fmt.Sprintf("Basic %s", auth)
}

func (c *simpleHTTPClient) PerformRequest(ctx context.Context, req Request, res Response) (err error) {
	requestStartTime := time.Now()

	var connectTrace traceDumper = dummyTracer{}
	if c.LogParams.LogConnectTrace {
		ctx, connectTrace = contextWithConnectTrace(ctx)
	}

	httpReq, err := buildHTTPRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to convert request model to http request")
	}

	logFields := log.Fields{}
	if c.LogParams.Debug && httpReq.Body != nil {
		var reqBody []byte
		reqBody, err = ioutil.ReadAll(httpReq.Body)
		if err != nil {
			return errors.Wrap(err, "failed to read request body")
		}
		httpReq.Body.Close()
		logFields["request_body"] = string(reqBody)
		httpReq.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
	}

	if c.ProxyBasicAuth != "" {
		httpReq.Header.Set("Proxy-Authorization", c.ProxyBasicAuth)
	}

	var httpResp *http.Response
	defer func() {
		c.logRequest(ctx, httpReq, httpResp, requestStartTime, logFields)
	}()

	httpResp, err = c.Client.Do(httpReq)
	if err != nil {
		connectTrace.DumpTrace(ctx, c.Logger)
		return errors.Wrap(err, "failed to do request")
	}

	if c.LogParams.Debug {
		var respBody []byte
		respBody, err = ioutil.ReadAll(httpResp.Body)
		if err != nil {
			return errors.Wrap(err, "failed to read response body")
		}
		httpResp.Body.Close()

		logFields["response_body"] = string(respBody)
		httpResp.Body = ioutil.NopCloser(bytes.NewReader(respBody))
	}
	defer httpResp.Body.Close()

	return res.ReadFrom(httpResp)
}

func (c *simpleHTTPClient) logRequest(
	ctx context.Context,
	req *http.Request,
	resp *http.Response,
	requestStartTime time.Time,
	logFields log.Fields,
) {
	statusCode := -1
	respHeader := http.Header{}
	if resp != nil {
		statusCode = resp.StatusCode
		respHeader = resp.Header
	}

	safeURL := c.maskSecretParams(*req.URL)
	urlToLog, err := url.QueryUnescape(safeURL.String())
	if err != nil {
		urlToLog = safeURL.String()
	}

	logFields.Extend(log.Fields{
		"client_request_method":    req.Method,
		"client_request_url":       urlToLog,
		"client_request_resp_code": statusCode,
		"client_request_duration":  time.Since(requestStartTime),
	})

	logFields.Extend(log.Fields{
		"request_headers":  c.extractLogHeaders(req.Header),
		"response_headers": c.extractLogHeaders(respHeader),
	})

	c.Logger.ForCtx(ctx).WithFields(logFields).Info("request")
}

// extractLogHeaders makes http.Header map to log
// with masked secret headers (LogParams.SecretHeaders)
func (c *simpleHTTPClient) extractLogHeaders(header http.Header) (ret http.Header) {
	ret = c.filterHeadersToLog(header)

	for _, secretHeader := range c.LogParams.SecretHeaders {
		if _, ok := ret[secretHeader]; ok {
			ret.Set(secretHeader, secretPlaceholder)
		}
	}

	return ret
}

// filterHeadersToLog create http.Header map with values than should be logged
// based on LogParams.Debug and LogParams.Headers config values
func (c *simpleHTTPClient) filterHeadersToLog(header http.Header) (ret http.Header) {
	if c.LogParams.Debug {
		// copy all params
		ret = make(http.Header, len(header))
		for headerName, val := range header {
			ret[headerName] = val
		}
		return ret
	}

	ret = make(http.Header, len(c.LogParams.Headers))
	for _, headerName := range c.LogParams.Headers {
		val := header.Get(headerName)
		if val != "" {
			ret.Set(headerName, val)
		}
	}

	return ret
}

func (c *simpleHTTPClient) maskSecretParams(u url.URL) url.URL {
	params := u.Query()

	for _, k := range c.LogParams.SecretURLQueryParams {
		if _, ok := params[k]; ok {
			params.Set(k, secretPlaceholder)
		}
	}

	u.RawQuery = params.Encode()

	return u
}

func buildHTTPRequest(ctx context.Context, req Request) (*http.Request, error) {
	var reqBody io.Reader

	if reqWithBody, ok := req.(RequestWithBody); ok {
		reqBodyRaw, err := reqWithBody.Body()
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch request body")
		}

		reqBody = bytes.NewReader(reqBodyRaw)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method(), req.URL(), reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request with context")
	}

	if reqWithRequestID, ok := req.(RequestWithRequestID); ok {
		reqID := reqWithRequestID.RequestID()
		if reqID == "" {
			reqID = xid.New().String()
		}
		httpReq.Header.Set(headerRequestID, reqID)
	}

	if reqWithHeaders, ok := req.(RequestWithHeaders); ok {
		reqHeaders := reqWithHeaders.Headers()
		for k := range reqHeaders {
			httpReq.Header.Set(k, reqHeaders.Get(k))
		}
	}

	return httpReq, nil
}
