package main

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"
	authv1 "vocabulary/backend/proto/auth/v1"
	authservice "vocabulary/backend/modules/auth/service"
)

type authGRPCHandler struct {
	svc *authservice.AuthService
}

type authGRPCService interface {
	Login(context.Context, *authv1.LoginRequest) (*authv1.LoginResponse, error)
	CreateAdmin(context.Context, *authv1.CreateAdminRequest) (*authv1.Admin, error)
	Health(context.Context, *authv1.HealthRequest) (*authv1.HealthResponse, error)
}

func (h *authGRPCHandler) Login(_ context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	resp, err := h.svc.Login(authservice.AuthLoginRequest{Email: strings.TrimSpace(req.Email), Password: req.Password})
	if err != nil {
		switch {
		case err == authservice.ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case err == authservice.ErrBootstrapAdminNotConfigured:
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to process login")
		}
	}
	return &authv1.LoginResponse{AccessToken: resp.AccessToken, TokenType: resp.TokenType, ExpiresIn: int32(resp.ExpiresIn)}, nil
}

func (h *authGRPCHandler) CreateAdmin(ctx context.Context, req *authv1.CreateAdminRequest) (*authv1.Admin, error) {
	admin, err := h.svc.CreateAdmin(ctx, authservice.AuthCreateAdminRequest{Email: req.Email, Password: req.Password, Role: req.Role})
	if err != nil {
		switch {
		case err == authservice.ErrAuthInvalidEmail, err == authservice.ErrAuthInvalidPassword, err == authservice.ErrAuthInvalidRole:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case err == authservice.ErrAuthAdminAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case err == authservice.ErrAuthRepoNotConfigured:
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create admin")
		}
	}
	return &authv1.Admin{ID: admin.ID, Email: admin.Email, Role: admin.Role, CreatedAt: admin.CreatedAt.UTC().Format(time.RFC3339)}, nil
}

func (*authGRPCHandler) Health(context.Context, *authv1.HealthRequest) (*authv1.HealthResponse, error) {
	return &authv1.HealthResponse{Status: "ok", Service: "auth-service"}, nil
}

func startAuthGRPCServer(port int, svc *authservice.AuthService) error {
	encoding.RegisterCodec(grpcJSONCodec{})

	lis, err := net.Listen("tcp", net.JoinHostPort("", strconvAuthPort(port)))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: authv1.ServiceName,
		HandlerType: (*authGRPCService)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "Login", Handler: authGRPCLoginHandler},
			{MethodName: "CreateAdmin", Handler: authGRPCCreateAdminHandler},
			{MethodName: "Health", Handler: authGRPCHealthHandler},
		},
	}, &authGRPCHandler{svc: svc})

	return server.Serve(lis)
}

func authGRPCLoginHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	req := &authv1.LoginRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	h := srv.(*authGRPCHandler)
	return h.Login(ctx, req)
}

func authGRPCCreateAdminHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	req := &authv1.CreateAdminRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	h := srv.(*authGRPCHandler)
	return h.CreateAdmin(ctx, req)
}

func authGRPCHealthHandler(_ any, _ context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	req := &authv1.HealthRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	h := srv.(*authGRPCHandler)
	return h.Health(ctx, req)
}

func strconvAuthPort(port int) string {
	if port <= 0 {
		return "9091"
	}
	return strconv.Itoa(port)
}
