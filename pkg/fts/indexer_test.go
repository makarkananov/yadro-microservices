package fts_test

import (
	"context"
	"reflect"
	"testing"
	"yadro-microservices/pkg/fts"
	"yadro-microservices/pkg/fts/mock"
)

func TestInvertedIndexer_Add(t *testing.T) {
	mockRepo := mock.NewIndexRepository()
	indexer := fts.NewInvertedIndexer(mockRepo)

	doc1 := &fts.Document{ID: 1, Tokens: []string{"apple", "banana", "apple"}}
	doc2 := &fts.Document{ID: 2, Tokens: []string{"banana", "orange", "banana"}}

	err := indexer.Add(context.Background(), []*fts.Document{doc1, doc2})
	if err != nil {
		t.Errorf("Error creating inverted index: %v", err)
	}

	expectedIndexes := map[string][]*fts.Index{
		"apple": {
			{ID: 1, Score: 2},
		},
		"banana": {
			{ID: 1, Score: 1},
			{ID: 2, Score: 2},
		},
		"orange": {
			{ID: 2, Score: 1},
		},
	}

	if !reflect.DeepEqual(mockRepo.Indexes, expectedIndexes) {
		t.Errorf("Indexes after adding documents are not as expected")
	}
}

func TestInvertedIndexer_Get(t *testing.T) {
	mockRepo := mock.NewIndexRepository()
	indexer := fts.NewInvertedIndexer(mockRepo)

	doc1 := &fts.Document{ID: 1, Tokens: []string{"apple", "banana", "apple"}}
	doc2 := &fts.Document{ID: 2, Tokens: []string{"banana", "orange", "banana"}}

	err := indexer.Add(context.Background(), []*fts.Document{doc1, doc2})
	if err != nil {
		t.Errorf("Error creating inverted index: %v", err)
	}

	expectedIndexes := []*fts.Index{
		{ID: 1, Score: 1},
		{ID: 2, Score: 2},
	}

	indexes, err := indexer.Get(context.Background(), "banana")
	if err != nil {
		t.Errorf("Error getting indexes for 'banana' token: %v", err)
	}

	if !reflect.DeepEqual(indexes, expectedIndexes) {
		t.Errorf("Indexes retrieved for 'banana' token are not as expected")
	}
}

func TestInvertedIndexer_Add_EmptyDocument(t *testing.T) {
	mockRepo := mock.NewIndexRepository()
	indexer := fts.NewInvertedIndexer(mockRepo)

	doc := &fts.Document{ID: 1, Tokens: []string{}}

	err := indexer.Add(context.Background(), []*fts.Document{doc})
	if err != nil {
		t.Errorf("Error creating inverted index: %v", err)
	}

	if len(mockRepo.Indexes) != 0 {
		t.Errorf("Indexes should remain empty when adding an empty document")
	}
}

func TestInvertedIndexer_Add_NoTokens(t *testing.T) {
	mockRepo := mock.NewIndexRepository()
	indexer := fts.NewInvertedIndexer(mockRepo)

	doc := &fts.Document{ID: 1, Tokens: nil}

	err := indexer.Add(context.Background(), []*fts.Document{doc})
	if err != nil {
		t.Errorf("Error creating inverted index: %v", err)
	}

	if len(mockRepo.Indexes) != 0 {
		t.Errorf("Indexes should remain empty when adding a document with no tokens")
	}
}
