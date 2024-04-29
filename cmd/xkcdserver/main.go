package main

import (
	"context"
	"flag"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	xkcdadapter "yadro-microservices/internal/adapter/client/xkcd"
	handler "yadro-microservices/internal/adapter/handler/http"
	"yadro-microservices/internal/adapter/repository/json"
	"yadro-microservices/internal/adapter/search"
	"yadro-microservices/internal/core/service"
	"yadro-microservices/pkg/database"
	"yadro-microservices/pkg/fts"
	"yadro-microservices/pkg/words"
	"yadro-microservices/pkg/xkcd"
)

func main() {
	// Parse command line flags
	var configPath string
	var port string
	flag.StringVar(&configPath, "c", "config.yaml", "Path to configuration file")
	flag.StringVar(&port, "p", "8080", "Port to start server on")
	flag.Parse()

	// Initialize and load configuration from file
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Error loading configuration:", err)
	}

	// Create context with cancel function
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create instances of necessary services and objects
	maxComics := viper.GetInt("max_comics_load")
	goroutinesLimit := viper.GetInt("parallel")
	gapsLimit := viper.GetUint32("gaps_limit")
	sourceURL := viper.GetString("source_url")
	xkcdClient := xkcd.NewClient(sourceURL, maxComics, goroutinesLimit, gapsLimit)
	comicsDB := database.NewJSONDB(viper.GetString("comics_db_file"))
	comicsRep := json.NewComicRepository(comicsDB)
	indexDB := database.NewJSONDB(viper.GetString("index_db_file"))
	indexRep := json.NewIndexRepository(indexDB)
	indexer := fts.NewInvertedIndexer()
	searcher := &fts.FullTextSearcher{}
	processor := words.NewTextProcessor("en", "extended_stopwords_eng.txt")
	searchEngine := search.NewFtsEngine(indexer, searcher, indexRep)
	comicClient := xkcdadapter.NewComicClient(xkcdClient, processor)

	// Initialize services
	err := searchEngine.Init()
	if err != nil {
		log.Panic("Error initializing Search Engine: ", err)
	}
	xkcdService := service.NewXkcdService(
		comicClient,
		comicsRep,
		processor,
		searchEngine,
	)
	err = xkcdService.Init()
	if err != nil {
		log.Panic("Error initializing xkcdService: ", err)
	}

	// Schedule comics update
	updateTimeStr := viper.GetString("update_time")
	updateTime, err := time.Parse("15:04", updateTimeStr)
	if err != nil {
		log.Panic("Error parsing update time:", err)
	}
	xkcdService.ScheduleUpdate(ctx, updateTime)

	// Configure HTTP server
	srv := &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Initialize HTTP handlers
	xkcdHandler := handler.NewXkcdHandler(xkcdService)
	http.HandleFunc("POST /update", xkcdHandler.Update)
	http.HandleFunc("GET /pics", xkcdHandler.Search)

	// Run the server
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		log.Printf("Server is running on port %s...", port)
		return srv.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Printf("Shut down signal received, shutting down server...")
		return srv.Shutdown(context.Background())
	})

	if err = g.Wait(); err != nil {
		log.Println("Exit reason:", err)
	}
}
