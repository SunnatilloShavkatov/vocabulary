package notificationcontroller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"vocabulary/backend/modules/notification/service"
)

type NotificationHandler struct {
	service *notificationservice.NotificationService
}

func RegisterNotificationRoutes(mux *http.ServeMux, svc *notificationservice.NotificationService, protected func(http.HandlerFunc) http.HandlerFunc) {
	h := &NotificationHandler{service: svc}
	mux.HandleFunc("POST /internal/notifications/word-added", protected(h.onWordAdded))
}

func (h *NotificationHandler) onWordAdded(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotificationJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var req notificationservice.WordAddedEvent
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeNotificationJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeNotificationJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	schedules, err := h.service.OnWordAdded(r.Context(), req)
	if err != nil {
		if errors.Is(err, notificationservice.ErrInvalidWordAddedEvent) {
			writeNotificationJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeNotificationJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process WordAdded event"})
		return
	}

	writeNotificationJSON(w, http.StatusAccepted, map[string]any{"schedules": schedules})
}
