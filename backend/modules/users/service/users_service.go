package usersservice

import (
	"encoding/json"
	"context"
	"os"
	"path/filepath"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type UserProfile struct {
	ID       string            `json:"id"`
	Role     string            `json:"role"`
	Settings map[string]string `json:"settings"`
}

type AdminUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AdminCreateUserInput struct {
	Name  string
	Email string
}

type AdminUpdateUserInput struct {
	Name  string
	Email string
}

type AdminStats struct {
	UsersCount int `json:"users_count"`
	WordsCount int `json:"words_count"`
}

type AdminExportResult struct {
	Format      string `json:"format"`
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
	Data        []byte `json:"-"`
}

type BackupExportJob struct {
	ID          string    `json:"id"`
	Format      string    `json:"format"`
	Status      string    `json:"status"`
	Filename    string    `json:"filename,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
	DownloadURL string    `json:"download_url,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID         string            `json:"id"`
	ActorID    string            `json:"actor_id"`
	Action     string            `json:"action"`
	TargetType string            `json:"target_type"`
	TargetID   string            `json:"target_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

type AuditLogCreateInput struct {
	ActorID    string
	Action     string
	TargetType string
	TargetID   string
	Metadata   map[string]string
}

type AuditLogListFilter struct {
	ActorID string
	Action  string
	Page    int
	Limit   int
}

type UsersRepository interface {
	FindProfile(ctx context.Context, userID string) (*UserProfile, error)
	AdminList(ctx context.Context, query string) ([]AdminUser, error)
	AdminCreate(ctx context.Context, input AdminCreateUserInput) (*AdminUser, error)
	AdminGet(ctx context.Context, id string) (*AdminUser, error)
	AdminUpdate(ctx context.Context, id string, input AdminUpdateUserInput) (*AdminUser, error)
	AdminDelete(ctx context.Context, id string) error
	AdminSetStatus(ctx context.Context, id string, status string) (*AdminUser, error)
	AdminStats(ctx context.Context) (*AdminStats, error)
	AdminExport(ctx context.Context, format string) (*AdminExportResult, error)
	CreateAuditLog(ctx context.Context, input AuditLogCreateInput) (*AuditLog, error)
	ListAuditLogs(ctx context.Context, filter AuditLogListFilter) ([]AuditLog, int, error)
}

type UsersService struct {
	repo UsersRepository

	mu        sync.RWMutex
	users     map[string]AdminUser
	nextUser  int
	auditLogs []AuditLog

	jobsMu      sync.RWMutex
	exportJobs  map[string]BackupExportJob
	exportFiles map[string]string
	exportDir   string
}

func NewUsersService(repo UsersRepository) *UsersService {
	exportDir := strings.TrimSpace(os.Getenv("BACKUP_EXPORT_DIR"))
	if exportDir == "" {
		exportDir = filepath.Join("storage", "exports")
	}
	return &UsersService{
		repo:        repo,
		users:       map[string]AdminUser{},
		nextUser:    1,
		auditLogs:   []AuditLog{},
		exportJobs:  map[string]BackupExportJob{},
		exportFiles: map[string]string{},
		exportDir:   exportDir,
	}
}

func (s *UsersService) Me(ctx context.Context, userID, role string) (*UserProfile, error) {
	userID = strings.TrimSpace(userID)
	role = strings.TrimSpace(role)
	if userID == "" || role == "" {
		return nil, ErrInvalidIdentity
	}

	if s.repo != nil {
		profile, err := s.repo.FindProfile(ctx, userID)
		if err != nil {
			return nil, err
		}
		if profile != nil {
			if strings.TrimSpace(profile.Role) == "" {
				profile.Role = role
			}
			return profile, nil
		}
	}

	return &UserProfile{
		ID:       userID,
		Role:     role,
		Settings: map[string]string{},
	}, nil
}

func (s *UsersService) AdminList(ctx context.Context, query string) ([]AdminUser, error) {
	if s.repo != nil {
		return s.repo.AdminList(ctx, query)
	}
	_ = ctx
	needle := strings.ToLower(strings.TrimSpace(query))

	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]AdminUser, 0, len(s.users))
	for _, u := range s.users {
		if needle == "" || strings.Contains(strings.ToLower(u.Name), needle) || strings.Contains(strings.ToLower(u.Email), needle) {
			items = append(items, u)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items, nil
}

func (s *UsersService) AdminCreate(ctx context.Context, input AdminCreateUserInput) (*AdminUser, error) {
	if s.repo != nil {
		return s.repo.AdminCreate(ctx, input)
	}
	_ = ctx
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if name == "" || !isValidEmail(email) {
		return nil, ErrInvalidAdminUserInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if strings.EqualFold(u.Email, email) {
			return nil, ErrAdminUserAlreadyExists
		}
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("user-%d", s.nextUser)
	s.nextUser++

	u := AdminUser{
		ID:        id,
		Name:      name,
		Email:     email,
		Role:      "client",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.users[id] = u

	return &u, nil
}

func (s *UsersService) AdminGet(ctx context.Context, id string) (*AdminUser, error) {
	if s.repo != nil {
		return s.repo.AdminGet(ctx, id)
	}
	_ = ctx
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrAdminUserNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return nil, ErrAdminUserNotFound
	}

	return &u, nil
}

func (s *UsersService) AdminUpdate(ctx context.Context, id string, input AdminUpdateUserInput) (*AdminUser, error) {
	if s.repo != nil {
		return s.repo.AdminUpdate(ctx, id, input)
	}
	_ = ctx
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrAdminUserNotFound
	}

	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if name == "" || !isValidEmail(email) {
		return nil, ErrInvalidAdminUserInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[id]
	if !ok {
		return nil, ErrAdminUserNotFound
	}

	for k, v := range s.users {
		if k != id && strings.EqualFold(v.Email, email) {
			return nil, ErrAdminUserAlreadyExists
		}
	}

	u.Name = name
	u.Email = email
	u.UpdatedAt = time.Now().UTC()
	s.users[id] = u

	return &u, nil
}

func (s *UsersService) AdminDelete(ctx context.Context, id string) error {
	if s.repo != nil {
		return s.repo.AdminDelete(ctx, id)
	}
	_ = ctx
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrAdminUserNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; !ok {
		return ErrAdminUserNotFound
	}

	delete(s.users, id)
	return nil
}

func (s *UsersService) AdminBlock(ctx context.Context, id string) (*AdminUser, error) {
	return s.adminSetStatus(ctx, id, "blocked")
}

func (s *UsersService) AdminUnblock(ctx context.Context, id string) (*AdminUser, error) {
	return s.adminSetStatus(ctx, id, "active")
}

func (s *UsersService) AdminStats(ctx context.Context) (*AdminStats, error) {
	if s.repo != nil {
		return s.repo.AdminStats(ctx)
	}
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &AdminStats{UsersCount: len(s.users), WordsCount: 0}, nil
}

func (s *UsersService) AdminExport(ctx context.Context, format string) (*AdminExportResult, error) {
	if s.repo != nil {
		return s.repo.AdminExport(ctx, format)
	}
	_ = ctx
	f := strings.TrimSpace(strings.ToLower(format))
	if f == "" {
		f = "json"
	}

	s.mu.RLock()
	users := make([]AdminUser, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	s.mu.RUnlock()

	sort.Slice(users, func(i, j int) bool {
		return users[i].CreatedAt.Before(users[j].CreatedAt)
	})

	switch f {
	case "json":
		payload, err := json.MarshalIndent(map[string]any{"users": users}, "", "  ")
		if err != nil {
			return nil, err
		}
		return &AdminExportResult{Format: "json", ContentType: "application/json", Filename: "admin_export.json", Data: payload}, nil
	case "sql":
		var b strings.Builder
		b.WriteString("-- admin export users\n")
		for _, u := range users {
			b.WriteString(fmt.Sprintf("INSERT INTO users (id, name, email, role, status) VALUES ('%s','%s','%s','%s','%s');\n",
				escapeSQL(u.ID), escapeSQL(u.Name), escapeSQL(u.Email), escapeSQL(u.Role), escapeSQL(u.Status)))
		}
		return &AdminExportResult{Format: "sql", ContentType: "application/sql", Filename: "admin_export.sql", Data: []byte(b.String())}, nil
	default:
		return nil, ErrUnsupportedExportFormat
	}
}

func (s *UsersService) StartAdminExportJob(ctx context.Context, format string) (*BackupExportJob, error) {
	f := strings.TrimSpace(strings.ToLower(format))
	if f == "" {
		f = "json"
	}
	if f != "json" && f != "sql" {
		return nil, ErrUnsupportedExportFormat
	}

	now := time.Now().UTC()
	job := BackupExportJob{
		ID:        newUsersID("export-job"),
		Format:    f,
		Status:    "running",
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.jobsMu.Lock()
	s.exportJobs[job.ID] = job
	s.jobsMu.Unlock()

	go s.runExportJob(context.Background(), job.ID, f)

	copyJob := job
	return &copyJob, nil
}

func (s *UsersService) ListAdminExportJobs(_ context.Context, page, limit int) ([]BackupExportJob, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	s.jobsMu.RLock()
	items := make([]BackupExportJob, 0, len(s.exportJobs))
	for _, job := range s.exportJobs {
		items = append(items, job)
	}
	s.jobsMu.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	total := len(items)
	start := (page - 1) * limit
	if start >= total {
		return []BackupExportJob{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total
}

func (s *UsersService) GetAdminExportJob(_ context.Context, jobID string) (*BackupExportJob, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, ErrExportJobNotFound
	}

	s.jobsMu.RLock()
	job, ok := s.exportJobs[jobID]
	s.jobsMu.RUnlock()
	if !ok {
		return nil, ErrExportJobNotFound
	}
	copyJob := job
	return &copyJob, nil
}

func (s *UsersService) ReadAdminExportJobFile(_ context.Context, jobID string) (*AdminExportResult, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, ErrExportJobNotFound
	}

	s.jobsMu.RLock()
	job, ok := s.exportJobs[jobID]
	filePath := s.exportFiles[jobID]
	s.jobsMu.RUnlock()
	if !ok || job.Status != "completed" || strings.TrimSpace(filePath) == "" {
		return nil, ErrExportJobNotReady
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &AdminExportResult{Format: job.Format, ContentType: job.ContentType, Filename: job.Filename, Data: data}, nil
}

func (s *UsersService) CreateAuditLog(ctx context.Context, input AuditLogCreateInput) (*AuditLog, error) {
	input.ActorID = strings.TrimSpace(input.ActorID)
	input.Action = strings.TrimSpace(strings.ToLower(input.Action))
	input.TargetType = strings.TrimSpace(strings.ToLower(input.TargetType))
	input.TargetID = strings.TrimSpace(input.TargetID)
	if input.ActorID == "" || input.Action == "" || input.TargetType == "" {
		return nil, ErrInvalidAuditLogInput
	}

	if s.repo != nil {
		return s.repo.CreateAuditLog(ctx, input)
	}

	entry := AuditLog{
		ID:         newUsersID("audit"),
		ActorID:    input.ActorID,
		Action:     input.Action,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		Metadata:   cloneStringMap(input.Metadata),
		CreatedAt:  time.Now().UTC(),
	}

	s.mu.Lock()
	s.auditLogs = append(s.auditLogs, entry)
	s.mu.Unlock()

	copyEntry := entry
	return &copyEntry, nil
}

func (s *UsersService) ListAuditLogs(ctx context.Context, filter AuditLogListFilter) ([]AuditLog, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	filter.ActorID = strings.TrimSpace(filter.ActorID)
	filter.Action = strings.TrimSpace(strings.ToLower(filter.Action))

	if s.repo != nil {
		return s.repo.ListAuditLogs(ctx, filter)
	}

	s.mu.RLock()
	items := make([]AuditLog, 0, len(s.auditLogs))
	for _, entry := range s.auditLogs {
		if filter.ActorID != "" && entry.ActorID != filter.ActorID {
			continue
		}
		if filter.Action != "" && entry.Action != filter.Action {
			continue
		}
		items = append(items, entry)
	}
	s.mu.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	total := len(items)
	start := (filter.Page - 1) * filter.Limit
	if start >= total {
		return []AuditLog{}, total, nil
	}
	end := start + filter.Limit
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

func (s *UsersService) adminSetStatus(ctx context.Context, id string, status string) (*AdminUser, error) {
	if s.repo != nil {
		return s.repo.AdminSetStatus(ctx, id, status)
	}
	_ = ctx
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrAdminUserNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[id]
	if !ok {
		return nil, ErrAdminUserNotFound
	}

	u.Status = status
	u.UpdatedAt = time.Now().UTC()
	s.users[id] = u

	return &u, nil
}

func escapeSQL(v string) string {
	return strings.ReplaceAll(v, "'", "''")
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	at := strings.Index(email, "@")
	return at > 0 && at < len(email)-1
}

func (s *UsersService) runExportJob(ctx context.Context, jobID, format string) {
	result, err := s.AdminExport(ctx, format)
	if err != nil {
		s.setExportJobFailed(jobID, err.Error())
		return
	}

	if err := os.MkdirAll(s.exportDir, 0o755); err != nil {
		s.setExportJobFailed(jobID, err.Error())
		return
	}

	baseFilename := result.Filename
	if strings.TrimSpace(baseFilename) == "" {
		baseFilename = "admin_export." + format
	}
	filename := fmt.Sprintf("%s_%s", jobID, baseFilename)
	filePath := filepath.Join(s.exportDir, filename)
	if err := os.WriteFile(filePath, result.Data, 0o644); err != nil {
		s.setExportJobFailed(jobID, err.Error())
		return
	}

	now := time.Now().UTC()
	s.jobsMu.Lock()
	job := s.exportJobs[jobID]
	job.Status = "completed"
	job.Filename = baseFilename
	job.ContentType = result.ContentType
	job.DownloadURL = "/v1/admin/backup/export-jobs/" + jobID + "/download"
	job.UpdatedAt = now
	s.exportJobs[jobID] = job
	s.exportFiles[jobID] = filePath
	s.jobsMu.Unlock()
}

func (s *UsersService) setExportJobFailed(jobID, errText string) {
	s.jobsMu.Lock()
	job := s.exportJobs[jobID]
	job.Status = "failed"
	job.Error = strings.TrimSpace(errText)
	job.UpdatedAt = time.Now().UTC()
	s.exportJobs[jobID] = job
	s.jobsMu.Unlock()
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func newUsersID(prefix string) string {
	return fmt.Sprintf("%s-%d", strings.TrimSpace(prefix), time.Now().UTC().UnixNano())
}
