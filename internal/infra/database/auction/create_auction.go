package auction

import (
	"context"
	"fullcycle-auction_go/configuration/configs"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection       *mongo.Collection
	expiringAuctions ExpiringAuctions
}

func NewAuctionRepository(database *mongo.Database, cf *configs.Configs) *AuctionRepository {
	ar := &AuctionRepository{
		Collection: database.Collection("auctions"),
		expiringAuctions: ExpiringAuctions{
			interval: cf.AuctionInterval,
			auctions: make(map[string]auction_entity.Auction),
			mu:       sync.Mutex{},
		},
	}
	if cf.CheckOpenAuctions {
		go ar.checkOpenAuctions(context.Background())
	}
	return ar
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}
	ar.handleCloseAuction(ctx, *auctionEntity)

	return nil
}
