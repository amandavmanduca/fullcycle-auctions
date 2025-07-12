package testutils

import (
	"context"
	"fullcycle-auction_go/configuration/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewDB(ctx context.Context) (*mongo.Database, error) {
	client, err := mongo.Connect(
		ctx, options.Client().ApplyURI("mongodb://localhost:27017/test"))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("Error trying to ping mongodb database", err)
		return nil, err
	}

	return client.Database("test"), nil
}

func WithDB(ctx context.Context, f func(ctx context.Context, db *mongo.Database)) {
	db, err := NewDB(ctx)
	if err != nil {
		return
	}

	f(ctx, db)

	db.Client().Disconnect(ctx)
}
