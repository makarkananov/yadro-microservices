package service

import (
	"context"
	"fmt"
	"log"
	"time"
	"yadro-microservices/internal/core/port"
)

// XkcdService provides methods for managing comics.
type XkcdService struct {
	client       port.ComicClient
	comicsRep    port.ComicRepository
	processor    port.ComicProcessor
	searchEngine port.SearchEngine
}

// NewXkcdService creates a new instance of XKCD service.
func NewXkcdService(
	client port.ComicClient,
	comicsRep port.ComicRepository,
	processor port.ComicProcessor,
	searchEngine port.SearchEngine,
) *XkcdService {
	return &XkcdService{
		client:       client,
		comicsRep:    comicsRep,
		processor:    processor,
		searchEngine: searchEngine,
	}
}

// ScheduleUpdate schedules comics update at a specific time.
func (xs *XkcdService) ScheduleUpdate(ctx context.Context, updateTime time.Time) {
	go func() {
		log.Println("Scheduling comics update...")
		updateTime = time.Date(
			time.Now().Year(),
			time.Now().Month(),
			time.Now().Day(),
			updateTime.Hour(),
			updateTime.Minute(),
			0,
			0,
			time.Local,
		)

		// Calculate the duration until the next update time
		durationUntilNextUpdate := time.Until(updateTime)
		// If the update time has already passed for today, set it for tomorrow
		if durationUntilNextUpdate < 0 {
			durationUntilNextUpdate += 24 * time.Hour
		}

		// Wait until the next update time
		timer := time.NewTimer(durationUntilNextUpdate)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				// If the context is canceled, stop the scheduling
				return
			case <-timer.C:
				// Perform the update and reset the timer for the next day
				log.Println("Scheduled comics update started...")
				err := xs.UpdateComics(ctx)
				if err != nil {
					log.Println("error while scheduled updating comics:", err)
				}

				timer.Reset(24 * time.Hour)
			}
		}
	}()
}

// UpdateComics retrieves comics from xkcd.com, processes them, and saves them to the database.
func (xs *XkcdService) UpdateComics(
	ctx context.Context,
) error {
	// Extract existing comic IDs into a map
	existingIDs, err := xs.comicsRep.GetAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("error extracting existing comic IDs: %w", err)
	}

	// Retrieve comics data from xkcd.com
	log.Println("Retrieving comics data from xkcd.com...")
	clientCtx, clientCancel := context.WithTimeout(ctx, 3*time.Minute)
	defer clientCancel()
	newComics, err := xs.client.GetComics(clientCtx, existingIDs)
	if err != nil {
		return fmt.Errorf("error retrieving comics data: %w", err)
	}

	// Save comics data to database
	log.Println("Saving comics data to database...")
	comicsRCtx, comicsCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer comicsCancel()
	if err = xs.comicsRep.Save(comicsRCtx, newComics); err != nil {
		return fmt.Errorf("error saving comics data to database: %w", err)
	}

	// Add comics to the search engine
	log.Println("Adding comics to search engine...")
	searchEngineCtx, searchEngineCancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer searchEngineCancel()
	err = xs.searchEngine.CreateIndex(searchEngineCtx, newComics)
	if err != nil {
		return fmt.Errorf("error adding comics to search engine: %w", err)
	}

	return nil
}

// Search searches for comics by the query and returns their URLs.
func (xs *XkcdService) Search(ctx context.Context, query string) ([]string, error) {
	queryTokens, err := xs.processor.FullProcess(query)
	if err != nil {
		return nil, fmt.Errorf("error processing query: %w", err)
	}

	ids, err := xs.searchEngine.Search(ctx, queryTokens)
	if err != nil {
		return nil, fmt.Errorf("error searching comics: %w", err)
	}

	urls := make([]string, 0, len(ids))
	for _, id := range ids {
		comic, err := xs.comicsRep.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error getting comic by ID: %w", err)
		}

		urls = append(urls, comic.Img)
	}

	return urls, nil
}

// GetNumberOfComics returns the total number of comics in the database.
func (xs *XkcdService) GetNumberOfComics(ctx context.Context) (int, error) {
	total, err := xs.comicsRep.GetTotalComics(ctx)
	if err != nil {
		return 0, fmt.Errorf("error getting total number of comics: %w", err)
	}

	return total, nil
}
