package vocabularycontroller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"vocabulary/backend/modules/vocabulary/service"
)

type VocabularyHandler struct {
	service           *vocabularyservice.VocabularyService
	createdByResolver func(context.Context) (string, bool)
}

func RegisterVocabularyRoutes(
	mux *http.ServeMux,
	svc *vocabularyservice.VocabularyService,
	protected func(http.HandlerFunc) http.HandlerFunc,
	adminProtected func(http.HandlerFunc) http.HandlerFunc,
	createdByResolver func(context.Context) (string, bool),
) {
	h := &VocabularyHandler{service: svc, createdByResolver: createdByResolver}
	mux.HandleFunc("GET /v1/vocabulary", h.list)
	mux.HandleFunc("POST /v1/vocabulary", protected(h.create))
	mux.HandleFunc("GET /v1/admin/vocabulary", adminProtected(h.adminList))
	mux.HandleFunc("POST /v1/admin/vocabulary", adminProtected(h.adminCreate))
	mux.HandleFunc("GET /v1/admin/vocabulary/{id}", adminProtected(h.adminGet))
	mux.HandleFunc("PATCH /v1/admin/vocabulary/{id}", adminProtected(h.adminUpdate))
	mux.HandleFunc("DELETE /v1/admin/vocabulary/{id}", adminProtected(h.adminDelete))
	mux.HandleFunc("POST /v1/admin/vocabulary/{id}/approve", adminProtected(h.adminApprove))
	mux.HandleFunc("POST /v1/admin/vocabulary/{id}/reject", adminProtected(h.adminReject))
}

type vocabularyCreateRequest struct {
	Word        string `json:"word"`
	Translation string `json:"translation"`
	Example     string `json:"example"`
	Category    string `json:"category"`
	Status      string `json:"status"`
}

func (h *VocabularyHandler) create(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var req vocabularyCreateRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeVocabularyJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeVocabularyJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Word) == "" || strings.TrimSpace(req.Translation) == "" {
		writeVocabularyJSON(w, http.StatusBadRequest, map[string]string{"error": "word and translation are required"})
		return
	}

	createdBy := ""
	if h.createdByResolver != nil {
		if subject, ok := h.createdByResolver(r.Context()); ok {
			createdBy = subject
		}
	}

	item, err := h.service.Create(r.Context(), req.Word, req.Translation, req.Example, createdBy)
	if err != nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create vocabulary"})
		return
	}

	writeVocabularyJSON(w, http.StatusCreated, item)
}

type vocabularyListResponse struct {
	Items []vocabularyservice.VocabularyItem `json:"items"`
	Meta  vocabularyMetaInfo       `json:"meta"`
}

type vocabularyMetaInfo struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func (h *VocabularyHandler) list(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	q := r.URL.Query()
	search := strings.TrimSpace(q.Get("search"))
	page := parseVocabularyIntOr(q.Get("page"), 1)
	limit := parseVocabularyIntOr(q.Get("limit"), 20)

	items, total, err := h.service.List(r.Context(), search, page, limit)
	if err != nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list vocabulary"})
		return
	}
	if items == nil {
		items = []vocabularyservice.VocabularyItem{}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	writeVocabularyJSON(w, http.StatusOK, vocabularyListResponse{Items: items, Meta: vocabularyMetaInfo{Page: page, Limit: limit, Total: total}})
}

