package port

import (
	"context"
	"yadro-microservices/internal/core/domain"
)

// ComicRepository defines the interface for saving comic data to the database.
type ComicRepository interface {
	Save(ctx context.Context, c domain.Comics) error
	GetAll(ctx context.Context) (domain.Comics, error)
	GetAllIDs(ctx context.Context) (map[int]bool, error)
	GetByID(ctx context.Context, id int) (*domain.Comic, error)
	GetTotalComics(ctx context.Context) (int, error)
}

// ComicProcessor defines the interface for processing text of the comic.
type ComicProcessor interface {
	FullProcess(text string) ([]string, error)
}

// SearchEngine defines the interface for a search engine.
type SearchEngine interface {
	Search(ctx context.Context, queryTokens []string) ([]int, error)
	CreateIndex(ctx context.Context, comics domain.Comics) error
}

// ComicService defines the interface for the comic service.
type ComicService interface {
	UpdateComics(ctx context.Context) error
	Search(ctx context.Context, query string) ([]string, error)
	GetNumberOfComics(ctx context.Context) (int, error)
}

// ComicClient defines the interface for the comic client.
type ComicClient interface {
	GetComics(ctx context.Context, existingIDs map[int]bool) (domain.Comics, error)
}
