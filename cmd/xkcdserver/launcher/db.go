package launcher

import (
	"database/sql"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Required for migrations
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq" // Required for postgres
	"github.com/spf13/viper"
	"log"
	"yadro-microservices/internal/migrations"
)

// NewPostgresClient creates a new instance of the Postgres client.
func NewPostgresClient() *sql.DB {
	postgresURL := viper.GetString("postgres_url")
	pgClient, err := sql.Open("postgres", postgresURL)
	if err != nil {
		log.Panic("Error connecting to the database:", err)
	}
	err = applyMigrations(postgresURL)
	if err != nil {
		log.Panic("Error applying migrations:", err)
	}

	return pgClient
}

// NewRedisClient creates a new instance of the Redis client.
func NewRedisClient() *redis.Client {
	opt, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		log.Panic("Error parsing redis url:", err)
	}
	redisClient := redis.NewClient(opt)

	return redisClient
}

// applyMigrations applies all available migrations to the database.
func applyMigrations(dbURL string) error {
	log.Println("Trying to apply migrations...")

	d, err := iofs.New(migrations.FS, "pg/xkcd")
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
