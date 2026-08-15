package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Counter struct {
	ID  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

func getNextSequence(ctx context.Context, db *mongo.Database, name string) (int64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result Counter
	err := db.Collection("counters").FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection("categories").Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return err
	}

	_, err = db.Collection("products").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "sku", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "category_id", Value: 1}}},
	})
	return err
}
