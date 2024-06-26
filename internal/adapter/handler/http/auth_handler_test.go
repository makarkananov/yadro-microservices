package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/mocks"
)

func TestAuthHandler_Login(t *testing.T) {
	authClient := new(mocks.AuthClient)
	authClient.On(
		"Login",
		mock.Anything,
		"valid_user",
		"valid_pass",
	).Return("valid_token", nil).Once()

	handler := NewAuthHandler(authClient)
	creds := map[string]string{
		"username": "valid_user",
		"password": "valid_pass",
	}
	credsBytes, _ := json.Marshal(creds)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(credsBytes))
	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	authClient.AssertExpectations(t)
}

func TestAuthHandler_LoginInvalidCredentials(t *testing.T) {
	authClient := new(mocks.AuthClient)
	authClient.On(
		"Login",
		mock.Anything,
		"invalid_user",
		"invalid_pass",
	).Return("", errors.New("invalid")).Once()

	handler := NewAuthHandler(authClient)
	creds := map[string]string{
		"username": "invalid_user",
		"password": "invalid_pass",
	}
	credsBytes, _ := json.Marshal(creds)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(credsBytes))
	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	authClient.AssertExpectations(t)
}

func TestAuthHandler_Register(t *testing.T) {
	authClient := new(mocks.AuthClient)
	authClient.On(
		"Register",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil).Once()

	handler := NewAuthHandler(authClient)
	creds := map[string]string{
		"username": "valid_user",
		"password": "valid_pass",
		"role":     "admin",
	}
	credsBytes, _ := json.Marshal(creds)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(credsBytes))
	req = req.WithContext(context.WithValue(req.Context(), currentUserKey, &domain.User{Role: domain.ADMIN}))
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	authClient.AssertExpectations(t)
}

func TestAuthHandler_RegisterInvalidRole(t *testing.T) {
	authClient := new(mocks.AuthClient)

	handler := NewAuthHandler(authClient)
	creds := map[string]string{
		"username": "valid_user",
		"password": "valid_pass",
		"role":     "unknown_role",
	}
	credsBytes, _ := json.Marshal(creds)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(credsBytes))
	req = req.WithContext(context.WithValue(req.Context(), currentUserKey, &domain.User{Role: domain.ADMIN}))
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	authClient.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidRole(t *testing.T) {
	authClient := new(mocks.AuthClient)

	handler := NewAuthHandler(authClient)
	creds := map[string]string{
		"username": "valid_user",
		"password": "valid_pass",
		"role":     "unknown_role",
	}
	credsBytes, _ := json.Marshal(creds)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(credsBytes))
	req = req.WithContext(context.WithValue(req.Context(), currentUserKey, &domain.User{Role: domain.ADMIN}))
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	authClient.AssertExpectations(t)
}

func TestAuthHandler_Register_DecodeError(t *testing.T) {
	authClient := new(mocks.AuthClient)

	handler := NewAuthHandler(authClient)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{invalid json}"))
	req = req.WithContext(context.WithValue(req.Context(), currentUserKey, &domain.User{Role: domain.ADMIN}))
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	authClient.AssertExpectations(t)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	authClient := new(mocks.AuthClient)

	handler := NewAuthHandler(authClient)
	creds := map[string]string{
		"username": "us",
		"password": "pass",
		"role":     "admin",
	}
	credsBytes, _ := json.Marshal(creds)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(credsBytes))
	req = req.WithContext(context.WithValue(req.Context(), currentUserKey, &domain.User{Role: domain.ADMIN}))
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	authClient.AssertExpectations(t)
}

func TestAuthHandler_Register_ServiceError(t *testing.T) {
	authClient := new(mocks.AuthClient)
	authClient.On(
		"Register",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(errors.New("service error")).Once()

	handler := NewAuthHandler(authClient)
	creds := map[string]string{
		"username": "valid_user",
		"password": "valid_pass",
		"role":     "admin",
	}
	credsBytes, _ := json.Marshal(creds)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(credsBytes))
	req = req.WithContext(context.WithValue(req.Context(), currentUserKey, &domain.User{Role: domain.ADMIN}))
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	authClient.AssertExpectations(t)
}
