package http

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"yadro-microservices/internal/core/domain"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, author *domain.User, newUser *domain.User) error {
	args := m.Called(ctx, author, newUser)
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
}

func TestAuthorizationMiddlewareWithAdminRole(t *testing.T) {
	handler := AuthorizationMiddleware(domain.ADMIN)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), currentUserKey, &domain.User{Role: domain.ADMIN}))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthorizationMiddlewareWithUserRole(t *testing.T) {
	handler := AuthorizationMiddleware(domain.ADMIN)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), currentUserKey, &domain.User{Role: domain.USER}))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAuthenticationMiddlewareWithValidToken(t *testing.T) {
	authService := new(MockAuthService)
	authService.On("ValidateToken", mock.Anything, "valid_token").Return(&domain.User{}, nil).Once()

	var handler = AuthenticationMiddleware(authService, true)(http.HandlerFunc(func(
		_ http.ResponseWriter,
		_ *http.Request,
	) {
	}))
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	authService.AssertExpectations(t)
}

func TestAuthenticationMiddlewareWithInvalidToken(t *testing.T) {
	authService := new(MockAuthService)
	authService.On("ValidateToken", mock.Anything, "invalid_token").Return(nil, http.ErrNoCookie).Once()

	handler := AuthenticationMiddleware(authService, true)(http.HandlerFunc(func(
		_ http.ResponseWriter,
		_ *http.Request,
	) {
	}))
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	authService.AssertExpectations(t)
}

func TestAuthenticationMiddlewareWithoutToken(t *testing.T) {
	authService := new(MockAuthService)

	handler := AuthenticationMiddleware(authService, true)(http.HandlerFunc(func(
		_ http.ResponseWriter,
		_ *http.Request,
	) {
	}))
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
