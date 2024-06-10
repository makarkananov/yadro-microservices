package launcher

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"log"
	"time"
	xkcdadapter "yadro-microservices/internal/adapter/client/xkcd"
	"yadro-microservices/internal/adapter/repository/pg"
	redisrep "yadro-microservices/internal/adapter/repository/redis"
	"yadro-microservices/internal/adapter/search"
	"yadro-microservices/internal/core/service"
	"yadro-microservices/pkg/fts"
	"yadro-microservices/pkg/words"
	"yadro-microservices/pkg/xkcd"
)

// NewXkcdService creates a new instance of the XkcdService.
func NewXkcdService(ctx context.Context, pgClient *sql.DB, redisClient *redis.Client) *service.XkcdService {
	maxComics := viper.GetInt("max_comics_load")
	goroutinesLimit := viper.GetInt("parallel")
	gapsLimit := viper.GetUint32("gaps_limit")
	sourceURL := viper.GetString("source_url")

	// Add comic client
	xkcdClient := xkcd.NewClient(sourceURL, maxComics, goroutinesLimit, gapsLimit)
	processor := words.NewTextProcessor("en", "config/extended_stopwords_eng.txt")
	comicClient := xkcdadapter.NewComicClient(xkcdClient, processor)

	// Add repositories
	comicsRep := pg.NewComicRepository(pgClient)
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

	return xkcdService
}
