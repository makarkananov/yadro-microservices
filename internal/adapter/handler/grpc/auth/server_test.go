package auth_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"log"
	"net"
	"testing"
	"yadro-microservices/internal/adapter/handler/grpc/auth"
	"yadro-microservices/internal/core/port"
	"yadro-microservices/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	authv1 "yadro-microservices/api/gen/go/auth"
	"yadro-microservices/internal/core/domain"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func startTestServer(authService port.AuthService) (*grpc.ClientConn, authv1.AuthClient) {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	server := auth.NewServer(authService)
	authv1.RegisterAuthServer(s, server)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	client := authv1.NewAuthClient(conn)

	return conn, client
}

func TestServer_Login(t *testing.T) {
	mockAuthService := new(mocks.AuthService)
	mockAuthService.On(
		"Login",
		mock.Anything,
		"testuser",
		"testpassword",
	).Return("token123", nil)
	mockAuthService.On(
		"Login",
		mock.Anything,
		"wronguser",
		"wrongpassword",
	).Return("", errors.New("invalid credentials"))

	conn, client := startTestServer(mockAuthService)
	defer conn.Close()

	t.Run("successful login", func(t *testing.T) {
		resp, err := client.Login(context.Background(), &authv1.LoginRequest{
			Username: "testuser",
			Password: "testpassword",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "token123", resp.GetToken())
	})

	t.Run("failed login", func(t *testing.T) {
		resp, err := client.Login(context.Background(), &authv1.LoginRequest{
			Username: "wronguser",
			Password: "wrongpassword",
		})

		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_Register(t *testing.T) {
	mockAuthService := new(mocks.AuthService)
	mockAuthService.On(
		"Register",
		mock.Anything,
		mock.AnythingOfType("*domain.User"),
	).Return(nil).Once()
	mockAuthService.On(
		"Register",
		mock.Anything,
		mock.AnythingOfType("*domain.User"),
	).Return(errors.New("registration error")).Once()

	conn, client := startTestServer(mockAuthService)
	defer conn.Close()

	t.Run("successful registration", func(t *testing.T) {
		resp, err := client.Register(context.Background(), &authv1.RegisterRequest{
			Username: "newuser",
			Password: "newpassword",
			Role:     "user",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("failed registration", func(t *testing.T) {
		resp, err := client.Register(context.Background(), &authv1.RegisterRequest{
			Username: "existinguser",
			Password: "password",
			Role:     "user",
		})

		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_ValidateToken(t *testing.T) {
	mockAuthService := new(mocks.AuthService)
	mockAuthService.On("ValidateToken", mock.Anything, "validtoken").Return(&domain.User{
		Username: "validuser",
		Password: "validpassword",
		Role:     domain.USER,
	}, nil)
	mockAuthService.On(
		"ValidateToken",
		mock.Anything,
		"invalidtoken",
	).Return(nil, errors.New("invalid token"))

	conn, client := startTestServer(mockAuthService)
	defer conn.Close()

	t.Run("successful token validation", func(t *testing.T) {
		resp, err := client.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{
			Token: "validtoken",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "validuser", resp.GetUsername())
		assert.Equal(t, "validpassword", resp.GetPassword())
		assert.Equal(t, "user", resp.GetRole())
	})

	t.Run("failed token validation", func(t *testing.T) {
		resp, err := client.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{
			Token: "invalidtoken",
		})

		require.Error(t, err)
		assert.Nil(t, resp)
	})
}
