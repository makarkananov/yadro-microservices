package mock

import (
	"context"
	"errors"
	"yadro-microservices/pkg/fts"
)

// IndexRepository is a mock implementation of the IndexRepository interface.
type IndexRepository struct {
	Indexes          map[string][]*fts.Index
	Documents        map[int]bool
	IndexedDocuments map[int]bool
}

func NewIndexRepository() *IndexRepository {
	return &IndexRepository{
		Indexes:          make(map[string][]*fts.Index),
		Documents:        make(map[int]bool),
		IndexedDocuments: make(map[int]bool),
	}
}

func (r *IndexRepository) Get(_ context.Context, word string) ([]*fts.Index, error) {
	indexList, found := r.Indexes[word]
	if !found {
		return nil, errors.New("word not found")
	}
	return indexList, nil
}

func (r *IndexRepository) Add(_ context.Context, indexes map[string][]*fts.Index, documents map[int]bool) error {
	for word, indexList := range indexes {
		r.Indexes[word] = append(r.Indexes[word], indexList...)
	}

	for id := range documents {
		r.Documents[id] = true
	}

	return nil
}

func (r *IndexRepository) DocumentIsIndexed(_ context.Context, id int) (bool, error) {
	return r.IndexedDocuments[id], nil
}

func (r *IndexRepository) MarkDocumentAsIndexed(_ context.Context, id int) error {
	r.IndexedDocuments[id] = true
	return nil
}
