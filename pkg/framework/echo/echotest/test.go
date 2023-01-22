package echotest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
)

type ContextParams struct {
	RequestBody    io.Reader
	RequestHeaders map[string]string
	PathParams     map[string]string
	QueryParams    map[string]string
}

func NewContext(params ContextParams) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api", params.RequestBody)
	for k, v := range params.RequestHeaders {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	ectx := e.NewContext(req, rec)

	keys := make([]string, len(params.PathParams))
	i := 0
	for k := range params.PathParams {
		keys[i] = k
		i++
	}

	values := make([]string, len(params.PathParams))
	i = 0
	for _, v := range params.PathParams {
		values[i] = v
		i++
	}

	ectx.SetParamNames(keys...)
	ectx.SetParamValues(values...)

	if len(params.QueryParams) > 0 {
		query := url.Values{}
		for k, v := range params.QueryParams {
			query.Set(k, v)
		}

		ectx.Request().URL.RawQuery = query.Encode()
	}

	ctx := req.Context()
	ectx.SetRequest(ectx.Request().WithContext(ctx))
	echoutils.StoreContext(ctx, ectx) // compatibility mode

	return ectx, rec
}

type Case struct {
	RequestQueryParams map[string]string
	RequestPathParams  map[string]string
	Request            interface{}
	RequestHeaders     map[string]string
	Response           interface{}

	ExactError   error
	WantAnyError bool
}

type CaseTools struct {
	Context     context.Context
	EchoContext echo.Context
	GomockCtrl  *gomock.Controller
}

func (c Case) RunWrapped(t *testing.T, name string, f func(tool CaseTools) error) {
	t.Run(name, func(t *testing.T) {
		c.Run(t, f)
	})
}

func (c Case) Run(t *testing.T, f func(tool CaseTools) error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqBody, err := extractBody(c.Request)
	require.NoError(t, err, "failed to extract request body")

	req := bytes.NewBuffer([]byte(reqBody))
	ectx, rec := NewContext(ContextParams{
		PathParams:     c.RequestPathParams,
		QueryParams:    c.RequestQueryParams,
		RequestBody:    req,
		RequestHeaders: c.RequestHeaders,
	})
	ctx := echoutils.MustGetContext(ectx)

	err = f(CaseTools{
		Context:     ctx,
		EchoContext: ectx,
		GomockCtrl:  ctrl,
	})
	test.CheckError(t, err, c.ExactError, c.WantAnyError)

	respBody, err := extractBody(c.Response)
	require.NoError(t, err, "failed to extract response body")
	require.Equal(t, respBody, strings.TrimSpace(rec.Body.String()))
}

func extractBody(i interface{}) (string, error) {
	if i == nil {
		// will check that body is empty
		return "", nil
	}

	switch data := i.(type) {
	case string:
		return data, nil
	case []byte:
		return string(data), nil
	}

	data, err := json.Marshal(i)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal object to json")
	}

	return string(data), nil
}
