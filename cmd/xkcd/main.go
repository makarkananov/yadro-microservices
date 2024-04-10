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
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize and load configuration from file
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Error loading configuration:", err)
	}

	xkcdClient := xkcd.NewClient(viper.GetString("source_url"))
	db := database.NewJSONDB(viper.GetString("db_file"))
	processor := words.NewTextProcessor("en", "")
	xkcdService := service.NewXkcdService(xkcdClient, db, processor)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	maxComics := viper.GetInt("max_comics_load")
	goroutinesLimit := viper.GetInt("parallel")
	gapsLimit := viper.GetUint32("gaps_limit")
	_, err := xkcdService.RetrieveAndSaveComics(ctx, maxComics, goroutinesLimit, gapsLimit)
	if err != nil {
		log.Panic("Error occurred in xkcdService:", err)
	}
}
