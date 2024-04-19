package fts

import (
	"reflect"
	"testing"
)

func TestReturnMostRelevant(t *testing.T) {
	results := SearchResults{
		{ID: 1, NumberOfTokens: 5, Score: 10},
		{ID: 2, NumberOfTokens: 3, Score: 8},
		{ID: 3, NumberOfTokens: 4, Score: 6},
		{ID: 4, NumberOfTokens: 2, Score: 4},
	}

	expectedResults := SearchResults{
		{ID: 1, NumberOfTokens: 5, Score: 10},
		{ID: 3, NumberOfTokens: 4, Score: 6},
	}

	modifier := ReturnMostRelevant(2)
	modifier(nil, &results)

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("ReturnMostRelevant modifier did not return the most relevant results")
	}
}

func TestThroughIndexes(t *testing.T) {
	mockIndexer := &MockIndexer{
		Indexes: map[string][]*Index{
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

	results := SearchResults{}

	modifier := ThroughIndexes(mockIndexer)
	modifier([]string{"apple", "banana"}, &results)

	expectedResults := SearchResults{
		{ID: 1, NumberOfTokens: 2, Score: 3},
		{ID: 2, NumberOfTokens: 1, Score: 2},
	}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("ThroughIndexes modifier did not return the expected results")
	}
}

func TestSearch(t *testing.T) {
	mockIndexer := &MockIndexer{
		Indexes: map[string][]*Index{
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

	searcher := FullTextSearcher{}

	results := searcher.Search([]string{"apple", "banana"}, ThroughIndexes(mockIndexer), ReturnMostRelevant(2))

	expectedResults := SearchResults{
		{ID: 1, NumberOfTokens: 2, Score: 3},
		{ID: 2, NumberOfTokens: 1, Score: 2},
	}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("Search did not return the expected results")
	}
}

func TestThroughDocs(t *testing.T) {
	docs := []*Document{
		{ID: 1, Tokens: []string{"apple", "banana", "apple"}},
		{ID: 2, Tokens: []string{"banana", "orange", "banana"}},
	}

	results := SearchResults{}

	modifier := ThroughDocs(docs)
	modifier([]string{"apple", "banana"}, &results)

	expectedResults := SearchResults{
		{ID: 1, NumberOfTokens: 2, Score: 3},
		{ID: 2, NumberOfTokens: 1, Score: 2},
	}

	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("ThroughDocs modifier did not return the expected results")
	}
}

type MockIndexer struct {
	Indexes map[string][]*Index
}

func (m MockIndexer) Add(_ *Document) {}

func (m MockIndexer) Get(token string) []*Index {
	return m.Indexes[token]
}
