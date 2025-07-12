package configs

import (
	"os"
	"time"
)

type Configs struct {
	AuctionInterval   time.Duration
	CheckOpenAuctions bool
}

func GetConfigs() *Configs {
	return &Configs{
		AuctionInterval: getAuctionInterval(),
	}
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}
