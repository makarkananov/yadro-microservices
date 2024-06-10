package main

import (
	"context"
	"flag"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
	"yadro-microservices/cmd/xkcdserver/launcher"
	"yadro-microservices/internal/adapter/client/auth"
)

func main() {
	// Parse command line flags
	var configPath string
	var port string
	flag.StringVar(&configPath, "c", "config/xkcdserver.yaml", "Path to configuration file")
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

	// Initialize services and server
	pgClient := launcher.NewPostgresClient()
	redisClient := launcher.NewRedisClient()
	xkcdService := launcher.NewXkcdService(ctx, pgClient, redisClient)
	authClient, err := auth.NewClient(viper.GetString("auth_server_url"))
	if err != nil {
		log.Panic("Error creating auth client:", err)
	}
	srv := launcher.NewServer(ctx, xkcdService, authClient, port)

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

	if err := g.Wait(); err != nil {
		log.Println("Exit reason:", err)
	}
}
