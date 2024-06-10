package http

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/core/port"
)

// AuthHandler provides methods for handling auth requests.
type AuthHandler struct {
	authClient port.AuthClient
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(authClient port.AuthClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

// Login handles login requests.
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username" validate:"required,min=5,max=20"`
		Password string `json:"password" validate:"required,min=5,max=20"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(creds)
	if err != nil {
		log.Printf("Error validating request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := ah.authClient.Login(r.Context(), creds.Username, creds.Password)
	if err != nil {
		log.Printf("Error logging in: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err = json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

// Register handles register requests.
func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username" validate:"required,min=5,max=20"`
		Password string `json:"password" validate:"required,min=5,max=20"`
		Role     string `json:"role" validate:"required,oneof=admin user"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err = validate.Struct(creds)
	if err != nil {
		log.Printf("Error validating request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = ah.authClient.Register(
		r.Context(),
		creds.Username,
		creds.Password,
		domain.Role(creds.Role),
	)

	if err != nil {
		log.Printf("Error registering user: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
