package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"github.com/go-redis/redis/v8"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
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
	"yadro-microservices/internal/adapter/repository/pg"
	redisrep "yadro-microservices/internal/adapter/repository/redis"
	"yadro-microservices/internal/adapter/search"
	"yadro-microservices/internal/core/service"
	"yadro-microservices/internal/migrations"
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

	// Add context with cancel function
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Add xkcd client
	maxComics := viper.GetInt("max_comics_load")
	goroutinesLimit := viper.GetInt("parallel")
	gapsLimit := viper.GetUint32("gaps_limit")
	sourceURL := viper.GetString("source_url")
	xkcdClient := xkcd.NewClient(sourceURL, maxComics, goroutinesLimit, gapsLimit)
	processor := words.NewTextProcessor("en", "extended_stopwords_eng.txt")
	comicClient := xkcdadapter.NewComicClient(xkcdClient, processor)

	// Add postgres client and repository, apply migrations
	postgresURL := viper.GetString("postgres_url")
	pgClient, err := sql.Open("postgres", postgresURL)
	if err != nil {
		log.Panic("Error connecting to the database:", err)
	}
	err = applyMigrations(postgresURL)
	if err != nil {
		log.Panic("Error applying migrations:", err)
	}
	comicsRep := pg.NewComicRepository(pgClient)

	// Add redis client and repository
	opt, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		log.Panic("Error parsing redis url:", err)
	}
	redisClient := redis.NewClient(opt)
	indexRep := redisrep.NewIndexRepository(redisClient)

	// Add search engine
	indexer := fts.NewInvertedIndexer(indexRep)
	searcher := &fts.FullTextSearcher{}
	searchEngine := search.NewFtsEngine(indexer, searcher)

	// Add xkcd service
	xkcdService := service.NewXkcdService(
		comicClient,
		comicsRep,
		processor,
		searchEngine,
	)

	// Schedule comics update
	updateTimeStr := viper.GetString("update_time")
	updateTime, err := time.Parse("15:04", updateTimeStr)
	if err != nil {
		log.Panic("Error parsing update time:", err)
	}
	xkcdService.ScheduleUpdate(ctx, updateTime)

	// Initialize http mux and handlers
	mux := http.NewServeMux()
	xkcdHandler := handler.NewXkcdHandler(xkcdService)
	mux.HandleFunc("POST /update", xkcdHandler.Update)
	mux.HandleFunc("GET /pics", xkcdHandler.Search)

	// Configure HTTP server
	srv := &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

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

func applyMigrations(dbURL string) error {
	log.Println("Trying to apply migrations...")

	d, err := iofs.New(migrations.FS, "pg")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
