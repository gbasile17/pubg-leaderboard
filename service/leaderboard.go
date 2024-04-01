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
	go service.startSeasonRefresher()
	go service.startLeaderboardRefresher()

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
		ls.logger.WithError(err).Error("Routine Refresher Error")
	} else {
		ls.logger.Info("Current season refreshed successfully")
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		err := ls.RefreshCurrentSeason(context.Background())
		if err != nil {
			ls.logger.WithError(err).Error("Routine Refresher Error")
		} else {
			ls.logger.Info("Current season refreshed successfully")
		}
	}
}

// GetCurrentSeason retrieves the current season from Redis or the external API.
func (ls *LeaderboardService) GetCurrentSeason(ctx context.Context) (*model.SeasonData, error) {
	seasonData, err := ls.redisClient.GetSeason(ctx)
	if err != nil {
		if err == store.ErrCacheMiss {
			// If data is not found in Redis, log the cache miss and continue to fetch from the PUBG API.
			ls.logger.Info("Current season not found in Redis, fetching from PUBG API")
		} else {
			// For actual errors, log and return the error.
			ls.logger.WithError(err).Error("Error retrieving current season from Redis")
			return nil, err
		}
	} else if seasonData != nil {
		// If data is successfully retrieved from Redis, return it.
		ls.logger.Info("Retrieved current season from Redis")
		return seasonData, nil
	}

	// Fetch the current season from the PUBG API if it's not in Redis or Redis had an actual error
	currentSeason, err := ls.pubgClient.GetCurrentSeason()
	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch current season from PUBG API: %w", err)
		ls.logger.WithError(wrappedErr).Error("Failed to fetch current season from PUBG API")
		return nil, wrappedErr
	}

	// Attempt to update the season in Redis after fetching from the PUBG API
	err = ls.redisClient.UpdateSeason(ctx, currentSeason)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to update current season in Redis: %w", err)
		ls.logger.WithError(wrappedErr).Warn("Failed to update current season in Redis, but season was fetched from PUBG API")
	} else {
		ls.logger.Info("Updated current season in Redis")
	}

	return currentSeason, nil
}

func (ls *LeaderboardService) GetCurrentLeaderboard(ctx context.Context) (*model.LeaderboardResponse, error) {
	season, err := ls.GetCurrentSeason(ctx)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to get current season for leaderboard retrieval")
		return nil, fmt.Errorf("failed to get current season for leaderboard retrieval: %w", err)
	}

	// Attempt to retrieve the leaderboard from Redis
	leaderboard, err := ls.redisClient.GetLeaderboard(ctx)
	if err != nil {
		if err == store.ErrCacheMiss {
			ls.logger.Info("Leaderboard cache miss in Redis, fetching from PUBG API")
		} else {
			// Log and return other Redis errors
			wrappedErr := fmt.Errorf("failed to retrieve leaderboard from Redis: %w", err)
			ls.logger.WithError(wrappedErr).Error("Failed to retrieve leaderboard from Redis")
			return nil, wrappedErr
		}
	} else {
		// Successfully retrieved leaderboard from Redis
		ls.logger.Info("Retrieved leaderboard from Redis")
		return leaderboard, nil
	}

	// Fetch from the PUBG API as either there was a cache miss or another Redis error
	leaderboardResp, err := ls.pubgClient.GetSeasonStats(season.ID, "squad")
	if err != nil {
		wrappedErr := fmt.Errorf("failed to fetch leaderboard from PUBG API: %w", err)
		ls.logger.WithError(wrappedErr).Error("Failed to fetch leaderboard from PUBG API")
		return nil, wrappedErr
	}

	// Update the cache with the new leaderboard data after successful fetch from PUBG API
	err = ls.redisClient.UpdateLeaderboard(ctx, leaderboardResp)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to update leaderboard in Redis: %w", err)
		ls.logger.WithError(wrappedErr).Warn("Failed to update leaderboard in Redis, but returning latest data from PUBG API")
	} else {
		ls.logger.Info("Updated leaderboard in Redis with the latest data from PUBG API")
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
