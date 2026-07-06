package usersservice

import "errors"

var ErrInvalidIdentity = errors.New("invalid user identity")

var (
	ErrInvalidAdminUserInput = errors.New("invalid admin user input")
	ErrAdminUserNotFound     = errors.New("admin user not found")
	ErrAdminUserAlreadyExists = errors.New("admin user already exists")
	ErrUnsupportedExportFormat = errors.New("unsupported export format")
	ErrExportJobNotFound      = errors.New("export job not found")
	ErrExportJobNotReady      = errors.New("export job not ready")
	ErrInvalidAuditLogInput   = errors.New("invalid audit log input")
)
