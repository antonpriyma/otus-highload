package echoapi

import (
	"net/http"
	"sync"

	"github.com/labstack/echo"
)

type healthChecker struct {
	healthy bool
	rwMutex sync.RWMutex
}

func newHealthChecker(initialHealthyStatus bool) *healthChecker {
	return &healthChecker{healthy: initialHealthyStatus}
}

func (h *healthChecker) HTTPHandler(ectx echo.Context) error {
	h.rwMutex.RLock()
	healthy := h.healthy
	h.rwMutex.RUnlock()

	if healthy {
		return ectx.NoContent(http.StatusOK)
	}

	return ectx.NoContent(http.StatusServiceUnavailable)
}

func (h *healthChecker) ChangeHealthStatus(healthy bool) {
	h.rwMutex.Lock()
	defer h.rwMutex.Unlock()

	h.healthy = healthy
}
