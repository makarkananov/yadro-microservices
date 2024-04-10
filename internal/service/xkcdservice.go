package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"yadro-microservices/internal/core"
	"yadro-microservices/pkg/xkcd"
)

// ComicDatabase defines the interface for saving data to the database.
type ComicDatabase interface {
	Save(data interface{}) error
	Load(v interface{}) error
}

// ComicProcessor defines the interface for processing text of the comic.
type ComicProcessor interface {
	FullProcess(text string) ([]string, error)
}

// XkcdService provides methods for managing comics.
type XkcdService struct {
	client    *xkcd.Client
	db        ComicDatabase
	processor ComicProcessor
}

// NewXkcdService creates a new instance of XKCD service.
func NewXkcdService(client *xkcd.Client, db ComicDatabase, processor ComicProcessor) *XkcdService {
	return &XkcdService{
		client:    client,
		db:        db,
		processor: processor,
	}
}

// RetrieveAndSaveComics retrieves comics from xkcd.com, processes them, and saves them to the database.
func (xs *XkcdService) RetrieveAndSaveComics(
	ctx context.Context,
	maxComics int,
	goroutinesLimit int,
	gapsLimit uint32,
) (map[int]*core.Comic, error) {
	// Load existing comic IDs from the database
	log.Println("Loading existing comics data from database...")
	var existingComics map[int]*core.Comic
	if err := xs.db.Load(&existingComics); err != nil {
		log.Panic("Error loading comics data from JSON database:", err)
	}

	// Extract existing comic IDs into a map
	existingIDs := make(map[int]bool)
	for i := range existingComics {
		existingIDs[i+1] = true
	}

	log.Println("Retrieving comics data from xkcd.com...")
	comicsResponses, err := xs.client.GetComics(ctx, maxComics, existingIDs, goroutinesLimit, gapsLimit)
	if err != nil {
		fmt.Println("Error retrieving some comics data:", err)
	}

	// Convert XKCD comics data to internal representation using processor
	comicsMap := make(map[int]*core.Comic, len(comicsResponses))
	for i := range comicsResponses {
		comicText := strings.Join([]string{comicsResponses[i].Alt +
			comicsResponses[i].Transcript +
			comicsResponses[i].Title}, " ")

		kw, err := xs.processor.FullProcess(comicText)
		if err != nil {
			return nil, fmt.Errorf("error extracting keywords: %w", err)
		}

		comicsMap[comicsResponses[i].Num] = &core.Comic{
			Img:      comicsResponses[i].Img,
			Keywords: kw,
		}
	}

	// Merge new comics with existing ones
	for id, comic := range existingComics {
		if _, ok := comicsMap[id]; !ok {
			comicsMap[id] = comic
		}
	}

	log.Println("Saving comics data to database...")
	// Save comics data to database
	if err := xs.db.Save(comicsMap); err != nil {
		return nil, fmt.Errorf("error saving comics data to database: %w", err)
	}

	return comicsMap, nil
}
