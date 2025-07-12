package auction

import (
	"context"
	"fullcycle-auction_go/configuration/configs"
	"fullcycle-auction_go/helpers/testutils"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCreateAuction(t *testing.T) {
	t.Run("should create an auction", func(t *testing.T) {
		ctx := context.Background()
		testutils.WithDB(ctx, func(ctx context.Context, db *mongo.Database) {
			if db == nil {
				t.Fatal("Database is nil")
			}
			auctionRepository := NewAuctionRepository(db, &configs.Configs{
				AuctionInterval: 30 * time.Second,
			})
			id := "1"
			err := auctionRepository.CreateAuction(ctx, &auction_entity.Auction{
				Id:          id,
				ProductName: "Product 1",
				Category:    "Category 1",
				Description: "Description 1",
				Condition:   auction_entity.New,
				Status:      auction_entity.Active,
				Timestamp:   time.Now(),
			})
			assert.Nil(t, err)

			auction, err := auctionRepository.FindAuctionById(ctx, id)
			assert.Nil(t, err)
			assert.Equal(t, "Product 1", auction.ProductName)
			assert.Equal(t, "Category 1", auction.Category)
			assert.Equal(t, "Description 1", auction.Description)
			assert.Equal(t, auction_entity.New, auction.Condition)
			assert.Equal(t, auction_entity.Active, auction.Status)

			db.Collection("auctions").DeleteOne(ctx, bson.M{"_id": id})
		})
	})

	t.Run("should create an auction and close it after the interval", func(t *testing.T) {
		ctx := context.Background()
		testutils.WithDB(ctx, func(ctx context.Context, db *mongo.Database) {
			if db == nil {
				t.Fatal("Database is nil")
			}
			interval := 5 * time.Second
			auctionRepository := NewAuctionRepository(db, &configs.Configs{
				AuctionInterval: interval,
			})
			id := "1"
			err := auctionRepository.CreateAuction(ctx, &auction_entity.Auction{
				Id:          id,
				ProductName: "Product 1",
				Category:    "Category 1",
				Description: "Description 1",
				Condition:   auction_entity.New,
				Status:      auction_entity.Active,
				Timestamp:   time.Now(),
			})
			assert.Nil(t, err)

			auction, err := auctionRepository.FindAuctionById(ctx, id)
			assert.Nil(t, err)
			assert.Equal(t, "Product 1", auction.ProductName)
			assert.Equal(t, "Category 1", auction.Category)
			assert.Equal(t, "Description 1", auction.Description)
			assert.Equal(t, auction_entity.New, auction.Condition)
			assert.Equal(t, auction_entity.Active, auction.Status)

			// Wait for the auction to expire
			time.Sleep(interval + 5*time.Second)

			auction, err = auctionRepository.FindAuctionById(ctx, id)
			assert.Nil(t, err)
			assert.Equal(t, auction_entity.Completed, auction.Status)

			db.Collection("auctions").DeleteOne(ctx, bson.M{"_id": id})
		})
	})
}
