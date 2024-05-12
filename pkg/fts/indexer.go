package fts

import (
	"context"
	"fmt"
)

// Index is a struct that represents the score for document with specific ID.
type Index struct {
	ID    int `json:"id"`
	Score int `json:"counter"`
}

// Indexer is an interface that defines the behavior of an indexing engine.
type Indexer interface {
	Get(ctx context.Context, token string) ([]*Index, error)
}

// IndexRepository is an interface that defines the behavior of a repository that stores indexes.
type IndexRepository interface {
	Get(ctx context.Context, word string) ([]*Index, error)
	Add(ctx context.Context, indexes map[string][]*Index, documents map[int]bool) error
	DocumentIsIndexed(ctx context.Context, id int) (bool, error)
	MarkDocumentAsIndexed(ctx context.Context, id int) error
}

// InvertedIndexer is an implementation of the Indexer interface that uses an inverted index.
type InvertedIndexer struct {
	IndexRep IndexRepository
}

// NewInvertedIndexer creates a new InvertedIndexer.
func NewInvertedIndexer(indexRep IndexRepository) *InvertedIndexer {
	return &InvertedIndexer{
		IndexRep: indexRep,
	}
}

// Add creates an inverted index from the given documents.
func (i *InvertedIndexer) Add(ctx context.Context, docs []*Document) error {
	indexes := make(map[string][]*Index)
	indexedDocuments := make(map[int]bool)

	for _, doc := range docs {
		isIndexed, err := i.IndexRep.DocumentIsIndexed(ctx, doc.ID)
		if err != nil {
			return fmt.Errorf("error checking if document is indexed: %w", err)
		}
		if isIndexed { // If the document is already indexed, skip it
			continue
		}

		for _, token := range doc.Tokens {
			if _, ok := indexes[token]; !ok {
				indexes[token] = []*Index{
					{
						ID:    doc.ID,
						Score: 1,
					},
				}

				continue
			}

			found := false
			for _, index := range indexes[token] {
				if index.ID == doc.ID {
					found = true
					index.Score++
				}
			}

			if !found {
				indexes[token] = append(indexes[token], &Index{
					ID:    doc.ID,
					Score: 1,
				})
			}
		}

		indexedDocuments[doc.ID] = true
	}

	err := i.IndexRep.Add(ctx, indexes, indexedDocuments)
	if err != nil {
		return fmt.Errorf("error saving index to db: %w", err)
	}

	return nil
}

// Get returns the indexes for a given token.
func (i *InvertedIndexer) Get(ctx context.Context, token string) ([]*Index, error) {
	index, err := i.IndexRep.Get(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("error getting indexes for token %s: %w", token, err)
	}

	return index, nil
}
