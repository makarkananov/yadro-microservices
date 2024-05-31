package web

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"
)

// ComicHandler is html handler for comics.
type ComicHandler struct {
	searchURL string
}

// NewComicHandler creates new ComicHandler.
func NewComicHandler(searchURL string) *ComicHandler {
	return &ComicHandler{searchURL: searchURL}
}

// SearchComics searches comics by query and renders them to the page.
func (ch *ComicHandler) SearchComics(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("search") {
		log.Print("Rendering comics page")
		tmpl := template.Must(template.ParseFiles("templates/comics.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			log.Printf("Failed to render template: %s", err)
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Searching comics: %s", r.URL.Query().Get("search"))
	query := r.URL.Query().Get("search")
	if query == "" {
		log.Print("Empty search query")
		http.Error(w, "Empty search query", http.StatusBadRequest)
		return
	}

	tokenCookie, err := r.Cookie("token")
	if err != nil {
		log.Printf("Failed to get token: %s", err)
		http.Error(w, "Failed to get token", http.StatusUnauthorized)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, ch.searchURL+"?search="+query, nil)
	if err != nil {
		log.Printf("Failed create request: %s", err)
		http.Error(w, "Failed to search comics", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+tokenCookie.Value)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to do request: %s", err)
		http.Error(w, "Failed to search comics", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to search comics: %d", resp.StatusCode)
		http.Error(w, "Failed to search comics", resp.StatusCode)
		return
	}

	var urls []string
	if err := json.NewDecoder(resp.Body).Decode(&urls); err != nil {
		log.Printf("Failed to parse response: %s", err)
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d comics", len(urls))
	tmpl := template.Must(template.New("comics.html").ParseFiles("templates/comics.html"))
	err = tmpl.Execute(w, map[string]interface{}{
		"Comics": urls,
	})
	if err != nil {
		log.Printf("Failed to render template: %s", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}
