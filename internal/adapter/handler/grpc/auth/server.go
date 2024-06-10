package auth

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	authv1 "yadro-microservices/api/gen/go/auth"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/core/port"
)

// Server represents a gRPC server for auth operations.
type Server struct {
	server      *grpc.Server
	authService port.AuthService
	authv1.UnimplementedAuthServer
}

// NewServer creates a new instance of Server.
func NewServer(authService port.AuthService) *Server {
	return &Server{
		authService: authService,
	}
}

// Start starts the gRPC server on the specified port.
func (s *Server) Start(port string) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Printf("Error starting listener: %v", err)
		return err
	}

	s.server = grpc.NewServer()

	authv1.RegisterAuthServer(s.server, s)
	log.Printf("gRPC server started on :%s\n", port)
	go func() {
		if err := s.server.Serve(listen); err != nil {
			log.Printf("Error serving gRPC: %v", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server gracefully.
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
		log.Println("gRPC server stopped")
	}
}

// Login logs in the user with the specified username and password.
func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	log.Printf("Logging in user: %s\n", req.GetUsername())
	token, err := s.authService.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &authv1.LoginResponse{Token: token}, nil
}

// Register registers a new user.
func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	log.Printf("Registering user: %s\n", req.GetUsername())
	err := s.authService.Register(ctx, &domain.User{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Role:     domain.Role(req.GetRole()),
	})
	if err != nil {
		log.Println("Error registering user:", err)
		return nil, fmt.Errorf("failed to register: %w", err)
	}

	return &authv1.RegisterResponse{}, nil
}

// ValidateToken validates the specified token and returns the user details.
func (s *Server) ValidateToken(
	ctx context.Context,
	req *authv1.ValidateTokenRequest,
) (*authv1.ValidateTokenResponse, error) {
	user, err := s.authService.ValidateToken(ctx, req.GetToken())
	if err != nil {
		log.Println("Error validating token:", err)
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	return &authv1.ValidateTokenResponse{
		Username: user.Username,
		Password: user.Password,
		Role:     string(user.Role),
	}, nil
}
