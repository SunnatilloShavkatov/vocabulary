package gatewayhttp

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

var (
	gatewayRequestsTotal   uint64
	gatewayRequests2xx     uint64
	gatewayRequests4xx     uint64
	gatewayRequests5xx     uint64
	gatewayUnauthorizedTotal uint64
)

func recordGatewayMetrics(status int) {
	atomic.AddUint64(&gatewayRequestsTotal, 1)
	switch {
	case status >= 500:
		atomic.AddUint64(&gatewayRequests5xx, 1)
	case status >= 400:
		atomic.AddUint64(&gatewayRequests4xx, 1)
	case status >= 200:
		atomic.AddUint64(&gatewayRequests2xx, 1)
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		atomic.AddUint64(&gatewayUnauthorizedTotal, 1)
	}
}

func GatewayMetricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(w,
		"# HELP gateway_http_requests_total Total HTTP requests handled by gateway.\n"+
			"# TYPE gateway_http_requests_total counter\n"+
			"gateway_http_requests_total %d\n"+
			"# HELP gateway_http_requests_2xx_total Total HTTP 2xx responses.\n"+
			"# TYPE gateway_http_requests_2xx_total counter\n"+
			"gateway_http_requests_2xx_total %d\n"+
			"# HELP gateway_http_requests_4xx_total Total HTTP 4xx responses.\n"+
			"# TYPE gateway_http_requests_4xx_total counter\n"+
			"gateway_http_requests_4xx_total %d\n"+
			"# HELP gateway_http_requests_5xx_total Total HTTP 5xx responses.\n"+
			"# TYPE gateway_http_requests_5xx_total counter\n"+
			"gateway_http_requests_5xx_total %d\n"+
			"# HELP gateway_auth_failures_total Total authentication/authorization failures.\n"+
			"# TYPE gateway_auth_failures_total counter\n"+
			"gateway_auth_failures_total %d\n",
		atomic.LoadUint64(&gatewayRequestsTotal),
		atomic.LoadUint64(&gatewayRequests2xx),
		atomic.LoadUint64(&gatewayRequests4xx),
		atomic.LoadUint64(&gatewayRequests5xx),
		atomic.LoadUint64(&gatewayUnauthorizedTotal),
	)
}
