package launcher

import (
	"database/sql"
	"github.com/spf13/viper"
	"time"
	"yadro-microservices/internal/adapter/repository/pg"
	"yadro-microservices/internal/core/service"
)

// NewAuthService creates a new instance of the AuthService.
func NewAuthService(pgClient *sql.DB) *service.AuthService {
	tokenMaxTime := viper.GetInt("token_max_time")
	usersRep := pg.NewUserRepository(pgClient)
	authService := service.NewAuthService(usersRep, time.Duration(tokenMaxTime)*time.Minute)

	return authService
}
