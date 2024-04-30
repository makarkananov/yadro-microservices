package json

import (
	"fmt"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/pkg/database"
)

type ComicRepository struct {
	db *database.JSONDB
}

func NewComicRepository(db *database.JSONDB) *ComicRepository {
	return &ComicRepository{db: db}
}

func (cr ComicRepository) Save(c domain.Comics) error {
	err := cr.db.Save(c)
	if err != nil {
		return fmt.Errorf("failed to save comic data: %w", err)
	}

	return nil
}

func (cr ComicRepository) Load() (domain.Comics, error) {
	var existingComics domain.Comics
	if err := cr.db.Load(&existingComics); err != nil {
		return nil, fmt.Errorf("error loading comics data from JSON database: %w", err)
	}

	return existingComics, nil
}
