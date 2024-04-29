package xkcd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/core/port"
	"yadro-microservices/pkg/xkcd"
)

type ComicClient struct {
	client    *xkcd.Client
	processor port.ComicProcessor
}

func NewComicClient(client *xkcd.Client, processor port.ComicProcessor) *ComicClient {
	return &ComicClient{
		client:    client,
		processor: processor,
	}
}

func (cc *ComicClient) GetComics(ctx context.Context, existingIDs map[int]bool) (domain.Comics, error) {
	comicsResponses, err := cc.client.GetComics(ctx, existingIDs)
	if err != nil {
		log.Println("Error retrieving some comics data:", err)
	}

	// Convert XKCD comics data to internal representation using processor
	comics := make(domain.Comics, len(comicsResponses))
	for i := range comicsResponses {
		comicText := strings.Join([]string{comicsResponses[i].Alt +
			comicsResponses[i].Transcript +
			comicsResponses[i].Title}, " ")

		kw, err := cc.processor.FullProcess(comicText)
		if err != nil {
			return nil, fmt.Errorf("error extracting keywords: %w", err)
		}

		comics[comicsResponses[i].Num] = &domain.Comic{
			Img:      comicsResponses[i].Img,
			Keywords: kw,
		}
	}

	return comics, nil
}
