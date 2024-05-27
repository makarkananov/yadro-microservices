package service

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Login(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	userRepo.On("GetByUsername", mock.Anything, "valid_user").Return(
		&domain.User{Password: "$2a$10$/md3ztppcKhB9sjDb/GMZuYlb9o3bxvPnwO2v3up3/KlHCjMOskcG"},
		nil,
	).Once()

	authService := NewAuthService(userRepo, 3600)
	token, err := authService.Login(context.Background(), "valid_user", "password")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	userRepo.AssertExpectations(t)
}

func TestAuthService_LoginInvalidPassword(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	userRepo.On("GetByUsername", mock.Anything, "valid_user").Return(
		&domain.User{Password: "$2a$10$N9qo8uLOickgx2ZMRZoHKuGnK.y39JZjiujDtJZN.gR7Oy.fXx8aG"},
		nil,
	).Once()

	authService := NewAuthService(userRepo, 3600)
	_, err := authService.Login(context.Background(), "valid_user", "invalid_password")

	require.Error(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Register(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

	authService := NewAuthService(userRepo, 3600)
	err := authService.Register(context.Background(), &domain.User{Role: domain.ADMIN}, &domain.User{
		Username: "new_user",
		Password: "password",
		Role:     domain.USER,
	})

	require.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_RegisterUnauthorized(t *testing.T) {
	userRepo := new(mocks.UserRepository)

	authService := NewAuthService(userRepo, 3600)
	err := authService.Register(context.Background(), &domain.User{Role: domain.USER}, &domain.User{
		Username: "new_user",
		Password: "password",
		Role:     domain.ADMIN,
	})

	require.Error(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
	userRepo.On("GetByUsername", mock.Anything, "valid_user").Return(domain.NewUser(
		"valid_user",
		"$2a$10$/md3ztppcKhB9sjDb/GMZuYlb9o3bxvPnwO2v3up3/KlHCjMOskcG",
		domain.USER), nil).Twice()

	authService := NewAuthService(userRepo, 1*time.Hour)
	err := authService.Register(context.Background(), &domain.User{Role: domain.USER}, domain.NewUser(
		"valid_user",
		"password",
		domain.USER))
	require.NoError(t, err)

	tokenString, err := authService.Login(context.Background(), "valid_user", "password")
	require.NoError(t, err)

	user, err := authService.ValidateToken(context.Background(), tokenString)

	require.NoError(t, err)
	assert.NotNil(t, user)
	userRepo.AssertExpectations(t)
}

func TestAuthService_ValidateTokenInvalidFormat(t *testing.T) {
	userRepo := new(mocks.UserRepository)

	authService := NewAuthService(userRepo, 3600)
	_, err := authService.ValidateToken(context.Background(), "invalid_token")

	require.Error(t, err)
}

func TestAuthService_ValidateUserNotFound(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	userRepo.On("GetByUsername", mock.Anything, mock.Anything).Return(nil, nil)

	authService := NewAuthService(userRepo, 3600)
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "valid_user",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	require.NoError(t, err)
	_, err = authService.ValidateToken(context.Background(), tokenString)

	require.Error(t, err)
}

func TestAuthService_ValidateTokenInvalidClaims(t *testing.T) {
	userRepo := new(mocks.UserRepository)

	authService := NewAuthService(userRepo, 3600*time.Second)
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	require.NoError(t, err)

	_, err = authService.ValidateToken(context.Background(), tokenString)
	require.Error(t, err)
	assert.Equal(t, "invalid username in claims", err.Error())
}

func TestAuthService_ValidateTokenGetByUsernameError(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	userRepo.On(
		"GetByUsername",
		mock.Anything,
		"valid_user",
	).Return(nil, errors.New("db error")).Once()

	authService := NewAuthService(userRepo, 3600*time.Second)
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "valid_user",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	require.NoError(t, err)

	_, err = authService.ValidateToken(context.Background(), tokenString)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
	userRepo.AssertExpectations(t)
}
