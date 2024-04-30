package search

import (
	"fmt"
	"log"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/pkg/fts"
)

// IndexRepository defines the interface for saving index data to the database.
type IndexRepository interface {
	Save(indexer *fts.InvertedIndexer) error
	Load(v *fts.InvertedIndexer) error
}

// FtsEngine provides methods for full-text search.
type FtsEngine struct {
	searcher *fts.FullTextSearcher
	indexer  *fts.InvertedIndexer
	indexRep IndexRepository
}

// NewFtsEngine creates a new instance of FTS engine.
func NewFtsEngine(indexer *fts.InvertedIndexer, searcher *fts.FullTextSearcher, indexRep IndexRepository) *FtsEngine {
	return &FtsEngine{
		indexer:  indexer,
		searcher: searcher,
		indexRep: indexRep,
	}
}

// Init initializes the FTS engine.
func (fe *FtsEngine) Init() error {
	err := fe.indexRep.Load(fe.indexer)
	if err != nil {
		return fmt.Errorf("error loading index data from database: %w", err)
	}

	return nil
}

// Search searches for documents ids by query tokens.
func (fe *FtsEngine) Search(queryTokens []string) ([]int, error) {
	log.Println("Searching... Query tokens:", queryTokens)

	searchResults := fe.searcher.Search(queryTokens, fts.ThroughIndexes(fe.indexer), fts.ReturnMostRelevant(10))

	return searchResults, nil
}

// Add adds documents to the index.
func (fe *FtsEngine) Add(comics domain.Comics) error {
	for id, comic := range comics {
		err := fe.indexer.Add(&fts.Document{
			ID:     id,
			Tokens: comic.Keywords,
		})

		if err != nil {
			return fmt.Errorf("error adding document to index: %w", err)
		}
	}

	err := fe.indexRep.Save(fe.indexer)
	if err != nil {
		return fmt.Errorf("error saving index data to database: %w", err)
	}

	return nil
}
