package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"vocabulary/backend/modules/vocabulary/service"
)

type VocabularyHandler struct {
	service *service.VocabularyService
}

func RegisterVocabularyRoutes(mux *http.ServeMux, svc *service.VocabularyService, protected func(http.HandlerFunc) http.HandlerFunc) {
	h := &VocabularyHandler{service: svc}
	mux.HandleFunc("GET /v1/vocabulary", h.list)
	mux.HandleFunc("POST /v1/vocabulary", protected(h.create))
}

type vocabularyCreateRequest struct {
	Word        string `json:"word"`
	Translation string `json:"translation"`
	Example     string `json:"example"`
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

	item, err := h.service.Create(r.Context(), req.Word, req.Translation, req.Example)
	if err != nil {
		writeVocabularyJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create vocabulary"})
		return
	}

	writeVocabularyJSON(w, http.StatusCreated, item)
}

type vocabularyListResponse struct {
	Items []service.VocabularyItem `json:"items"`
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
		items = []service.VocabularyItem{}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	writeVocabularyJSON(w, http.StatusOK, vocabularyListResponse{Items: items, Meta: vocabularyMetaInfo{Page: page, Limit: limit, Total: total}})
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

