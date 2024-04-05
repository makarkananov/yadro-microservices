package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"sync"
	"time"
)

// ComicResponse struct represents a single XKCD comic.
type ComicResponse struct {
	Num        int    `json:"num"`
	Title      string `json:"title"`
	Img        string `json:"img"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
}

// Client struct represents a client to interact with XKCD API.
type Client struct {
	baseURL string       // The base URL of the XKCD API
	client  *http.Client // HTTP client
}

// NewClient creates a new instance of XKCD client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// getComic retrieves information about a single comic by its ID.
func (c *Client) getComic(ctx context.Context, comicID int) (*ComicResponse, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, comicID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			// If it's a 404 error, it means there is simply no comic with this ID
			// Such behavior is expected and should not be treated as an error
			// Unfortunately, xkcd API does not provide a way to check if a comic exists and there are some gaps in IDs
			// Gap example: 404 joke on https://xkcd.com/404/info.0.json
			return nil, nil
		}

		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	var comic ComicResponse
	if err = json.NewDecoder(resp.Body).Decode(&comic); err != nil {
		return nil, err
	}

	return &comic, nil
}

// GetComics retrieves information about all XKCD comics.
func (c *Client) GetComics(ctx context.Context, maxID int, existingIDs map[int]struct{}) ([]*ComicResponse, error) {
	latestComic, err := c.GetLatest(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting last comic: %w", err)
	}

	totalComics := latestComic.Num
	if maxID == 0 || maxID > totalComics {
		maxID = totalComics
	}

	comics := make([]*ComicResponse, 0, maxID)
	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(30)

	for i := 1; i <= maxID; i++ {
		if _, ok := existingIDs[i]; ok {
			continue // Skip if the comic ID already exists
		}

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				comic, err := c.getComic(ctx, i)
				if err != nil {
					return err
				}
				if comic != nil {
					mu.Lock()
					comics = append(comics, comic)
					mu.Unlock()
				}
				return nil
			}
		})
	}

	if err = g.Wait(); err != nil {
		return comics, err
	}

	return comics, nil
}

// GetLatest retrieves information about the latest XKCD comic.
func (c *Client) GetLatest(ctx context.Context) (*ComicResponse, error) {
	url := fmt.Sprintf("%s/info.0.json", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	var comic ComicResponse
	if err = json.NewDecoder(resp.Body).Decode(&comic); err != nil {
		return nil, err
	}

	return &comic, nil
}
