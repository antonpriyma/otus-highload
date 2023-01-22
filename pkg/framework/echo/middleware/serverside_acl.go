package middleware

import (
	"context"
	"net"
	"net/http"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoerrors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/utils"

	"github.com/labstack/echo"
)

type ServersideACLConfig struct {
	HeaderName string    `mapstructure:"header_name"`
	Enabled    bool      `mapstructure:"enabled"`
	ACL        []ACLNode `mapstructure:"acl"`
}

func NewServersideACLMiddleware(cfg ServersideACLConfig, logger log.Logger) (echo.MiddlewareFunc, error) {
	err := cfg.prepare()
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare config")
	}

	return serversideACLMiddleware{
		Config: cfg,
		Logger: logger,
	}.MiddlewareFunc, nil
}

func (c *ServersideACLConfig) prepare() error {
	for i := range c.ACL {
		err := c.ACL[i].prepareNetList()
		if err != nil {
			return errors.Wrap(err, "failed to prepare config")
		}
	}

	return nil
}

type serversideACLMiddleware struct {
	Config ServersideACLConfig
	Logger log.Logger
}

func (m serversideACLMiddleware) MiddlewareFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if m.Config.Enabled {
			if !m.isRequestAllowed(c) {
				return echoerrors.UnauthorizedError(
					errors.New("invalid serverside token"),
					echoerrors.ReasonServersideTokenInvalid,
					"invalid serverside token",
				)
			}
		}

		return next(c)
	}
}

func (m serversideACLMiddleware) extractIP(ctx context.Context, req *http.Request) net.IP {
	ipStr := req.Header.Get(echoutils.HeaderUserIP)
	if ipStr == "" {
		m.Logger.ForCtx(ctx).Info("empty ip from header, fallback to remote addr")
		ipStr, _ = utils.Split2(req.RemoteAddr, ":")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		m.Logger.ForCtx(ctx).WithFields(log.Fields{"raw_ip": ipStr}).Error("failed to parse ip")
		return nil
	}

	return ip
}

func (m serversideACLMiddleware) isRequestAllowed(c echo.Context) bool {
	ctx := c.Request().Context()
	headers := c.Request().Header

	ip := m.extractIP(ctx, c.Request())
	if ip == nil {
		return false
	}

	token := headers.Get(m.Config.HeaderName)
	for _, node := range m.Config.ACL {
		if node.isRequestAllowed(ip, token) {
			ctx = log.AddCtxFields(ctx, log.Fields{
				"token_owner": node.Owner,
			})
			echoutils.StoreContext(ctx, c)

			return true
		}
	}

	m.Logger.ForCtx(ctx).WithField("ip", ip.String()).Warn("request is not allowed")

	return false
}

type ACLNode struct {
	Owner      string   `mapstructure:"owner"`
	Token      string   `mapstructure:"token" json:"-"`
	RawNetList []string `mapstructure:"net_list"`
	netList    NetList
}

func (n ACLNode) isRequestAllowed(ip net.IP, token string) bool {
	return token == n.Token && n.netList.Contains(ip)
}

func (n *ACLNode) prepareNetList() error {
	netList := make(NetList, 0, len(n.RawNetList))
	for _, rawNet := range n.RawNetList {
		_, net, err := net.ParseCIDR(rawNet)
		if err != nil {
			return errors.Wrapf(err, "failed to parse node net %q", rawNet)
		}

		netList = append(netList, *net)
	}

	n.netList = netList
	return nil
}

type NetList []net.IPNet

func (l NetList) Contains(ip net.IP) bool {
	for _, net := range l {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}
