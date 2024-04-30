package service

import (
	"context"
	"fmt"
	"log"
	"maps"
	"time"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/core/port"
)

// XkcdService provides methods for managing comics.
type XkcdService struct {
	client       port.ComicClient
	comicsRep    port.ComicRepository
	processor    port.ComicProcessor
	searchEngine port.SearchEngine
	comics       domain.Comics
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
		comics:       make(domain.Comics),
	}
}

// Init initializes the XKCD service.
func (xs *XkcdService) Init() error {
	// Load existing comic IDs from the database
	err := xs.loadComics()
	if err != nil {
		log.Panic("Error loading comics data:", err)
	}

	return nil
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
	existingIDs := make(map[int]bool)
	for i := range xs.comics {
		existingIDs[i] = true
	}

	// Retrieve comics data from xkcd.com
	log.Println("Retrieving comics data from xkcd.com...")
	newComics, err := xs.client.GetComics(ctx, existingIDs)
	if err != nil {
		return fmt.Errorf("error retrieving comics data: %w", err)
	}

	// Merge new comics with existing ones
	maps.Copy(xs.comics, newComics)

	// Add comics to the search engine
	err = xs.searchEngine.Add(xs.comics)
	if err != nil {
		return fmt.Errorf("error adding comics to search engine: %w", err)
	}

	// Save comics data to database
	err = xs.saveComics()
	if err != nil {
		return fmt.Errorf("error saving comics data: %w", err)
	}

	return nil
}

// Search searches for comics by the query and returns their URLs.
func (xs *XkcdService) Search(query string) ([]string, error) {
	queryTokens, err := xs.processor.FullProcess(query)
	if err != nil {
		return nil, fmt.Errorf("error processing query: %w", err)
	}

	ids, err := xs.searchEngine.Search(queryTokens)
	if err != nil {
		return nil, fmt.Errorf("error searching comics: %w", err)
	}

	urls := make([]string, 0, len(ids))
	for _, id := range ids {
		urls = append(urls, xs.comics[id].Img)
	}

	return urls, nil
}

func (xs *XkcdService) loadComics() error {
	log.Println("Loading comics data from database...")
	existingComics, err := xs.comicsRep.Load()
	if err != nil {
		return fmt.Errorf("error loading comics data from database: %w", err)
	}

	if existingComics != nil {
		xs.comics = existingComics
	}

	return nil
}

func (xs *XkcdService) saveComics() error {
	log.Println("Saving comics data to database...")
	if err := xs.comicsRep.Save(xs.comics); err != nil {
		return fmt.Errorf("error saving comics data to database: %w", err)
	}

	return nil
}

func (xs *XkcdService) GetNumberOfComics() int {
	return len(xs.comics)
}
