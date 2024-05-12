package search

import (
	"context"
	"fmt"
	"log"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/pkg/fts"
)

// FtsEngine provides methods for full-text search.
type FtsEngine struct {
	searcher *fts.FullTextSearcher
	indexer  *fts.InvertedIndexer
}

// NewFtsEngine creates a new instance of FTS engine.
func NewFtsEngine(indexer *fts.InvertedIndexer, searcher *fts.FullTextSearcher) *FtsEngine {
	return &FtsEngine{
		indexer:  indexer,
		searcher: searcher,
	}
}

// Search searches for documents ids by query tokens.
func (fe *FtsEngine) Search(ctx context.Context, queryTokens []string) ([]int, error) {
	log.Println("Searching... Query tokens:", queryTokens)

	searchResults, err := fe.searcher.Search(queryTokens, fts.ThroughIndexes(ctx, fe.indexer), fts.ReturnMostRelevant(10))
	if err != nil {
		return nil, fmt.Errorf("error searching for documents: %w", err)
	}

	return searchResults, nil
}

// CreateIndex builds index based on comics.
func (fe *FtsEngine) CreateIndex(ctx context.Context, comics domain.Comics) error {
	log.Println("Adding documents to the index...")
	var docs []*fts.Document

	for id, comic := range comics {
		docs = append(docs, &fts.Document{
			ID:     id,
			Tokens: comic.Keywords,
		})
	}

	err := fe.indexer.Add(ctx, docs)
	if err != nil {
		return fmt.Errorf("error creating index with documents: %w", err)
	}

	return nil
}
