package activity

import (
	"context"

	"github.com/ajiana01/portfolio-go/internal/event"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct{ collection *mongo.Collection }

func NewMongoRepository(database *mongo.Database) *MongoRepository {
	return &MongoRepository{collection: database.Collection("activity_log")}
}

func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "event_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "occurred_at", Value: -1}}},
	})
	return err
}

func (r *MongoRepository) Store(ctx context.Context, value event.Event) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"event_id": value.ID},
		bson.M{"$setOnInsert": value},
		options.Update().SetUpsert(true),
	)
	return err
}
