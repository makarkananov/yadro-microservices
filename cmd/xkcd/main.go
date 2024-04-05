package main

import (
	"context"
	"flag"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
	"yadro-microservices/internal/service"
	"yadro-microservices/pkg/database"
	"yadro-microservices/pkg/words"
	"yadro-microservices/pkg/xkcd"
)

func main() {
	// Parse command line flags
	configPath := flag.String("c", "config.yaml", "Path to configuration file")
	outputFlag := flag.Bool("o", false, "Flag to output data to screen")
	maxComicsShown := flag.Int("n", 10, "Flag to limit the number of comics")
	flag.Parse()

	// Initialize and load configuration from file
	viper.SetConfigFile(*configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Error loading configuration:", err)
	}

	xkcdClient := xkcd.NewClient(viper.GetString("source_url"))
	db := database.NewJSONDB(viper.GetString("db_file"))
	processor := words.NewTextProcessor("en", "")
	xkcdService := service.NewXkcdService(xkcdClient, db, processor)

	maxComics := viper.GetInt("max_comics_load")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	comics, err := xkcdService.RetrieveAndSaveComics(ctx, maxComics)
	if err != nil {
		log.Panic("Error occurred in xkcdService:", err)
	}

	if *outputFlag {
		xkcdService.OutputComics(comics, *maxComicsShown)
	}
}
