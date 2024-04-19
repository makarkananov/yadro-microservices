package service_test

import (
	"testing"
	"yadro-microservices/internal/core"
	"yadro-microservices/pkg/database"
	"yadro-microservices/pkg/fts"
)

func BenchmarkThroughDocs(b *testing.B) {
	comicsDB := database.NewJSONDB("../../database.json")
	searcher := fts.FullTextSearcher{}

	queryTokens := []string{"apple", "banana"}

	var comics map[int]*core.Comic
	err := comicsDB.Load(&comics)
	if err != nil {
		b.Fatalf("Failed to load comics: %v", err)
	}

	docs := make([]*fts.Document, 0, len(comics))
	for id, comic := range comics {
		docs = append(docs, &fts.Document{
			ID:     id,
			Tokens: comic.Keywords,
		})
	}

	for i := 0; i < b.N; i++ {
		_ = searcher.Search(queryTokens, fts.ThroughDocs(docs))
	}
}

func BenchmarkThroughIndexes(b *testing.B) {
	comicsDB := database.NewJSONDB("../../database.json")
	indexDB := database.NewJSONDB("../../index.json")
	indexer := fts.NewInvertedIndexer()
	searcher := fts.FullTextSearcher{}

	queryTokens := []string{"apple", "banana"}

	var comics map[int]*core.Comic
	err := comicsDB.Load(&comics)
	if err != nil {
		b.Fatalf("Failed to load comics: %v", err)
	}

	err = indexDB.Load(&indexer)
	if err != nil {
		b.Fatalf("Failed to load indexes: %v", err)
	}

	for i := 0; i < b.N; i++ {
		_ = searcher.Search(queryTokens, fts.ThroughIndexes(indexer))
	}
}
