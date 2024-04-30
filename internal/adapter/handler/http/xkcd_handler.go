package http

import (
	"encoding/json"
	"log"
	"net/http"
	"yadro-microservices/internal/core/port"
)

type XkcdHandler struct {
	service port.ComicService
}

func NewXkcdHandler(service port.ComicService) *XkcdHandler {
	return &XkcdHandler{service: service}
}

func (xh *XkcdHandler) Update(w http.ResponseWriter, r *http.Request) {
	before := xh.service.GetNumberOfComics()
	err := xh.service.UpdateComics(r.Context())
	if err != nil {
		log.Printf("Error updating comics: %v", err)
		http.Error(w, "Failed to update comics", http.StatusInternalServerError)
		return
	}
	after := xh.service.GetNumberOfComics()

	response := struct {
		NewComics   int `json:"new"`
		TotalComics int `json:"total"`
	}{
		NewComics:   after - before,
		TotalComics: after,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (xh *XkcdHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("search")
	if query == "" {
		http.Error(w, "Empty search query", http.StatusBadRequest)
		return
	}

	urls, err := xh.service.Search(query)
	log.Println("Search results:", urls)
	if err != nil {
		log.Printf("Error searching comics: %v", err)
		http.Error(w, "Failed to search comics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
