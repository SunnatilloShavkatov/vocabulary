package userscontroller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"vocabulary/backend/modules/users/service"
)

const (
	usersHeaderUserID = "X-User-ID"
	usersHeaderRole   = "X-User-Role"
)

type UsersHandler struct {
	service *usersservice.UsersService
}

func RegisterUsersRoutes(
	mux *http.ServeMux,
	svc *usersservice.UsersService,
	protected func(http.HandlerFunc) http.HandlerFunc,
	adminProtected func(http.HandlerFunc) http.HandlerFunc,
) {
	h := &UsersHandler{service: svc}
	mux.HandleFunc("GET /v1/users/me", protected(h.me))
	mux.HandleFunc("GET /v1/admin/users", adminProtected(h.adminListUsers))
	mux.HandleFunc("POST /v1/admin/users", adminProtected(h.adminCreateUser))
	mux.HandleFunc("GET /v1/admin/users/{id}", adminProtected(h.adminGetUser))
	mux.HandleFunc("PATCH /v1/admin/users/{id}", adminProtected(h.adminUpdateUser))
	mux.HandleFunc("DELETE /v1/admin/users/{id}", adminProtected(h.adminDeleteUser))
	mux.HandleFunc("POST /v1/admin/users/{id}/block", adminProtected(h.adminBlockUser))
	mux.HandleFunc("POST /v1/admin/users/{id}/unblock", adminProtected(h.adminUnblockUser))
	mux.HandleFunc("GET /v1/admin/stats", adminProtected(h.adminStats))
	mux.HandleFunc("GET /v1/admin/backup/export", adminProtected(h.adminExport))
	mux.HandleFunc("POST /v1/admin/backup/export-jobs", adminProtected(h.adminStartExportJob))
	mux.HandleFunc("GET /v1/admin/backup/export-jobs", adminProtected(h.adminListExportJobs))
	mux.HandleFunc("GET /v1/admin/backup/export-jobs/{id}", adminProtected(h.adminGetExportJob))
	mux.HandleFunc("GET /v1/admin/backup/export-jobs/{id}/download", adminProtected(h.adminDownloadExportJob))
	mux.HandleFunc("POST /v1/admin/audit-logs", adminProtected(h.adminWriteAuditLog))
	mux.HandleFunc("GET /v1/admin/audit-logs", adminProtected(h.adminListAuditLogs))
}

func (h *UsersHandler) me(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	userID, role, ok := usersIdentityFromHeaders(r)
	if !ok {
		writeUsersJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing forwarded user identity"})
		return
	}

	profile, err := h.service.Me(r.Context(), userID, role)
	if err != nil {
		if errors.Is(err, usersservice.ErrInvalidIdentity) {
			writeUsersJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load profile"})
		return
	}

	writeUsersJSON(w, http.StatusOK, profile)
}

func usersIdentityFromHeaders(r *http.Request) (string, string, bool) {
	if r == nil {
		return "", "", false
	}
	userID := strings.TrimSpace(r.Header.Get(usersHeaderUserID))
	role := strings.TrimSpace(r.Header.Get(usersHeaderRole))
	if userID == "" || role == "" {
		return "", "", false
	}
	return userID, role, true
}

type adminUserUpsertRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *UsersHandler) adminListUsers(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	items, err := h.service.AdminList(r.Context(), strings.TrimSpace(r.URL.Query().Get("q")))
	if err != nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}

	writeUsersJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *UsersHandler) adminCreateUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var req adminUserUpsertRequest
	if ok := decodeUsersJSONBody(w, r, &req); !ok {
		return
	}

	item, err := h.service.AdminCreate(r.Context(), usersservice.AdminCreateUserInput{Name: req.Name, Email: req.Email})
	if err != nil {
		switch {
		case errors.Is(err, usersservice.ErrInvalidAdminUserInput):
			writeUsersJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, usersservice.ErrAdminUserAlreadyExists):
			writeUsersJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		}
		return
	}

	writeUsersJSON(w, http.StatusCreated, item)
	h.writeAuditTrail(r, "admin_user_created", "user", item.ID, map[string]string{"email": item.Email})
}

func (h *UsersHandler) adminGetUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	item, err := h.service.AdminGet(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, usersservice.ErrAdminUserNotFound) {
			writeUsersJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
		return
	}

	writeUsersJSON(w, http.StatusOK, item)
	h.writeAuditTrail(r, "admin_user_updated", "user", item.ID, map[string]string{"email": item.Email})
}

func (h *UsersHandler) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var req adminUserUpsertRequest
	if ok := decodeUsersJSONBody(w, r, &req); !ok {
		return
	}

	item, err := h.service.AdminUpdate(r.Context(), r.PathValue("id"), usersservice.AdminUpdateUserInput{Name: req.Name, Email: req.Email})
	if err != nil {
		switch {
		case errors.Is(err, usersservice.ErrInvalidAdminUserInput):
			writeUsersJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, usersservice.ErrAdminUserAlreadyExists):
			writeUsersJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		case errors.Is(err, usersservice.ErrAdminUserNotFound):
			writeUsersJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user"})
		}
		return
	}

	writeUsersJSON(w, http.StatusOK, item)
}

func (h *UsersHandler) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	err := h.service.AdminDelete(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, usersservice.ErrAdminUserNotFound) {
			writeUsersJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete user"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.writeAuditTrail(r, "admin_user_deleted", "user", r.PathValue("id"), nil)
}

func (h *UsersHandler) adminBlockUser(w http.ResponseWriter, r *http.Request) {
	h.adminSetUserBlockState(w, r, true)
}

