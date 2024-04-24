package main

import (
	"context"
	"flag"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
	"yadro-microservices/internal/service"
	"yadro-microservices/pkg/database"
	"yadro-microservices/pkg/fts"
	"yadro-microservices/pkg/words"
	"yadro-microservices/pkg/xkcd"
)

func main() {
	// Parse command line flags
	var configPath string
	var searchQuery string
	var useIndex bool
	flag.StringVar(&configPath, "c", "config.yaml", "Path to configuration file")
	flag.StringVar(&searchQuery, "s", "Hello World!", "Search query")
	flag.BoolVar(&useIndex, "i", false, "Shows if index should be used for search")
	flag.Parse()

	// Initialize and load configuration from file
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Error loading configuration:", err)
	}

	// Create instances of necessary services and objects
	xkcdClient := xkcd.NewClient(viper.GetString("source_url"))
	comicsDB := database.NewJSONDB(viper.GetString("comics_db_file"))
	indexDB := database.NewJSONDB(viper.GetString("index_db_file"))
	indexer := fts.NewInvertedIndexer()
	searcher := &fts.FullTextSearcher{}
	processor := words.NewTextProcessor("en", "extended_stopwords_eng.txt")
	xkcdService := service.NewXkcdService(xkcdClient, comicsDB, indexDB, processor, indexer, searcher)

	// Create a context that cancels when an interrupt or SIGTERM signal is caught
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Retrieve and save comics
	maxComics := viper.GetInt("max_comics_load")
	goroutinesLimit := viper.GetInt("parallel")
	gapsLimit := viper.GetUint32("gaps_limit")
	err := xkcdService.RetrieveAndSaveComics(ctx, maxComics, goroutinesLimit, gapsLimit)
	if err != nil {
		log.Panic("Error occurred while retrieving and saving comics:", err)
	}

	var urls []string
	// Use index for search if useIndex flag is true
	if useIndex {
		err = xkcdService.Index()
		if err != nil {
			log.Panic("Error occurred while indexing comics:", err)
		}

		urls, err = xkcdService.SearchUrlsWithIndex(searchQuery, 10)
		if err != nil {
			log.Panic("Error occurred while searching comics with index:", err)
		}
	} else {
		urls, err = xkcdService.SearchUrls(searchQuery, 10)
		if err != nil {
			log.Panic("Error occurred while searching comics:", err)
		}
	}

	// Print the search results
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Img URL"})

	for i, url := range urls {
		t.AppendRow([]interface{}{
			i + 1,
			url,
		})
		t.AppendSeparator()
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}
