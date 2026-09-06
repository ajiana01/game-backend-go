package inventory

import (
	"context"
	"errors"
	"time"

	"github.com/ajiana01/portfolio-go/internal/domainerr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct{ collection *mongo.Collection }

func NewMongoRepository(database *mongo.Database) *MongoRepository {
	return &MongoRepository{collection: database.Collection("inventory")}
}

func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "player_id", Value: 1}, {Key: "item_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *MongoRepository) List(ctx context.Context, playerID string) ([]Item, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"player_id": playerID}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []struct {
		ItemID    string    `bson:"item_id"`
		Name      string    `bson:"name"`
		Quantity  int       `bson:"quantity"`
		UpdatedAt time.Time `bson:"updated_at"`
	}
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(docs))
	for _, doc := range docs {
		items = append(items, Item{ItemID: doc.ItemID, Name: doc.Name, Quantity: doc.Quantity, UpdatedAt: doc.UpdatedAt})
	}
	return items, nil
}

func (r *MongoRepository) Add(ctx context.Context, playerID string, input AddInput) (Item, error) {
	now := time.Now().UTC()
	filter := bson.M{"player_id": playerID, "item_id": input.ItemID}
	update := bson.M{
		"$set":         bson.M{"name": input.Name, "updated_at": now},
		"$inc":         bson.M{"quantity": input.Quantity},
		"$setOnInsert": bson.M{"player_id": playerID, "item_id": input.ItemID},
	}
	var doc struct {
		ItemID    string    `bson:"item_id"`
		Name      string    `bson:"name"`
		Quantity  int       `bson:"quantity"`
		UpdatedAt time.Time `bson:"updated_at"`
	}
	err := r.collection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&doc)
	if err != nil {
		return Item{}, err
	}
	return Item{ItemID: doc.ItemID, Name: doc.Name, Quantity: doc.Quantity, UpdatedAt: doc.UpdatedAt}, nil
}

func (r *MongoRepository) AddReward(ctx context.Context, playerID, rewardID string, input AddInput) (Item, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"player_id":          playerID,
		"item_id":            input.ItemID,
		"applied_reward_ids": bson.M{"$ne": rewardID},
	}
	update := bson.M{
		"$set":         bson.M{"name": input.Name, "updated_at": now},
		"$inc":         bson.M{"quantity": input.Quantity},
		"$addToSet":    bson.M{"applied_reward_ids": rewardID},
		"$setOnInsert": bson.M{"player_id": playerID, "item_id": input.ItemID},
	}
	var doc struct {
		ItemID    string    `bson:"item_id"`
		Name      string    `bson:"name"`
		Quantity  int       `bson:"quantity"`
		UpdatedAt time.Time `bson:"updated_at"`
	}
	err := r.collection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&doc)
	if mongo.IsDuplicateKeyError(err) {
		err = r.collection.FindOne(ctx, bson.M{"player_id": playerID, "item_id": input.ItemID}).Decode(&doc)
	}
	if err != nil {
		return Item{}, err
	}
	return Item{ItemID: doc.ItemID, Name: doc.Name, Quantity: doc.Quantity, UpdatedAt: doc.UpdatedAt}, nil
}

func (r *MongoRepository) Remove(ctx context.Context, playerID, itemID string, quantity int) error {
	filter := bson.M{"player_id": playerID, "item_id": itemID, "quantity": bson.M{"$gte": quantity}}
	result, err := r.collection.UpdateOne(ctx, filter, bson.M{
		"$inc": bson.M{"quantity": -quantity},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domainerr.ErrNotFound
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"player_id": playerID, "item_id": itemID, "quantity": 0})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	return err
}
