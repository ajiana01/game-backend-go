package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ajiana01/portfolio-go/internal/config"
	"github.com/ajiana01/portfolio-go/internal/event"
	"github.com/ajiana01/portfolio-go/internal/inventory"
	"github.com/ajiana01/portfolio-go/internal/leaderboard"
	"github.com/ajiana01/portfolio-go/internal/player"
	"github.com/ajiana01/portfolio-go/internal/reward"
	"github.com/ajiana01/portfolio-go/internal/server"
	"github.com/ajiana01/portfolio-go/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// @title Game Backend Services API
// @version 1.0
// @description Modular monolith backend for player profiles, inventory, rewards, and leaderboards.
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @BasePath /
// @schemes http https
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	startupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	postgres, err := pgxpool.New(startupCtx, cfg.PostgresURL)
	if err != nil {
		return err
	}
	defer postgres.Close()
	if err := postgres.Ping(startupCtx); err != nil {
		return err
	}
	if err := migrations.Up(startupCtx, postgres); err != nil {
		return err
	}

	mongoClient, err := mongo.Connect(startupCtx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return err
	}
	defer mongoClient.Disconnect(context.Background())
	if err := mongoClient.Ping(startupCtx, readpref.Primary()); err != nil {
		return err
	}
	database := mongoClient.Database(cfg.MongoDatabase)

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	defer redisClient.Close()
	if err := redisClient.Ping(startupCtx).Err(); err != nil {
		return err
	}

	playerRepo := player.NewPostgresRepository(postgres)
	playerService := player.NewService(playerRepo)
	inventoryRepo := inventory.NewMongoRepository(database)
	if err := inventoryRepo.EnsureIndexes(startupCtx); err != nil {
		return err
	}
	inventoryService := inventory.NewService(inventoryRepo, playerRepo)
	publisher := event.NewKafkaPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer publisher.Close()
	rewardService := reward.NewService(reward.NewPostgresRepository(postgres), inventoryService, publisher)
	leaderboardService := leaderboard.NewService(leaderboard.NewRedisRepository(redisClient), playerService)

	router := server.NewRouter(
		logger,
		player.NewHandler(playerService),
		inventory.NewHandler(inventoryService),
		reward.NewHandler(rewardService),
		leaderboard.NewHandler(leaderboardService),
		func(r *http.Request) error {
			checkCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := postgres.Ping(checkCtx); err != nil {
				return err
			}
			if err := mongoClient.Ping(checkCtx, readpref.Primary()); err != nil {
				return err
			}
			return redisClient.Ping(checkCtx).Err()
		},
	)
	server := &http.Server{
		Addr: cfg.HTTPAddr, Handler: router,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("api listening", "address", cfg.HTTPAddr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
