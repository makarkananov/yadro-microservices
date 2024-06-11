package auth

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	authv1 "yadro-microservices/api/gen/go/auth"
	"yadro-microservices/internal/core/domain"
)

// Client represents a gRPC client for auth operations.
type Client struct {
	conn   *grpc.ClientConn
	client authv1.AuthClient
}

// NewClient creates a new Client instance.
// It establishes a connection to the gRPC server at the specified address.
func NewClient(serverAddress string) (*Client, error) {
	conn, err := grpc.Dial(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	authClient := authv1.NewAuthClient(conn)

	return &Client{
		conn:   conn,
		client: authClient,
	}, nil
}

// Close closes the connection to the gRPC server.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Login logs in the user with the specified username and password.
func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
	resp, err := c.client.Login(
		ctx,
		&authv1.LoginRequest{
			Username: username,
			Password: password,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}

	return resp.GetToken(), nil
}

// Register registers a new user with the specified username, password, and role.
func (c *Client) Register(ctx context.Context, username, password string, role domain.Role) error {
	// Default role is USER
	if role == domain.UNDEFINED {
		role = domain.USER
	}

	_, err := c.client.Register(
		ctx,
		&authv1.RegisterRequest{
			Username: username,
			Password: password,
			Role:     string(role),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	return nil
}

// ValidateToken validates the specified token.
func (c *Client) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	resp, err := c.client.ValidateToken(
		ctx,
		&authv1.ValidateTokenRequest{
			Token: token,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	return &domain.User{
		Username: resp.GetUsername(),
		Password: resp.GetPassword(),
		Role:     domain.Role(resp.GetRole()),
	}, nil
}
