package fts_test

import (
	"context"
	"reflect"
	"testing"
	"yadro-microservices/pkg/fts"
)

func TestReturnMostRelevant(t *testing.T) {
	results := fts.SearchResults{
		{ID: 1, NumberOfTokens: 5, Score: 10},
		{ID: 2, NumberOfTokens: 3, Score: 8},
		{ID: 3, NumberOfTokens: 4, Score: 6},
		{ID: 4, NumberOfTokens: 2, Score: 4},
	}

	expectedResults := fts.SearchResults{
		{ID: 1, NumberOfTokens: 5, Score: 10},
		{ID: 3, NumberOfTokens: 4, Score: 6},
	}

	modifier := fts.ReturnMostRelevant(2)
	err := modifier(nil, &results)
	if err != nil {
		t.Errorf("ReturnMostRelevant modifier returned an error: %v", err)
	}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("ReturnMostRelevant modifier did not return the most relevant results")
	}
}

func TestThroughIndexes(t *testing.T) {
	mockIndexer := &MockIndexer{
		Indexes: map[string][]*fts.Index{
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
		},
	}

	results := fts.SearchResults{}

	modifier := fts.ThroughIndexes(context.Background(), mockIndexer)
	err := modifier([]string{"apple", "banana"}, &results)
	if err != nil {
		t.Errorf("ThroughIndexes modifier returned an error: %v", err)
	}

	expectedResults := fts.SearchResults{
		{ID: 1, NumberOfTokens: 2, Score: 3},
		{ID: 2, NumberOfTokens: 1, Score: 2},
	}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("ThroughIndexes modifier did not return the expected results")
	}
}

func TestSearch(t *testing.T) {
	mockIndexer := &MockIndexer{
		Indexes: map[string][]*fts.Index{
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
		},
	}

	searcher := fts.FullTextSearcher{}

	results, _ := searcher.Search(
		[]string{"apple", "banana"},
		fts.ThroughIndexes(context.Background(), mockIndexer),
		fts.ReturnMostRelevant(2),
	)

	expectedResults := []int{1, 2}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("Search did not return the expected results")
	}
}

func TestThroughDocs(t *testing.T) {
	docs := []*fts.Document{
		{ID: 1, Tokens: []string{"apple", "banana", "apple"}},
		{ID: 2, Tokens: []string{"banana", "orange", "banana"}},
	}

	results := fts.SearchResults{}

	modifier := fts.ThroughDocs(docs)
	err := modifier([]string{"apple", "banana"}, &results)
	if err != nil {
		t.Errorf("ThroughDocs modifier returned an error: %v", err)
	}

	expectedResults := fts.SearchResults{
		{ID: 1, NumberOfTokens: 2, Score: 3},
		{ID: 2, NumberOfTokens: 1, Score: 2},
	}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("ThroughDocs modifier did not return the expected results")
	}
}

type MockIndexer struct {
	Indexes map[string][]*fts.Index
}

func (m MockIndexer) Add(_ *fts.Document) error {
	return nil
}

func (m MockIndexer) Get(_ context.Context, token string) ([]*fts.Index, error) {
	return m.Indexes[token], nil
}
