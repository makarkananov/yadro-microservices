package service

import (
	"context"
	"fmt"
	"log"
	"maps"
	"strings"
	"yadro-microservices/internal/core"
	"yadro-microservices/pkg/fts"
	"yadro-microservices/pkg/xkcd"
)

// ComicDatabase defines the interface for saving comic data to the database.
type ComicDatabase interface {
	Save(data any) error
	Load(v any) error
}

// IndexDatabase defines the interface for saving index data to the database.
type IndexDatabase interface {
	Save(data any) error
	Load(v any) error
}

// ComicProcessor defines the interface for processing text of the comic.
type ComicProcessor interface {
	FullProcess(text string) ([]string, error)
}

// XkcdService provides methods for managing comics.
type XkcdService struct {
	client    *xkcd.Client
	comicsDB  ComicDatabase
	indexDB   IndexDatabase
	processor ComicProcessor
	indexer   fts.Indexer
	searcher  fts.Searcher
	comics    map[int]*core.Comic
}

// NewXkcdService creates a new instance of XKCD service.
func NewXkcdService(
	client *xkcd.Client,
	comicsDB ComicDatabase,
	indexDB IndexDatabase,
	processor ComicProcessor,
	indexer fts.Indexer,
	searcher fts.Searcher,
) *XkcdService {
	return &XkcdService{
		client:    client,
		comicsDB:  comicsDB,
		indexDB:   indexDB,
		processor: processor,
		indexer:   indexer,
		searcher:  searcher,
	}
}

// RetrieveAndSaveComics retrieves comics from xkcd.com, processes them, and saves them to the database.
func (xs *XkcdService) RetrieveAndSaveComics(
	ctx context.Context,
	maxComics int,
	goroutinesLimit int,
	gapsLimit uint32,
) error {
	// Load existing comic IDs from the database
	log.Println("Loading existing comics data from database...")
	var existingComics map[int]*core.Comic
	if err := xs.comicsDB.Load(&existingComics); err != nil {
		log.Panic("Error loading comics data from JSON database:", err)
	}

	// Extract existing comic IDs into a map
	existingIDs := make(map[int]bool)
	for i := range existingComics {
		existingIDs[i] = true
	}

	log.Println("Retrieving comics data from xkcd.com...")
	comicsResponses, err := xs.client.GetComics(ctx, maxComics, existingIDs, goroutinesLimit, gapsLimit)
	if err != nil {
		log.Println("Error retrieving some comics data:", err)
	}

	// Convert XKCD comics data to internal representation using processor
	xs.comics = make(map[int]*core.Comic, len(comicsResponses))
	for i := range comicsResponses {
		comicText := strings.Join([]string{comicsResponses[i].Alt +
			comicsResponses[i].Transcript +
			comicsResponses[i].Title}, " ")

		kw, err := xs.processor.FullProcess(comicText)
		if err != nil {
			return fmt.Errorf("error extracting keywords: %w", err)
		}

		xs.comics[comicsResponses[i].Num] = &core.Comic{
			Img:      comicsResponses[i].Img,
			Keywords: kw,
		}
	}

	// Merge new comics with existing ones
	maps.Copy(xs.comics, existingComics)

	log.Println("Saving comics data to database...")
	// Save comics data to database
	if err = xs.comicsDB.Save(xs.comics); err != nil {
		return fmt.Errorf("error saving comics data to database: %w", err)
	}

	return nil
}

// Index updates the search index and saves it.
func (xs *XkcdService) Index() error {
	log.Println("Updating indexes...")
	err := xs.indexDB.Load(&xs.indexer)
	if err != nil {
		return fmt.Errorf("error loading index data from database: %w", err)
	}

	for id, comic := range xs.comics {
		xs.indexer.Add(&fts.Document{
			ID:     id,
			Tokens: comic.Keywords,
		})
	}

	err = xs.indexDB.Save(xs.indexer)
	if err != nil {
		return fmt.Errorf("error saving index data to database: %w", err)
	}

	return nil
}

// SearchUrlsWithIndex searches for comics by query using the search index.
func (xs *XkcdService) SearchUrlsWithIndex(query string, n int) ([]string, error) {
	queryTokens, err := xs.processor.FullProcess(query)
	if err != nil {
		return nil, fmt.Errorf("error processing query: %w", err)
	}

	log.Println("Searching... query tokens:", queryTokens)
	searchResults := xs.searcher.Search(queryTokens, fts.ThroughIndexes(xs.indexer), fts.ReturnMostRelevant(n))

	urls := make([]string, 0, len(searchResults))
	for _, searchResult := range searchResults {
		urls = append(urls, xs.comics[searchResult.ID].Img)
	}

	return urls, nil
}

// SearchUrls searches for comics by query.
func (xs *XkcdService) SearchUrls(query string, n int) ([]string, error) {
	queryTokens, err := xs.processor.FullProcess(query)
	log.Println("Searching trough indexes... query tokens:", queryTokens)
	if err != nil {
		return nil, fmt.Errorf("error processing query: %w", err)
	}

	docs := make([]*fts.Document, 0, len(xs.comics))
	for id, comic := range xs.comics {
		docs = append(docs, &fts.Document{
			ID:     id,
			Tokens: comic.Keywords,
		})
	}

	searchResults := xs.searcher.Search(queryTokens, fts.ThroughDocs(docs), fts.ReturnMostRelevant(n))

	urls := make([]string, 0, len(searchResults))
	for _, searchResult := range searchResults {
		urls = append(urls, xs.comics[searchResult.ID].Img)
	}

	return urls, nil
}
