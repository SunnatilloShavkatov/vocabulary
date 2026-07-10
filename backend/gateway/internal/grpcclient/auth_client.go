package grpcclient

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	authv1 "vocabulary/backend/proto/auth/v1"
)

type AuthClient struct {
	target  string
	timeout time.Duration
	conn    *grpc.ClientConn
}

type AuthServiceClient interface {
	Target() string
	CheckConnection(context.Context) error
	Login(context.Context, authv1.LoginRequest) (authv1.LoginResponse, error)
	CreateAdmin(context.Context, authv1.CreateAdminRequest) (authv1.Admin, error)
	Health(context.Context) (authv1.HealthResponse, error)
	Close() error
}

func init() {
	encoding.RegisterCodec(grpcJSONCodec{})
}

func NewAuthClient(target string) *AuthClient {
	if target == "" {
		target = "localhost:9091"
	}

	// grpc.NewClient is the modern replacement for grpc.Dial in gRPC v1.62+.
	// It performs lazy connection initialization in the background, which is extremely resilient.
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error: failed to create gRPC client for %s: %v", target, err)
	}

	return &AuthClient{
		target:  target,
		timeout: 2 * time.Second,
		conn:    conn,
	}
}

func (c *AuthClient) Target() string {
	return c.target
}

func (c *AuthClient) Close() error {
	if c != nil && c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *AuthClient) CheckConnection(ctx context.Context) error {
	_, err := c.Health(ctx)
	return err
}

func (c *AuthClient) Login(ctx context.Context, req authv1.LoginRequest) (authv1.LoginResponse, error) {
	if c == nil {
		return authv1.LoginResponse{}, errors.New("auth grpc client is not initialized")
	}
	if c.conn == nil {
		return authv1.LoginResponse{}, errors.New("auth grpc connection is not initialized")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp := authv1.LoginResponse{}
	err := c.conn.Invoke(timeoutCtx, authv1.MethodLogin, &req, &resp, grpc.ForceCodec(grpcJSONCodec{}))
	if err != nil {
		return authv1.LoginResponse{}, err
	}
	return resp, nil
}

func (c *AuthClient) CreateAdmin(ctx context.Context, req authv1.CreateAdminRequest) (authv1.Admin, error) {
	if c == nil {
		return authv1.Admin{}, errors.New("auth grpc client is not initialized")
	}
	if c.conn == nil {
		return authv1.Admin{}, errors.New("auth grpc connection is not initialized")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp := authv1.Admin{}
	err := c.conn.Invoke(timeoutCtx, authv1.MethodCreateAdmin, &req, &resp, grpc.ForceCodec(grpcJSONCodec{}))
	if err != nil {
		return authv1.Admin{}, err
	}
	return resp, nil
}

func (c *AuthClient) Health(ctx context.Context) (authv1.HealthResponse, error) {
	if c == nil {
		return authv1.HealthResponse{}, errors.New("auth grpc client is not initialized")
	}
	if c.conn == nil {
		return authv1.HealthResponse{}, errors.New("auth grpc connection is not initialized")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp := authv1.HealthResponse{}
	err := c.conn.Invoke(timeoutCtx, authv1.MethodHealth, &authv1.HealthRequest{}, &resp, grpc.ForceCodec(grpcJSONCodec{}))
	if err != nil {
		return authv1.HealthResponse{}, err
	}
	return resp, nil
}

