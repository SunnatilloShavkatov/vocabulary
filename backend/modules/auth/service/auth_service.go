package authservice

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"vocabulary/backend/libs/shared/config"
)

type AuthService struct {
	cfg config.Config
	repo AuthRepository
}

type AuthLoginRequest struct {
	Email    string
	Password string
}

type AuthLoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type AuthAdmin struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthCreateAdminRequest struct {
	Email    string
	Password string
	Role     string
}

type AuthRepository interface {
	CreateAdmin(ctx context.Context, email, passwordHash, role string) (*AuthAdmin, error)
}

var (
	ErrInvalidCredentials          = errors.New("invalid credentials")
	ErrBootstrapAdminNotConfigured = errors.New("bootstrap admin is not configured")
	ErrAuthAdminAlreadyExists      = errors.New("admin with this email already exists")
	ErrAuthRepoNotConfigured       = errors.New("auth repository is not configured")
	ErrAuthInvalidEmail            = errors.New("email is invalid")
	ErrAuthInvalidPassword         = errors.New("password must be between 8 and 72 characters")
	ErrAuthInvalidRole             = errors.New("unsupported role")
)

func NewAuthService(cfg config.Config, repos ...AuthRepository) *AuthService {
	var repo AuthRepository
	if len(repos) > 0 {
		repo = repos[0]
	}
	return &AuthService{cfg: cfg, repo: repo}
}

func (s *AuthService) Login(req AuthLoginRequest) (AuthLoginResponse, error) {
	if strings.TrimSpace(s.cfg.BootstrapAdmin.Email) == "" || s.cfg.BootstrapAdmin.Password == "" {
		return AuthLoginResponse{}, ErrBootstrapAdminNotConfigured
	}

	if !strings.EqualFold(strings.TrimSpace(req.Email), s.cfg.BootstrapAdmin.Email) ||
		subtle.ConstantTimeCompare([]byte(req.Password), []byte(s.cfg.BootstrapAdmin.Password)) != 1 {
		return AuthLoginResponse{}, ErrInvalidCredentials
	}

	expiresIn := s.accessTokenTTLSeconds()
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":   "bootstrap-admin",
		"email": s.cfg.BootstrapAdmin.Email,
		"role":  "admin",
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(expiresIn) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret()))
	if err != nil {
		return AuthLoginResponse{}, err
	}

	return AuthLoginResponse{AccessToken: signedToken, TokenType: "Bearer", ExpiresIn: expiresIn}, nil
}

func (s *AuthService) accessTokenTTLSeconds() int {
	minutes := s.cfg.JWT.AccessTTLMinutes
	if minutes <= 0 {
		minutes = 15
	}
	return minutes * 60
}

func (s *AuthService) jwtSecret() string {
	if s.cfg.JWT.Secret == "" {
		return "change-me"
	}
	return s.cfg.JWT.Secret
}

func (s *AuthService) CreateAdmin(ctx context.Context, req AuthCreateAdminRequest) (*AuthAdmin, error) {
	if s.repo == nil {
		return nil, ErrAuthRepoNotConfigured
	}

	email := strings.TrimSpace(req.Email)
	if !isValidEmail(email) {
		return nil, ErrAuthInvalidEmail
	}

	if len(req.Password) < 8 || len(req.Password) > 72 {
		return nil, ErrAuthInvalidPassword
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "admin"
	}
	if role != "admin" {
		return nil, ErrAuthInvalidRole
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.repo.CreateAdmin(ctx, email, string(passwordHash), role)
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	at := strings.Index(email, "@")
	if at <= 0 || at >= len(email)-1 {
		return false
	}
	return true
}

