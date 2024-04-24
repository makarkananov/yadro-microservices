package fts

import (
	"reflect"
	"testing"
)

func TestInvertedIndexer_Add(t *testing.T) {
	indexer := NewInvertedIndexer()

	doc1 := &Document{ID: 1, Tokens: []string{"apple", "banana", "apple"}}
	doc2 := &Document{ID: 2, Tokens: []string{"banana", "orange", "banana"}}

	indexer.Add(doc1)
	indexer.Add(doc2)

	expectedIndexes := map[string][]*Index{
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

	if !reflect.DeepEqual(indexer.Indexes, expectedIndexes) {
		t.Errorf("Indexes after adding documents are not as expected")
	}
}

func TestInvertedIndexer_Get(t *testing.T) {
	indexer := NewInvertedIndexer()

	doc1 := &Document{ID: 1, Tokens: []string{"apple", "banana", "apple"}}
	doc2 := &Document{ID: 2, Tokens: []string{"banana", "orange", "banana"}}

	indexer.Add(doc1)
	indexer.Add(doc2)

	expectedIndexes := []*Index{
		{ID: 1, Score: 1},
		{ID: 2, Score: 2},
	}

	indexes := indexer.Get("banana")

	if !reflect.DeepEqual(indexes, expectedIndexes) {
		t.Errorf("Indexes retrieved for 'banana' token are not as expected")
	}
}

func TestInvertedIndexer_Get_NotFound(t *testing.T) {
	indexer := NewInvertedIndexer()

	doc1 := &Document{ID: 1, Tokens: []string{"apple", "banana", "apple"}}
	doc2 := &Document{ID: 2, Tokens: []string{"banana", "orange", "banana"}}

	indexer.Add(doc1)
	indexer.Add(doc2)

	indexes := indexer.Get("grape")

	if len(indexes) != 0 {
		t.Errorf("Unexpected indexes retrieved for 'grape' token")
	}
}

func TestInvertedIndexer_Get_EmptyIndex(t *testing.T) {
	indexer := NewInvertedIndexer()

	indexes := indexer.Get("banana")

	if len(indexes) != 0 {
		t.Errorf("Unexpected indexes retrieved for 'banana' token from an empty index")
	}
}

func TestInvertedIndexer_Add_EmptyDocument(t *testing.T) {
	indexer := NewInvertedIndexer()

	doc := &Document{ID: 1, Tokens: []string{}}

	indexer.Add(doc)

	if len(indexer.Indexes) != 0 {
		t.Errorf("Indexes should remain empty when adding an empty document")
	}
}

func TestInvertedIndexer_Add_NoTokens(t *testing.T) {
	indexer := NewInvertedIndexer()

	doc := &Document{ID: 1, Tokens: nil}

	indexer.Add(doc)

	if len(indexer.Indexes) != 0 {
		t.Errorf("Indexes should remain empty when adding a document with no tokens")
	}
}
