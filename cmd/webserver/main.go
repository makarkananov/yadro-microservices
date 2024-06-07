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
	"yadro-microservices/internal/adapter/handler/web"
)

func main() {
	// Parse command line flags
	var configPath string
	var port string
	flag.StringVar(&configPath, "c", "config.yaml", "Path to configuration file")
	flag.StringVar(&port, "p", "8081", "Port to start server on")
	flag.Parse()

	// Initialize and load configuration from file
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Error loading configuration:", err)
	}

	// Add context with cancel function
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize http mux and handlers
	mux := http.NewServeMux()
	authHandler := web.NewAuthHandler(
		viper.GetString("auth_url"),
		time.Duration(viper.GetInt("token_max_time"))*time.Minute,
	)
	comicsHandler := web.NewComicHandler(viper.GetString("comics_url"))

	mux.HandleFunc("GET /comics", comicsHandler.SearchComics)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("GET /login", authHandler.LoginForm)

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

	if err := g.Wait(); err != nil {
		log.Println("Exit reason:", err)
	}
}
