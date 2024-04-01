package service

import (
	"context"
	"fmt"
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
	if err != nil {
		wrappedErr := fmt.Errorf("failed to retrieve current season from Redis: %w", err)
		ls.logger.WithError(wrappedErr).Error("GetCurrentSeason error")
		return nil, wrappedErr
	} else if seasonData != nil {
		ls.logger.Info("Retrieved current season from Redis")
		return seasonData, nil
	}

	// Similar error handling for the PUBG client
	currentSeason, err := ls.pubgClient.GetCurrentSeason()
	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch current season from PUBG API: %w", err)
		ls.logger.WithError(wrappedErr).Error("GetCurrentSeason error")
		return nil, wrappedErr
	}

	err = ls.redisClient.UpdateSeason(ctx, currentSeason)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to update current season in Redis: %w", err)
		ls.logger.WithError(wrappedErr).Error("GetCurrentSeason error")
		return currentSeason, wrappedErr
	}

	ls.logger.Info("Updated current season in Redis")
	return currentSeason, nil
}

// GetCurrentLeaderboard retrieves the current leaderboard, either from Redis or the external API.
func (ls *LeaderboardService) GetCurrentLeaderboard(ctx context.Context) (*model.LeaderboardResponse, error) {
	season, err := ls.GetCurrentSeason(ctx)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to get current season for leaderboard retrieval")
		return nil, fmt.Errorf("failed to get current season for leaderboard retrieval: %w", err)
	}

	leaderboard, err := ls.redisClient.GetLeaderboard(ctx)
	if err != nil && err != store.ErrCacheMiss {
		wrappedErr := fmt.Errorf("failed to retrieve leaderboard from Redis: %w", err)
		ls.logger.WithError(wrappedErr).Error("Failed to retrieve leaderboard from Redis")
		return nil, wrappedErr
	} else if leaderboard != nil {
		ls.logger.Info("Retrieved leaderboard from Redis")
		return leaderboard, nil
	}

	// If the leaderboard wasn't in Redis or there was a cache miss, fetch from PUBG API
	leaderboardResp, err := ls.pubgClient.GetSeasonStats(season.ID, "squad")
	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch leaderboard from PUBG API: %w", err)
		ls.logger.WithError(wrappedErr).Error("Failed to fetch leaderboard from PUBG API")
		return nil, wrappedErr
	}

	// Update the cache with the new leaderboard data
	err = ls.redisClient.UpdateLeaderboard(ctx, leaderboardResp)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to update leaderboard in Redis: %w", err)
		ls.logger.WithError(wrappedErr).Error("Failed to update leaderboard in Redis")
		// Continue returning the fetched leaderboard despite Redis update failure
	} else {
		ls.logger.Info("Updated leaderboard in Redis")
	}

	return leaderboardResp, nil
}

// RefreshCurrentSeason refreshes the current season data and updates the cache.
func (ls *LeaderboardService) RefreshCurrentSeason(ctx context.Context) error {
	currentSeason, err := ls.pubgClient.GetCurrentSeason()
	if err != nil {
		wrappedErr := fmt.Errorf("failed to refresh current season from PUBG API: %w", err)
		ls.logger.WithError(wrappedErr).Error("Failed to refresh current season from PUBG API")
		return wrappedErr
	}

	err = ls.redisClient.UpdateSeason(ctx, currentSeason)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to update current season in Redis: %w", err)
		ls.logger.WithError(wrappedErr).Error("Failed to update current season in Redis")
		return wrappedErr
	}

	ls.logger.Info("Successfully refreshed and updated current season in Redis")
	return nil
}

// RefreshLeaderboard refreshes the leaderboard data and updates the cache.
func (ls *LeaderboardService) RefreshLeaderboard(ctx context.Context) error {
	season, err := ls.GetCurrentSeason(ctx)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to get current season for leaderboard refresh")
		return fmt.Errorf("failed to get current season for leaderboard refresh: %w", err)
	}

	leaderboardResp, err := ls.pubgClient.GetSeasonStats(season.ID, "squad")
	if err != nil {
		wrappedErr := fmt.Errorf("failed to refresh leaderboard from PUBG API: %w", err)
		ls.logger.WithError(wrappedErr).Error("RefreshLeaderboard error")
		return wrappedErr
	}

	err = ls.redisClient.UpdateLeaderboard(ctx, leaderboardResp)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to update leaderboard in Redis: %w", err)
		ls.logger.WithError(wrappedErr).Error("RefreshLeaderboard error")
		return wrappedErr
	}

	ls.logger.Info("Refreshed and updated leaderboard in Redis")
	return nil
}
