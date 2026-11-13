package db

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/fehmicorp/s3object/v1/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func Connect(ctx context.Context, c *types.Database) (*DB, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		url.QueryEscape(c.User),
		url.QueryEscape(c.Password),
		c.Host,
		c.Port,
	)
	dbName := c.Name
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	database := client.Database(dbName)
	db := &DB{
		Client:   client,
		Database: database,
	}
	if err := db.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) ensureIndexes(ctx context.Context) error {
	// Buckets collection: unique index on name
	_, err := db.Database.Collection("buckets").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	// Objects collection: unique compound index on bucket + object_key
	_, err = db.Database.Collection("objects").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "bucket", Value: 1},
			{Key: "object_key", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (db *DB) Close(ctx context.Context) error {
	return db.Client.Disconnect(ctx)
}

// CreateBucket inserts a new bucket record.
func (db *DB) CreateBucket(ctx context.Context, name string) error {
	collection := db.Database.Collection("buckets")
	_, err := collection.InsertOne(ctx, types.Bucket{
		Name:      name,
		CreatedAt: time.Now(),
	})
	return err
}

// UpsertObject saves or updates object metadata in MongoDB.
func (db *DB) UpsertObject(ctx context.Context, obj types.Object) error {
	collection := db.Database.Collection("objects")
	filter := bson.M{
		"bucket":     obj.Bucket,
		"object_key": obj.Key,
	}
	obj.UpdatedAt = time.Now()
	update := bson.M{"$set": obj}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// DeleteObject removes an object's metadata record.
func (db *DB) DeleteObject(ctx context.Context, bucket, key string) error {
	collection := db.Database.Collection("objects")
	_, err := collection.DeleteOne(ctx, bson.M{
		"bucket":     bucket,
		"object_key": key,
	})
	return err
}
