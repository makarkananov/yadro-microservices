package launcher

import (
	"context"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"time"
	"yadro-microservices/internal/adapter/client/auth"
	handler "yadro-microservices/internal/adapter/handler/http"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/core/service"
	"yadro-microservices/pkg/middleware"
)

// NewServer creates a new instance of the HTTP server with the specified services and port.
func NewServer(
	ctx context.Context,
	xkcdService *service.XkcdService,
	authClient *auth.Client,
	port string,
) *http.Server {
	// Initialize http mux and handlers
	mux := http.NewServeMux()
	xkcdHandler := handler.NewXkcdHandler(xkcdService)
	authHandler := handler.NewAuthHandler(authClient)
	mux.HandleFunc("POST /update", middleware.Chain(
		xkcdHandler.Update,
		handler.AuthenticationMiddleware(authClient, true),
		handler.AuthorizationMiddleware(domain.ADMIN),
	))
	mux.HandleFunc("GET /pics", middleware.Chain(
		xkcdHandler.Search,
		handler.AuthenticationMiddleware(authClient, true),
		handler.AuthorizationMiddleware(domain.USER),
	))
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /register", middleware.Chain(
		authHandler.Register,
		handler.AuthenticationMiddleware(authClient, true),
		handler.AuthorizationMiddleware(domain.ADMIN),
	))

	rl := middleware.NewRateLimiter(viper.GetInt64("rate_limit"), viper.GetInt64("max_tokens"))
	cl := middleware.NewConcurrencyLimiter(viper.GetInt("concurrency_limit"))

	// Configure HTTP server
	srv := &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		Addr:              ":" + port,
		Handler:           middleware.Chain(mux.ServeHTTP, rl.Limit, cl.Limit),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return srv
}
