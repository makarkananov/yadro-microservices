package port

import (
	"context"
	"yadro-microservices/internal/core/domain"
)

// ComicRepository defines the interface for saving comic data to the database.
type ComicRepository interface {
	Save(c domain.Comics) error
	Load() (domain.Comics, error)
}

// ComicProcessor defines the interface for processing text of the comic.
type ComicProcessor interface {
	FullProcess(text string) ([]string, error)
}

// SearchEngine defines the interface for a search engine.
type SearchEngine interface {
	Search(queryTokens []string) ([]int, error)
	Add(comics domain.Comics) error
}

// ComicService defines the interface for the comic service.
type ComicService interface {
	UpdateComics(ctx context.Context) error
	Search(query string) ([]string, error)
	GetNumberOfComics() int
}

// ComicClient defines the interface for the comic client.
type ComicClient interface {
	GetComics(ctx context.Context, existingIDs map[int]bool) (domain.Comics, error)
}