func (h *VocabularyHandler) adminList(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	q := r.URL.Query()
	items, total, err := h.service.AdminList(r.Context(), vocabularyservice.AdminVocabularySearch{
		Word:        q.Get("word"),
		Translation: q.Get("translation"),
		Category:    q.Get("category"),
		Status:      q.Get("status"),
		Page:        parseVocabularyIntOr(q.Get("page"), 1),
		Limit:       parseVocabularyIntOr(q.Get("limit"), 20),
	})
	if err != nil {
		if errors.Is(err, vocabularyservice.ErrInvalidVocabularyStatus) {
			writeVocabularyJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list vocabulary"})
		return
	}

	page := parseVocabularyIntOr(q.Get("page"), 1)
	limit := parseVocabularyIntOr(q.Get("limit"), 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	writeVocabularyJSON(w, http.StatusOK, vocabularyListResponse{Items: items, Meta: vocabularyMetaInfo{Page: page, Limit: limit, Total: total}})
}

func (h *VocabularyHandler) adminCreate(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var req vocabularyCreateRequest
	if ok := decodeVocabularyJSONBody(w, r, &req); !ok {
		return
	}

	createdBy := ""
	if h.createdByResolver != nil {
		if subject, ok := h.createdByResolver(r.Context()); ok {
			createdBy = subject
		}
	}

	item, err := h.service.AdminCreate(r.Context(), vocabularyservice.AdminVocabularyUpsertInput{
		Word:        req.Word,
		Translation: req.Translation,
		Example:     req.Example,
		Category:    req.Category,
		Status:      req.Status,
		CreatedBy:   createdBy,
	})
	if err != nil {
		switch {
		case errors.Is(err, vocabularyservice.ErrInvalidVocabularyInput), errors.Is(err, vocabularyservice.ErrInvalidVocabularyStatus):
			writeVocabularyJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create vocabulary"})
		}
		return
	}

	writeVocabularyJSON(w, http.StatusCreated, item)
}

func (h *VocabularyHandler) adminGet(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	item, err := h.service.AdminGet(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, vocabularyservice.ErrVocabularyNotFound) {
			writeVocabularyJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get vocabulary"})
		return
	}

	writeVocabularyJSON(w, http.StatusOK, item)
}

func (h *VocabularyHandler) adminUpdate(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var req vocabularyCreateRequest
	if ok := decodeVocabularyJSONBody(w, r, &req); !ok {
		return
	}

	item, err := h.service.AdminUpdate(r.Context(), r.PathValue("id"), vocabularyservice.AdminVocabularyUpsertInput{
		Word:        req.Word,
		Translation: req.Translation,
		Example:     req.Example,
		Category:    req.Category,
		Status:      req.Status,
	})
	if err != nil {
		switch {
		case errors.Is(err, vocabularyservice.ErrVocabularyNotFound):
			writeVocabularyJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, vocabularyservice.ErrInvalidVocabularyInput), errors.Is(err, vocabularyservice.ErrInvalidVocabularyStatus):
			writeVocabularyJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update vocabulary"})
		}
		return
	}

	writeVocabularyJSON(w, http.StatusOK, item)
}

func (h *VocabularyHandler) adminDelete(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	if err := h.service.AdminDelete(r.Context(), r.PathValue("id")); err != nil {
		if errors.Is(err, vocabularyservice.ErrVocabularyNotFound) {
			writeVocabularyJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete vocabulary"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *VocabularyHandler) adminApprove(w http.ResponseWriter, r *http.Request) {
	h.adminSetStatus(w, r, true)
}

func (h *VocabularyHandler) adminReject(w http.ResponseWriter, r *http.Request) {
	h.adminSetStatus(w, r, false)
}

func (h *VocabularyHandler) adminSetStatus(w http.ResponseWriter, r *http.Request, approve bool) {
	if h.service == nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var (
		item *vocabularyservice.VocabularyItem
		err  error
	)
	if approve {
		item, err = h.service.AdminApprove(r.Context(), r.PathValue("id"))
	} else {
		item, err = h.service.AdminReject(r.Context(), r.PathValue("id"))
	}
	if err != nil {
		if errors.Is(err, vocabularyservice.ErrVocabularyNotFound) {
			writeVocabularyJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update vocabulary status"})
		return
	}

	writeVocabularyJSON(w, http.StatusOK, item)
}

func decodeVocabularyJSONBody(w http.ResponseWriter, r *http.Request, out any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		writeVocabularyJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeVocabularyJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	return true
}

func parseVocabularyIntOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}

func writeVocabularyJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

