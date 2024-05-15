package service

import (
	"context"
	"testing"
	"yadro-microservices/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Save(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestAuthService_Login(t *testing.T) {
	userRepo := new(MockUserRepository)
	userRepo.On("GetByUsername", mock.Anything, "valid_user").Return(
		&domain.User{Password: "$2a$10$/md3ztppcKhB9sjDb/GMZuYlb9o3bxvPnwO2v3up3/KlHCjMOskcG"},
		nil,
	).Once()

	authService := NewAuthService(userRepo, 3600)
	token, err := authService.Login(context.Background(), "valid_user", "password")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	userRepo.AssertExpectations(t)
}

func TestAuthService_LoginInvalidPassword(t *testing.T) {
	userRepo := new(MockUserRepository)
	userRepo.On("GetByUsername", mock.Anything, "valid_user").Return(
		&domain.User{Password: "$2a$10$N9qo8uLOickgx2ZMRZoHKuGnK.y39JZjiujDtJZN.gR7Oy.fXx8aG"},
		nil,
	).Once()

	authService := NewAuthService(userRepo, 3600)
	_, err := authService.Login(context.Background(), "valid_user", "invalid_password")

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Register(t *testing.T) {
	userRepo := new(MockUserRepository)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

	authService := NewAuthService(userRepo, 3600)
	err := authService.Register(context.Background(), &domain.User{Role: domain.ADMIN}, &domain.User{
		Username: "new_user",
		Password: "password",
		Role:     domain.USER,
	})

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_RegisterUnauthorized(t *testing.T) {
	userRepo := new(MockUserRepository)

	authService := NewAuthService(userRepo, 3600)
	err := authService.Register(context.Background(), &domain.User{Role: domain.USER}, &domain.User{
		Username: "new_user",
		Password: "password",
		Role:     domain.ADMIN,
	})

	assert.Error(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
	userRepo.On("GetByUsername", mock.Anything, "valid_user").Return(&domain.User{
		Username: "valid_user",
		Role:     domain.USER,
		Password: "$2a$10$/md3ztppcKhB9sjDb/GMZuYlb9o3bxvPnwO2v3up3/KlHCjMOskcG",
	}, nil).Twice()

	authService := NewAuthService(userRepo, 3600)
	err := authService.Register(context.Background(), &domain.User{Role: domain.USER}, &domain.User{
		Username: "valid_user",
		Password: "password",
		Role:     domain.USER,
	})
	assert.NoError(t, err)

	tokenString, err := authService.Login(context.Background(), "valid_user", "password")
	assert.NoError(t, err)

	user, err := authService.ValidateToken(context.Background(), tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	userRepo.AssertExpectations(t)
}

func TestAuthService_ValidateTokenInvalid(t *testing.T) {
	userRepo := new(MockUserRepository)

	authService := NewAuthService(userRepo, 3600)
	_, err := authService.ValidateToken(context.Background(), "invalid_token")

	assert.Error(t, err)
}
