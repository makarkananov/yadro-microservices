package search

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/pkg/fts"
	"yadro-microservices/pkg/fts/mock"

	"github.com/stretchr/testify/assert"
)

func TestFtsEngine_CreateIndexAndSearch(t *testing.T) {
	indexRepo := mock.NewIndexRepository()
	indexer := fts.NewInvertedIndexer(indexRepo)
	searcher := &fts.FullTextSearcher{}
	engine := NewFtsEngine(indexer, searcher)

	comics := domain.Comics{
		1: {Keywords: []string{"comic", "funny"}},
		2: {Keywords: []string{"comic", "sad"}},
		3: {Keywords: []string{"funny", "story"}},
	}

	ctx := context.Background()
	err := engine.CreateIndex(ctx, comics)
	require.NoError(t, err)

	queryTokens := []string{"comic"}
	results, err := engine.Search(ctx, queryTokens)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 2}, results)

	queryTokens = []string{"funny"}
	results, err = engine.Search(ctx, queryTokens)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 3}, results)
}

func TestFtsEngine_Search_NoResults(t *testing.T) {
	indexRepo := mock.NewIndexRepository()
	indexer := fts.NewInvertedIndexer(indexRepo)
	searcher := &fts.FullTextSearcher{}
	engine := NewFtsEngine(indexer, searcher)

	comics := domain.Comics{
		1: {Keywords: []string{"comic", "funny"}},
		2: {Keywords: []string{"comic", "sad"}},
		3: {Keywords: []string{"funny", "story"}},
	}

	ctx := context.Background()
	err := engine.CreateIndex(ctx, comics)
	require.NoError(t, err)

	queryTokens := []string{"missing"}
	results, err := engine.Search(ctx, queryTokens)
	require.Error(t, err)
	assert.Empty(t, results)
}

func TestFtsEngine_CreateIndex_ErrorHandling(t *testing.T) {
	indexRepo := mock.NewIndexRepository()
	indexer := fts.NewInvertedIndexer(indexRepo)
	searcher := &fts.FullTextSearcher{}
	engine := NewFtsEngine(indexer, searcher)

	comics := domain.Comics{
		1: {Keywords: []string{"comic", "funny"}},
		2: {Keywords: []string{"comic", "sad"}},
		3: {Keywords: []string{"funny", "story"}},
	}

	ctx := context.Background()
	err := engine.CreateIndex(ctx, comics)
	require.NoError(t, err)

	err = engine.CreateIndex(ctx, comics)
	require.NoError(t, err)
}
