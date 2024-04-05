package service

import (
	"context"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"log"
	"os"
	"sort"
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
func (xs *XkcdService) RetrieveAndSaveComics(ctx context.Context, maxComics int) (map[int]*core.Comic, error) {
	// Load existing comic IDs from the database
	var existingComics map[int]*core.Comic
	if err := xs.db.Load(&existingComics); err != nil {
		log.Panic("Error loading comics data from JSON database:", err)
	}

	// Extract existing comic IDs into a map
	existingIDs := make(map[int]struct{})
	for i := range existingComics {
		existingIDs[i+1] = struct{}{}
	}

	comicsResponses, err := xs.client.GetComics(ctx, maxComics, existingIDs)
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

	// Save comics data to database
	if err := xs.db.Save(comicsMap); err != nil {
		return nil, fmt.Errorf("error saving comics data to database: %w", err)
	}

	return comicsMap, nil
}

// OutputComics outputs the comics to the console.
func (xs *XkcdService) OutputComics(comicsMap map[int]*core.Comic, maxComicsShown int) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Img URL", "Keywords"})

	// Sort keys of comicsMap
	sortedKeys := make([]int, 0, len(comicsMap))
	for num := range comicsMap {
		sortedKeys = append(sortedKeys, num)
	}
	sort.Ints(sortedKeys)

	// Iterate over sorted keys
	for _, num := range sortedKeys {
		t.AppendRow([]interface{}{
			num,
			comicsMap[num].Img,
			text.WrapText(strings.Join(comicsMap[num].Keywords, " "), 70),
		})
		t.AppendSeparator()

		// Break the loop if maxComicsShown is reached
		if t.Length() >= maxComicsShown {
			break
		}
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}
