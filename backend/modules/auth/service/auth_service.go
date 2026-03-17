package authservice

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"vocabulary/backend/libs/shared/config"
)

type AuthService struct {
	cfg config.Config
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

var (
	ErrInvalidCredentials          = errors.New("invalid credentials")
	ErrBootstrapAdminNotConfigured = errors.New("bootstrap admin is not configured")
)

func NewAuthService(cfg config.Config) *AuthService {
	return &AuthService{cfg: cfg}
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

