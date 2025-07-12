package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/configs"
	"fullcycle-auction_go/helpers/testutils"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCloseAuction(t *testing.T) {
	t.Run("should return error if auction is not active", func(t *testing.T) {
		ctx := context.Background()
		testutils.WithDB(ctx, func(ctx context.Context, db *mongo.Database) {
			if db == nil {
				t.Fatal("Database is nil")
			}
			auctionRepository := NewAuctionRepository(db, &configs.Configs{
				AuctionInterval: 10 * time.Second,
			})
			id := "1"
			auctionEntityMongo := &AuctionEntityMongo{
				Id:          id,
				ProductName: "Product 1",
				Category:    "Category 1",
				Description: "Description 1",
				Condition:   auction_entity.New,
				Status:      auction_entity.Completed,
				Timestamp:   time.Now().Unix(),
			}
			_, err := db.Collection("auctions").InsertOne(ctx, auctionEntityMongo)
			assert.Nil(t, err)

			auction, err := auctionRepository.FindAuctionById(ctx, id)

			err = auctionRepository.closeAuction(ctx, auction)
			assert.NotNil(t, err)
			assert.Equal(t, err, internal_error.NewInternalServerError("Auction already closed"))

			db.Collection("auctions").DeleteOne(ctx, bson.M{"_id": id})
		})
	})
	t.Run("should close an auction", func(t *testing.T) {
		ctx := context.Background()
		testutils.WithDB(ctx, func(ctx context.Context, db *mongo.Database) {
			if db == nil {
				t.Fatal("Database is nil")
			}
			auctionRepository := NewAuctionRepository(db, &configs.Configs{
				AuctionInterval: 10 * time.Second,
			})
			id := "1"
			auctionEntityMongo := &AuctionEntityMongo{
				Id:          id,
				ProductName: "Product 1",
				Category:    "Category 1",
				Description: "Description 1",
				Condition:   auction_entity.New,
				Status:      auction_entity.Active,
				Timestamp:   time.Now().Unix(),
			}
			_, err := db.Collection("auctions").InsertOne(ctx, auctionEntityMongo)
			assert.Nil(t, err)

			auction, err := auctionRepository.FindAuctionById(ctx, id)
			assert.Nil(t, err)
			assert.Equal(t, auction_entity.Active, auction.Status)

			err = auctionRepository.closeAuction(ctx, auction)
			assert.Nil(t, err)

			updatedAuction, err := auctionRepository.FindAuctionById(ctx, id)
			assert.Nil(t, err)
			assert.Equal(t, auction_entity.Completed, updatedAuction.Status)

			db.Collection("auctions").DeleteOne(ctx, bson.M{"_id": id})
		})
	})
}

func TestCheckAndCloseAuctions(t *testing.T) {
	t.Run("should check and close auctions", func(t *testing.T) {
		ctx := context.Background()
		testutils.WithDB(ctx, func(ctx context.Context, db *mongo.Database) {
			if db == nil {
				t.Fatal("Database is nil")
			}
			for i := 0; i < 10; i++ {
				auctionEntityMongo := &AuctionEntityMongo{
					ProductName: "Product 1",
					Category:    "Category 1",
					Description: "Description 1",
					Condition:   auction_entity.New,
					Status:      auction_entity.Active,
					Timestamp:   time.Now().Unix(),
				}
				auctionEntityMongo.Id = fmt.Sprintf("%d", i)
				_, err := db.Collection("auctions").InsertOne(ctx, auctionEntityMongo)
				assert.Nil(t, err)
			}

			interval := 5 * time.Second
			auctionRepository := NewAuctionRepository(db, &configs.Configs{
				AuctionInterval: interval,
			})
			auctionRepository.checkOpenAuctions(ctx)

			auctions, err := auctionRepository.FindAuctions(ctx, nil, "", "")
			assert.Nil(t, err)
			assert.Equal(t, 10, len(auctions))
			for _, auction := range auctions {
				assert.Equal(t, auction_entity.Active, auction.Status)
			}

			time.Sleep(interval + 5*time.Second)

			updatedAuctions, err := auctionRepository.FindAuctions(ctx, nil, "", "")
			assert.Nil(t, err)
			assert.Equal(t, 10, len(updatedAuctions))
			for _, auction := range updatedAuctions {
				assert.Equal(t, auction_entity.Completed, auction.Status)
			}
			db.Collection("auctions").DeleteMany(ctx, bson.M{
				"_id": bson.M{
					"$in": bson.A{
						"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
					},
				},
			})
		})
	})
}
