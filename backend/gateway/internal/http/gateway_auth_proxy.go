package gatewayhttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"vocabulary/backend/gateway/internal/grpcclient"
	authv1 "vocabulary/backend/proto/auth/v1"
)

type gatewayAuthProxyHandler struct {
	client grpcclient.AuthServiceClient
}

func newGatewayAuthProxyHandler(client grpcclient.AuthServiceClient) *gatewayAuthProxyHandler {
	return &gatewayAuthProxyHandler{client: client}
}

func (h *gatewayAuthProxyHandler) login(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.client == nil {
		writeGatewayJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth upstream is not configured"})
		return
	}

	var req authv1.LoginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeGatewayJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeGatewayJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.client.Login(r.Context(), req)
	if err != nil {
		writeGatewayAuthGRPCError(w, err)
		return
	}
	writeGatewayJSON(w, http.StatusOK, resp)
}

func (h *gatewayAuthProxyHandler) createAdmin(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.client == nil {
		writeGatewayJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth upstream is not configured"})
		return
	}

	var req authv1.CreateAdminRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeGatewayJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeGatewayJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.client.CreateAdmin(r.Context(), req)
	if err != nil {
		writeGatewayAuthGRPCError(w, err)
		return
	}
	writeGatewayJSON(w, http.StatusCreated, resp)
}

func writeGatewayAuthGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeGatewayJSON(w, http.StatusBadGateway, map[string]string{"error": "auth upstream request failed"})
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		writeGatewayJSON(w, http.StatusBadRequest, map[string]string{"error": st.Message()})
	case codes.Unauthenticated:
		writeGatewayJSON(w, http.StatusUnauthorized, map[string]string{"error": st.Message()})
	case codes.PermissionDenied:
		writeGatewayJSON(w, http.StatusForbidden, map[string]string{"error": st.Message()})
	case codes.AlreadyExists:
		writeGatewayJSON(w, http.StatusConflict, map[string]string{"error": st.Message()})
	case codes.FailedPrecondition:
		writeGatewayJSON(w, http.StatusServiceUnavailable, map[string]string{"error": st.Message()})
	case codes.Unavailable:
		writeGatewayJSON(w, http.StatusBadGateway, map[string]string{"error": "auth upstream unavailable"})
	default:
		writeGatewayJSON(w, http.StatusInternalServerError, map[string]string{"error": "auth upstream internal error"})
	}
}
