package grpcclient

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	authv1 "vocabulary/backend/proto/auth/v1"
)

type AuthClient struct {
	target string
	timeout time.Duration
}

type AuthServiceClient interface {
	Target() string
	CheckConnection(context.Context) error
	Login(context.Context, authv1.LoginRequest) (authv1.LoginResponse, error)
	CreateAdmin(context.Context, authv1.CreateAdminRequest) (authv1.Admin, error)
	Health(context.Context) (authv1.HealthResponse, error)
}

func init() {
	encoding.RegisterCodec(grpcJSONCodec{})
}

func NewAuthClient(target string) *AuthClient {
	if target == "" {
		target = "localhost:9091"
	}
	return &AuthClient{target: target, timeout: 2 * time.Second}
}

func (c *AuthClient) Target() string {
	return c.target
}

func (c *AuthClient) CheckConnection(ctx context.Context) error {
	_, err := c.Health(ctx)
	return err
}

func (c *AuthClient) Login(ctx context.Context, req authv1.LoginRequest) (authv1.LoginResponse, error) {
	if c == nil {
		return authv1.LoginResponse{}, errors.New("auth grpc client is not initialized")
	}

	conn, timeoutCtx, cancel, err := c.openConn(ctx)
	if err != nil {
		return authv1.LoginResponse{}, err
	}
	defer cancel()
	defer conn.Close()

	resp := authv1.LoginResponse{}
	err = conn.Invoke(timeoutCtx, authv1.MethodLogin, &req, &resp, grpc.ForceCodec(grpcJSONCodec{}))
	if err != nil {
		return authv1.LoginResponse{}, err
	}
	return resp, nil
}

func (c *AuthClient) CreateAdmin(ctx context.Context, req authv1.CreateAdminRequest) (authv1.Admin, error) {
	if c == nil {
		return authv1.Admin{}, errors.New("auth grpc client is not initialized")
	}

	conn, timeoutCtx, cancel, err := c.openConn(ctx)
	if err != nil {
		return authv1.Admin{}, err
	}
	defer cancel()
	defer conn.Close()

	resp := authv1.Admin{}
	err = conn.Invoke(timeoutCtx, authv1.MethodCreateAdmin, &req, &resp, grpc.ForceCodec(grpcJSONCodec{}))
	if err != nil {
		return authv1.Admin{}, err
	}
	return resp, nil
}

func (c *AuthClient) Health(ctx context.Context) (authv1.HealthResponse, error) {
	if c == nil {
		return authv1.HealthResponse{}, errors.New("auth grpc client is not initialized")
	}

	conn, timeoutCtx, cancel, err := c.openConn(ctx)
	if err != nil {
		return authv1.HealthResponse{}, err
	}
	defer cancel()
	defer conn.Close()

	resp := authv1.HealthResponse{}
	err = conn.Invoke(timeoutCtx, authv1.MethodHealth, &authv1.HealthRequest{}, &resp, grpc.ForceCodec(grpcJSONCodec{}))
	if err != nil {
		return authv1.HealthResponse{}, err
	}
	return resp, nil
}


func (c *AuthClient) openConn(ctx context.Context) (*grpc.ClientConn, context.Context, context.CancelFunc, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	conn, err := grpc.DialContext(timeoutCtx, c.target, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}
	return conn, timeoutCtx, cancel, nil
}
