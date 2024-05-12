package fts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"yadro-microservices/pkg/fts"

	"github.com/go-redis/redis/v8"
)

// IndexRepository implements the fts.IndexRepository interface.
type IndexRepository struct {
	client *redis.Client
}

// NewIndexRepository returns a new IndexRepository.
func NewIndexRepository(client *redis.Client) *IndexRepository {
	return &IndexRepository{
		client: client,
	}
}

// Get retrieves indexes for a word from Redis.
func (r *IndexRepository) Get(ctx context.Context, word string) ([]*fts.Index, error) {
	vals, err := r.client.HGetAll(ctx, word).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes for word %s: %w", word, err)
	}

	var indexes []*fts.Index
	for idStr, val := range vals {
		var index fts.Index
		if err = json.Unmarshal([]byte(val), &index); err != nil {
			return nil, fmt.Errorf("failed to unmarshal index for word %s: %w", word, err)
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert index ID to integer: %w", err)
		}
		index.ID = id
		indexes = append(indexes, &index)
	}

	return indexes, nil
}

// Add efficiently saves indexes and documents to Redis.
func (r *IndexRepository) Add(
	ctx context.Context,
	indexes map[string][]*fts.Index,
	documents map[int]bool,
) error {
	pipe := r.client.Pipeline()
	defer pipe.Close()

	existingIndexesMap := make(map[string][]*fts.Index)

	// Getting existing indexes for words
	wordsToRetrieve := make([]string, 0, len(indexes))
	for word := range indexes {
		pipe.HGetAll(ctx, word)
		wordsToRetrieve = append(wordsToRetrieve, word)
	}

	results, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	// Unpacking existing indexes
	for i, word := range wordsToRetrieve {
		cmd := results[i]
		vals, err := cmd.(*redis.StringStringMapCmd).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return fmt.Errorf("failed to get indexes for word %s: %w", word, err)
		}

		var existingIndexes []*fts.Index
		for idStr, val := range vals {
			var index fts.Index
			if err = json.Unmarshal([]byte(val), &index); err != nil {
				return fmt.Errorf("failed to unmarshal index for word %s: %w", word, err)
			}

			id, err := strconv.Atoi(idStr)
			if err != nil {
				return fmt.Errorf("failed to convert index ID to integer: %w", err)
			}

			index.ID = id
			existingIndexes = append(existingIndexes, &index)
		}

		existingIndexesMap[word] = existingIndexes
	}

	// Adding new indexes to existing ones and saving them
	for word, indexList := range indexes {
		existingIndexes := existingIndexesMap[word]
		existingIndexes = append(existingIndexes, indexList...)

		pipe.Del(ctx, word)
		for _, index := range existingIndexes {
			data, err := json.Marshal(index)
			if err != nil {
				return fmt.Errorf("failed to marshal index for word %s: %w", word, err)
			}
			pipe.HSet(ctx, word, strconv.Itoa(index.ID), data)
		}
	}

	// Add indexed documents
	for id := range documents {
		pipe.SAdd(ctx, "indexed_documents", strconv.Itoa(id))
	}

	// Executing the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set indexes and documents: %w", err)
	}

	return nil
}

// DocumentIsIndexed checks if a document with the given ID is indexed in Redis.
func (r *IndexRepository) DocumentIsIndexed(ctx context.Context, id int) (bool, error) {
	return r.client.SIsMember(ctx, "indexed_documents", strconv.Itoa(id)).Result()
}

// MarkDocumentAsIndexed marks a document with the given ID as indexed in Redis.
func (r *IndexRepository) MarkDocumentAsIndexed(ctx context.Context, id int) error {
	return r.client.SAdd(ctx, "indexed_documents", strconv.Itoa(id)).Err()
}
