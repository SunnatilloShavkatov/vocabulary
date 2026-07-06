package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

var (
	authRequestsTotal         uint64
	authRequests2xxTotal      uint64
	authRequests4xxTotal      uint64
	authRequests5xxTotal      uint64
	authLoginRequestsTotal    uint64
	authLoginSuccessTotal     uint64
	authLoginFailureTotal     uint64
	authCreateAdminTotal      uint64
	authCreateAdminDeniedTotal uint64
)

func recordAuthMetrics(path string, status int) {
	atomic.AddUint64(&authRequestsTotal, 1)

	switch {
	case status >= 500:
		atomic.AddUint64(&authRequests5xxTotal, 1)
	case status >= 400:
		atomic.AddUint64(&authRequests4xxTotal, 1)
	case status >= 200:
		atomic.AddUint64(&authRequests2xxTotal, 1)
	}

	if path == "/v1/auth/login" {
		atomic.AddUint64(&authLoginRequestsTotal, 1)
		if status >= 200 && status < 300 {
			atomic.AddUint64(&authLoginSuccessTotal, 1)
		} else {
			atomic.AddUint64(&authLoginFailureTotal, 1)
		}
	}

	if path == "/v1/admins" {
		atomic.AddUint64(&authCreateAdminTotal, 1)
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			atomic.AddUint64(&authCreateAdminDeniedTotal, 1)
		}
	}
}

func authMetricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(
		w,
		"# HELP auth_http_requests_total Total HTTP requests handled by auth-service.\n"+
			"# TYPE auth_http_requests_total counter\n"+
			"auth_http_requests_total %d\n"+
			"# HELP auth_http_requests_2xx_total Total HTTP 2xx responses.\n"+
			"# TYPE auth_http_requests_2xx_total counter\n"+
			"auth_http_requests_2xx_total %d\n"+
			"# HELP auth_http_requests_4xx_total Total HTTP 4xx responses.\n"+
			"# TYPE auth_http_requests_4xx_total counter\n"+
			"auth_http_requests_4xx_total %d\n"+
			"# HELP auth_http_requests_5xx_total Total HTTP 5xx responses.\n"+
			"# TYPE auth_http_requests_5xx_total counter\n"+
			"auth_http_requests_5xx_total %d\n"+
			"# HELP auth_login_requests_total Total login requests.\n"+
			"# TYPE auth_login_requests_total counter\n"+
			"auth_login_requests_total %d\n"+
			"# HELP auth_login_success_total Total successful login requests.\n"+
			"# TYPE auth_login_success_total counter\n"+
			"auth_login_success_total %d\n"+
			"# HELP auth_login_failure_total Total failed login requests.\n"+
			"# TYPE auth_login_failure_total counter\n"+
			"auth_login_failure_total %d\n"+
			"# HELP auth_create_admin_total Total create-admin requests.\n"+
			"# TYPE auth_create_admin_total counter\n"+
			"auth_create_admin_total %d\n"+
			"# HELP auth_create_admin_denied_total Total denied create-admin requests.\n"+
			"# TYPE auth_create_admin_denied_total counter\n"+
			"auth_create_admin_denied_total %d\n",
		atomic.LoadUint64(&authRequestsTotal),
		atomic.LoadUint64(&authRequests2xxTotal),
		atomic.LoadUint64(&authRequests4xxTotal),
		atomic.LoadUint64(&authRequests5xxTotal),
		atomic.LoadUint64(&authLoginRequestsTotal),
		atomic.LoadUint64(&authLoginSuccessTotal),
		atomic.LoadUint64(&authLoginFailureTotal),
		atomic.LoadUint64(&authCreateAdminTotal),
		atomic.LoadUint64(&authCreateAdminDeniedTotal),
	)
}

type authStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *authStatusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func withAuthMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &authStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		recordAuthMetrics(r.URL.Path, rec.status)
	})
}
