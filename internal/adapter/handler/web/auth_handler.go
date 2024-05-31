package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"
)

// AuthHandler is html handler for authentication.
type AuthHandler struct {
	loginURL string
}

// NewAuthHandler creates new AuthHandler.
func NewAuthHandler(loginURL string) *AuthHandler {
	return &AuthHandler{loginURL: loginURL}
}

// LoginForm renders login page.
func (ah *AuthHandler) LoginForm(w http.ResponseWriter, _ *http.Request) {
	log.Println("Rendering login page")
	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Failed to render template: %s", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// Login logs in user and sets token cookie if successful. Redirects to comics page on success.
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("Logging in user: %s", r.FormValue("username"))
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	creds := struct {
		Username string `json:"Username"`
		Password string `json:"Password"`
	}{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	credsJSON, err := json.Marshal(creds)
	if err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, ah.loginURL, bytes.NewBuffer(credsJSON))
	if err != nil {
		log.Printf("Failed create request: %s", err)
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to do request: %s", err)
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Invalid response: %d", resp.StatusCode)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: result["token"],
		Path:  "/",
	})
	http.Redirect(w, r, "/comics", http.StatusSeeOther)
}
