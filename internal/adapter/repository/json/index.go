package json

import (
	"fmt"
	"yadro-microservices/pkg/database"
	"yadro-microservices/pkg/fts"
)

type IndexRepository struct {
	db *database.JSONDB
}

func NewIndexRepository(db *database.JSONDB) *IndexRepository {
	return &IndexRepository{db: db}
}

func (ir *IndexRepository) Save(indexer *fts.InvertedIndexer) error {
	err := ir.db.Save(indexer)
	if err != nil {
		return fmt.Errorf("failed to save inverted indexer: %w", err)
	}

	return nil
}

func (ir *IndexRepository) Load(v *fts.InvertedIndexer) error {
	err := ir.db.Load(v)
	if err != nil {
		return fmt.Errorf("failed to load inverted indexer: %w", err)
	}

	return nil
}
