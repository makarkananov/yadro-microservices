package xkcd

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/pkg/words"
	"yadro-microservices/pkg/xkcd"
)

func TestComicClient_GetComics(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		comic := xkcd.ComicResponse{
			Num:        1,
			Title:      "Test Comic.",
			Img:        "https://example.com/comic.png",
			Transcript: "Test Transcription.",
			Alt:        "Test Alt.",
		}
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(comic)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := xkcd.NewClient(mockServer.URL, 10, 1, 1)
	processor := words.NewTextProcessor("en", "")
	cc := NewComicClient(client, processor)

	ctx := context.Background()
	existingIDs := map[int]bool{1: true}
	comics, err := cc.GetComics(ctx, existingIDs)

	require.NoError(t, err)
	assert.Len(t, comics, 1)
	assert.Equal(t, "https://example.com/comic.png", comics[1].Img)
	assert.ElementsMatch(t, []string{"test", "alt", "test", "transcript", "test", "comic"}, comics[1].Keywords)
}

func TestComicClient_GetComics_NoError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	client := xkcd.NewClient(mockServer.URL, 10, 2, 1)
	processor := words.NewTextProcessor("en", "")
	cc := NewComicClient(client, processor)

	ctx := context.Background()
	existingIDs := map[int]bool{}
	comics, err := cc.GetComics(ctx, existingIDs)

	require.NoError(t, err)
	assert.Equal(t, domain.Comics{}, comics)
}

func TestComicClient_GetComics_WithGaps(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "1") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			comic := xkcd.ComicResponse{
				Num:        2,
				Title:      "Test Comic.",
				Img:        "https://example.com/comic.png",
				Transcript: "Test Transcription.",
				Alt:        "Test Alt.",
			}
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(comic)
			require.NoError(t, err)
		}
	}))
	defer mockServer.Close()

	client := xkcd.NewClient(mockServer.URL, 10, 2, 1)
	processor := words.NewTextProcessor("en", "")
	cc := NewComicClient(client, processor)

	ctx := context.Background()
	existingIDs := map[int]bool{}
	comics, err := cc.GetComics(ctx, existingIDs)

	require.NoError(t, err)
	assert.Len(t, comics, 1)
	assert.Equal(t, "https://example.com/comic.png", comics[2].Img)
	assert.ElementsMatch(t, []string{"test", "alt", "test", "transcript", "test", "comic"}, comics[2].Keywords)
}
