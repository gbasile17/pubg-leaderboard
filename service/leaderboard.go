package service

import (
	"context"
	"time"

	"github.com/gbasileGP/pubg-leaderboard/internal/client"
	"github.com/gbasileGP/pubg-leaderboard/internal/model"
	"github.com/gbasileGP/pubg-leaderboard/internal/store"
	"github.com/sirupsen/logrus"
)

// LeaderboardService contains methods to interact with leaderboard functionalities.
type LeaderboardService struct {
	redisClient *store.RedisClient
	pubgClient  *client.PUBGClient
	logger      *logrus.Logger
}

// NewLeaderboardService creates a new service for leaderboard operations.
func NewLeaderboardService(redisClient *store.RedisClient, pubgClient *client.PUBGClient, logger *logrus.Logger) *LeaderboardService {
	service := &LeaderboardService{
		redisClient: redisClient,
		pubgClient:  pubgClient,
		logger:      logger,
	}

	// Start the background refreshers within the service initialization
	go service.startLeaderboardRefresher()
	go service.startSeasonRefresher()

	return service
}

// startLeaderboardRefresher runs a loop that refreshes the leaderboard every 10 minutes.
func (ls *LeaderboardService) startLeaderboardRefresher() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		err := ls.RefreshLeaderboard(context.Background())
		if err != nil {
			ls.logger.WithError(err).Error("Error refreshing leaderboard")
		} else {
			ls.logger.Info("Leaderboard refreshed successfully")
		}
	}
}

// startSeasonRefresher runs a loop that refreshes the current season daily.
func (ls *LeaderboardService) startSeasonRefresher() {
	// Refresh immediately on startup, then use the ticker for subsequent refreshes
	err := ls.RefreshCurrentSeason(context.Background())
	if err != nil {
		ls.logger.WithError(err).Error("Error refreshing current season")
	} else {
		ls.logger.Info("Current season refreshed successfully")
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		err := ls.RefreshCurrentSeason(context.Background())
		if err != nil {
			ls.logger.WithError(err).Error("Error refreshing current season")
		} else {
			ls.logger.Info("Current season refreshed successfully")
		}
	}
}

// GetCurrentSeason retrieves the current season from Redis or the external API.
func (ls *LeaderboardService) GetCurrentSeason(ctx context.Context) (*model.SeasonData, error) {
	seasonData, err := ls.redisClient.GetSeason(ctx)
	if err == nil && seasonData != nil {
		ls.logger.Info("Retrieved current season from Redis")
		return seasonData, nil
	}

	currentSeason, err := ls.pubgClient.GetCurrentSeason()
	if err != nil {
		ls.logger.WithError(err).Error("Failed to fetch current season from PUBG API")
		return nil, err
	}

	err = ls.redisClient.UpdateSeason(ctx, currentSeason)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to update current season in Redis")
	} else {
		ls.logger.Info("Updated current season in Redis")
	}

	return currentSeason, nil
}

// GetCurrentLeaderboard retrieves the current leaderboard, either from Redis or the external API.
func (ls *LeaderboardService) GetCurrentLeaderboard(ctx context.Context) (*model.LeaderboardResponse, error) {
	season, err := ls.GetCurrentSeason(ctx)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to get current season")
		return nil, err
	}

	leaderboard, err := ls.redisClient.GetLeaderboard(ctx)
	if err == nil && leaderboard != nil {
		ls.logger.Info("Retrieved leaderboard from Redis")
		return leaderboard, nil
	}

	leaderboardResp, err := ls.pubgClient.GetSeasonStats(season.ID, "squad")
	if err != nil {
		ls.logger.WithError(err).Error("Failed to fetch leaderboard from PUBG API")
		return nil, err
	}

	err = ls.redisClient.UpdateLeaderboard(ctx, leaderboardResp)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to update leaderboard in Redis")
	} else {
		ls.logger.Info("Updated leaderboard in Redis")
	}

	return leaderboardResp, nil
}

// RefreshCurrentSeason refreshes the current season data and updates the cache.
func (ls *LeaderboardService) RefreshCurrentSeason(ctx context.Context) error {
	currentSeason, err := ls.pubgClient.GetCurrentSeason()
	if err != nil {
		ls.logger.WithError(err).Error("Failed to refresh current season from PUBG API")
		return err
	}

	err = ls.redisClient.UpdateSeason(ctx, currentSeason)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to update current season in Redis")
	} else {
		ls.logger.Info("Refreshed and updated current season in Redis")
	}

	return nil
}

// RefreshLeaderboard refreshes the leaderboard data and updates the cache.
func (ls *LeaderboardService) RefreshLeaderboard(ctx context.Context) error {
	season, err := ls.GetCurrentSeason(ctx)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to get current season for leaderboard refresh")
		return err
	}
	leaderboardResp, err := ls.pubgClient.GetSeasonStats(season.ID, "squad")
	if err != nil {
		ls.logger.WithError(err).Error("Failed to refresh leaderboard from PUBG API")
		return err
	}

	err = ls.redisClient.UpdateLeaderboard(ctx, leaderboardResp)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to update leaderboard in Redis")
	} else {
		ls.logger.Info("Refreshed and updated leaderboard in Redis")
	}

	return nil
}