func (h *UsersHandler) adminUnblockUser(w http.ResponseWriter, r *http.Request) {
	h.adminSetUserBlockState(w, r, false)
}

func (h *UsersHandler) adminSetUserBlockState(w http.ResponseWriter, r *http.Request, blocked bool) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var (
		item *usersservice.AdminUser
		err  error
	)
	if blocked {
		item, err = h.service.AdminBlock(r.Context(), r.PathValue("id"))
	} else {
		item, err = h.service.AdminUnblock(r.Context(), r.PathValue("id"))
	}
	if err != nil {
		if errors.Is(err, usersservice.ErrAdminUserNotFound) {
			writeUsersJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user status"})
		return
	}

	writeUsersJSON(w, http.StatusOK, item)
	action := "admin_user_unblocked"
	if blocked {
		action = "admin_user_blocked"
	}
	h.writeAuditTrail(r, action, "user", item.ID, map[string]string{"status": item.Status})
}

func (h *UsersHandler) adminStats(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	stats, err := h.service.AdminStats(r.Context())
	if err != nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load stats"})
		return
	}

	writeUsersJSON(w, http.StatusOK, stats)
}

func (h *UsersHandler) adminExport(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	result, err := h.service.AdminExport(r.Context(), r.URL.Query().Get("format"))
	if err != nil {
		if errors.Is(err, usersservice.ErrUnsupportedExportFormat) {
			writeUsersJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to export backup"})
		return
	}

	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+result.Filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.Data)
	h.writeAuditTrail(r, "backup_export_sync", "backup", result.Format, map[string]string{"filename": result.Filename})
}

func (h *UsersHandler) adminStartExportJob(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	job, err := h.service.StartAdminExportJob(r.Context(), r.URL.Query().Get("format"))
	if err != nil {
		if errors.Is(err, usersservice.ErrUnsupportedExportFormat) {
			writeUsersJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start export job"})
		return
	}

	writeUsersJSON(w, http.StatusAccepted, job)
	h.writeAuditTrail(r, "backup_export_job_started", "backup_job", job.ID, map[string]string{"format": job.Format})
}

func (h *UsersHandler) adminListExportJobs(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	page := parseUsersIntOr(r.URL.Query().Get("page"), 1)
	limit := parseUsersIntOr(r.URL.Query().Get("limit"), 20)
	items, total := h.service.ListAdminExportJobs(r.Context(), page, limit)
	writeUsersJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"meta": map[string]int{"page": normalizeUsersPage(page), "limit": normalizeUsersLimit(limit), "total": total},
	})
}

func (h *UsersHandler) adminGetExportJob(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	job, err := h.service.GetAdminExportJob(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, usersservice.ErrExportJobNotFound) {
			writeUsersJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get export job"})
		return
	}

	writeUsersJSON(w, http.StatusOK, job)
}

func (h *UsersHandler) adminDownloadExportJob(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	result, err := h.service.ReadAdminExportJobFile(r.Context(), r.PathValue("id"))
	if err != nil {
		switch {
		case errors.Is(err, usersservice.ErrExportJobNotFound):
			writeUsersJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, usersservice.ErrExportJobNotReady):
			writeUsersJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read export file"})
		}
		return
	}

	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+result.Filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.Data)
	h.writeAuditTrail(r, "backup_export_job_downloaded", "backup", r.PathValue("id"), map[string]string{"filename": result.Filename})
}

type adminAuditLogRequest struct {
	Action     string            `json:"action"`
	TargetType string            `json:"target_type"`
	TargetID   string            `json:"target_id"`
	Metadata   map[string]string `json:"metadata"`
}

func (h *UsersHandler) adminWriteAuditLog(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	actorID, _, ok := usersIdentityFromHeaders(r)
	if !ok {
		writeUsersJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing forwarded user identity"})
		return
	}

	var req adminAuditLogRequest
	if ok := decodeUsersJSONBody(w, r, &req); !ok {
		return
	}

	entry, err := h.service.CreateAuditLog(r.Context(), usersservice.AuditLogCreateInput{
		ActorID:    actorID,
		Action:     req.Action,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Metadata:   req.Metadata,
	})
	if err != nil {
		if errors.Is(err, usersservice.ErrInvalidAuditLogInput) {
			writeUsersJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to write audit log"})
		return
	}

	writeUsersJSON(w, http.StatusCreated, entry)
}

func (h *UsersHandler) adminListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	q := r.URL.Query()
	page := parseUsersIntOr(q.Get("page"), 1)
	limit := parseUsersIntOr(q.Get("limit"), 20)
	items, total, err := h.service.ListAuditLogs(r.Context(), usersservice.AuditLogListFilter{
		ActorID: q.Get("actor_id"),
		Action:  q.Get("action"),
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		writeUsersJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list audit logs"})
		return
	}

	writeUsersJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"meta": map[string]int{"page": normalizeUsersPage(page), "limit": normalizeUsersLimit(limit), "total": total},
	})
}

func (h *UsersHandler) writeAuditTrail(r *http.Request, action, targetType, targetID string, metadata map[string]string) {
	if h.service == nil || r == nil {
		return
	}
	actorID, _, ok := usersIdentityFromHeaders(r)
	if !ok {
		return
	}
	_, _ = h.service.CreateAuditLog(r.Context(), usersservice.AuditLogCreateInput{
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
	})
}

func parseUsersIntOr(s string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return v
}

func normalizeUsersPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeUsersLimit(limit int) int {
	if limit < 1 || limit > 100 {
		return 20
	}
	return limit
}

func decodeUsersJSONBody(w http.ResponseWriter, r *http.Request, out any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		writeUsersJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeUsersJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	return true
}
