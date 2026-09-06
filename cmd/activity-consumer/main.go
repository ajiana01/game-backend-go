package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ajiana01/portfolio-go/internal/activity"
	"github.com/ajiana01/portfolio-go/internal/config"
	"github.com/ajiana01/portfolio-go/internal/event"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("consumer stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	startupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	client, err := mongo.Connect(startupCtx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())
	if err := client.Ping(startupCtx, readpref.Primary()); err != nil {
		return err
	}
	repo := activity.NewMongoRepository(client.Database(cfg.MongoDatabase))
	if err := repo.EnsureIndexes(startupCtx); err != nil {
		return err
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers, Topic: cfg.KafkaTopic,
		GroupID: "activity-consumer", MinBytes: 1, MaxBytes: 10e6,
	})
	defer reader.Close()
	logger.Info("activity consumer started", "topic", cfg.KafkaTopic)
	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		var value event.Event
		if err := json.Unmarshal(message.Value, &value); err != nil {
			logger.Error("discarding malformed event", "error", err, "offset", message.Offset)
			if err := reader.CommitMessages(ctx, message); err != nil {
				return err
			}
			continue
		}
		if err := repo.Store(ctx, value); err != nil {
			logger.Error("failed to store event", "error", err, "event_id", value.ID)
			continue
		}
		if err := reader.CommitMessages(ctx, message); err != nil {
			return err
		}
	}
}
