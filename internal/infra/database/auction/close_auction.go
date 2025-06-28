package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ExpiringAuctions struct {
	mu       sync.Mutex
	auctions map[string]auction_entity.Auction
	interval time.Duration
}

func (ar *AuctionRepository) handleCloseAuction(ctx context.Context, auction auction_entity.Auction) {
	ar.expiringAuctions.mu.Lock()
	defer ar.expiringAuctions.mu.Unlock()
	if _, found := ar.expiringAuctions.auctions[auction.Id]; found {
		return
	}
	ar.expiringAuctions.auctions[auction.Id] = auction

	ctx, _ = context.WithCancel(ctx)
	go func(bgCtx context.Context) {
		endTime := auction.Timestamp.Add(ar.expiringAuctions.interval)
		sleep := time.Until(endTime)
		select {
		case <-time.After(sleep):
			err := ar.closeAuction(bgCtx, &auction)
			if err != nil {
				logger.Error("Error closing auction:", err)
			}
			ar.removeAuction(bgCtx, auction)
		case <-bgCtx.Done():
			logger.Info("Auction expiration goroutine canceled", zap.String("auction_id", auction.Id))
			return
		}
	}(ctx)
}

func (ar *AuctionRepository) removeAuction(_ context.Context, auction auction_entity.Auction) {
	ar.expiringAuctions.mu.Lock()
	defer ar.expiringAuctions.mu.Unlock()
	delete(ar.expiringAuctions.auctions, auction.Id)
}

func (ar *AuctionRepository) checkOpenAuctions(ctx context.Context) {
	statusFilter := auction_entity.Active
	activeAuctions, err := ar.FindAuctions(ctx, &statusFilter, "", "")
	if err != nil {
		logger.Error("Error trying to find active auctions", err)
		return
	}

	for _, auction := range activeAuctions {
		ar.handleCloseAuction(ctx, auction)
	}
}

func (ar *AuctionRepository) closeAuction(ctx context.Context, auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	foundAuction, err := ar.FindAuctionById(ctx, auctionEntity.Id)
	if err != nil {
		return err
	}

	if foundAuction.Status != auction_entity.Active {
		err = internal_error.NewInternalServerError("Auction already closed")
		logger.Error(fmt.Sprintf("Error trying to find close auction = %s", auctionEntity.Id), err)
		return err
	}

	foundAuction.Status = auction_entity.Completed
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          foundAuction.Id,
		ProductName: foundAuction.ProductName,
		Category:    foundAuction.Category,
		Description: foundAuction.Description,
		Condition:   foundAuction.Condition,
		Status:      foundAuction.Status,
		Timestamp:   foundAuction.Timestamp.Unix(),
	}
	_, updateErr := ar.Collection.UpdateByID(ctx, foundAuction.Id, auctionEntityMongo)
	if updateErr != nil {
		logger.Error(fmt.Sprintf("Error trying to update auction = %s", auctionEntity.Id), updateErr)
		return internal_error.NewInternalServerError("Error trying to update auction")
	}

	return nil
}
