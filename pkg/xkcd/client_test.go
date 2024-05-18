package xkcd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetComic(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/1/info.0.json", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(ComicResponse{
			Num:        1,
			Title:      "Test Comic",
			Img:        "https://example.com/comic1.png",
			Transcript: "Test Transcript",
			Alt:        "Test Alt",
		})
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 1, 1, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	comic, err := client.getComic(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, comic)
	assert.Equal(t, 1, comic.Num)
	assert.Equal(t, "Test Comic", comic.Title)
	assert.Equal(t, "https://example.com/comic1.png", comic.Img)
	assert.Equal(t, "Test Transcript", comic.Transcript)
	assert.Equal(t, "Test Alt", comic.Alt)
}

func TestGetComic_NotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 1, 1, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	comic, err := client.getComic(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, comic)
}

func TestGetComic_Error(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 1, 1, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	comic, err := client.getComic(ctx, 1)
	require.Error(t, err)
	assert.Nil(t, comic)
}

func TestGetComics(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		comicID := 1
		if r.URL.Path == "/1/info.0.json" {
			comicID = 1
		}
		if r.URL.Path == "/2/info.0.json" {
			comicID = 2
		}
		switch {
		case comicID == 1:
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(ComicResponse{
				Num:        1,
				Title:      "Test Comic 1",
				Img:        "https://example.com/comic1.png",
				Transcript: "Test Transcript 1",
				Alt:        "Test Alt 1",
			})
			require.NoError(t, err)
		case comicID == 2:
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(ComicResponse{
				Num:        2,
				Title:      "Test Comic 2",
				Img:        "https://example.com/comic2.png",
				Transcript: "Test Transcript 2",
				Alt:        "Test Alt 2",
			})
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 2, 1, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	existingIDs := map[int]bool{
		3: true,
	}

	comics, err := client.GetComics(ctx, existingIDs)
	require.NoError(t, err)
	require.Len(t, comics, 2)

	assert.Equal(t, 1, comics[0].Num)
	assert.Equal(t, "Test Comic 1", comics[0].Title)
	assert.Equal(t, "https://example.com/comic1.png", comics[0].Img)
	assert.Equal(t, "Test Transcript 1", comics[0].Transcript)
	assert.Equal(t, "Test Alt 1", comics[0].Alt)

	assert.Equal(t, 2, comics[1].Num)
	assert.Equal(t, "Test Comic 2", comics[1].Title)
	assert.Equal(t, "https://example.com/comic2.png", comics[1].Img)
	assert.Equal(t, "Test Transcript 2", comics[1].Transcript)
	assert.Equal(t, "Test Alt 2", comics[1].Alt)
}

func TestGetComicsWithGaps(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/1/info.0.json" {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(ComicResponse{
				Num:        1,
				Title:      "Test Comic 1",
				Img:        "https://example.com/comic1.png",
				Transcript: "Test Transcript 1",
				Alt:        "Test Alt 1",
			})
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 2, 2, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	existingIDs := map[int]bool{}

	comics, err := client.GetComics(ctx, existingIDs)
	require.NoError(t, err)
	require.Len(t, comics, 1)

	assert.Equal(t, 1, comics[0].Num)
	assert.Equal(t, "Test Comic 1", comics[0].Title)
	assert.Equal(t, "https://example.com/comic1.png", comics[0].Img)
	assert.Equal(t, "Test Transcript 1", comics[0].Transcript)
	assert.Equal(t, "Test Alt 1", comics[0].Alt)
}

func TestGetComics_Error(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 2, 2, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	existingIDs := map[int]bool{}

	comics, err := client.GetComics(ctx, existingIDs)
	require.Error(t, err)
	require.Empty(t, comics, 0)
}

func TestGetComics_ContextCancelled(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(ComicResponse{
			Num:        1,
			Title:      "Test Comic 1",
			Img:        "https://example.com/comic1.png",
			Transcript: "Test Transcript 1",
			Alt:        "Test Alt 1",
		})
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, 2, 2, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	existingIDs := map[int]bool{}

	comics, err := client.GetComics(ctx, existingIDs)
	require.Error(t, err)
	require.Empty(t, comics, 0)
}
