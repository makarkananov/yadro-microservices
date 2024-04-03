package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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
	baseURL string // The base URL of the XKCD API
}

// NewClient creates a new instance of XKCD client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
	}

}

// getComic retrieves information about a single comic by its ID.
func (c *Client) getComic(comicID int) (*ComicResponse, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, comicID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
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
func (c *Client) GetComics(ctx context.Context, n int) ([]*ComicResponse, error) {
	latestComic, err := c.GetLatest()
	if err != nil {
		return nil, fmt.Errorf("error getting last comic: %w", err)
	}

	totalComics := latestComic.Num
	if n == 0 || n > totalComics {
		n = totalComics
	}

	comics := make([]*ComicResponse, 0, n)
	var mu sync.Mutex
	errors := make(chan error)
	done := make(chan bool)

	semaphore := make(chan struct{}, 30)
	var wg sync.WaitGroup

	for i := 1; i <= n; i++ {
		semaphore <- struct{}{} // Block the semaphore if it is full
		wg.Add(1)
		go func(i int) {
			defer func() {
				<-semaphore // Release the semaphore when the goroutine finishes
				wg.Done()
			}()
			select {
			case <-ctx.Done():
				return
			default:
				comic, err := c.getComic(i)
				if err != nil {
					errors <- err
					return
				}

				if comic != nil {
					mu.Lock()
					comics = append(comics, comic)
					mu.Unlock()
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return comics, ctx.Err()
		case err := <-errors:
			return comics, err
		case <-done:
			return comics, nil
		}
	}
}

// GetLatest retrieves information about the latest XKCD comic.
func (c *Client) GetLatest() (*ComicResponse, error) {
	url := fmt.Sprintf("%s/info.0.json", c.baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
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
