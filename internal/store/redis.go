package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gbasileGP/pubg-leaderboard/internal/model"
	"github.com/go-redis/redis/v8"
)

var ErrCacheMiss = errors.New("data not found in Redis")

type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new Redis Cluster client and checks the connection.
func NewRedisClient(addrs []string, password string, db int) (*RedisClient, error) {
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,    // List of cluster node addresses
		Password: password, // Password for Redis cluster authentication
	})

	// Check the connection by sending a PING command to one of the cluster nodes
	_, err := clusterClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	// You can return your RedisClient wrapping the cluster client instead of a regular client
	return &RedisClient{Client: clusterClient}, nil
}

// Ping tests connectivity to the Redis server.
func (rc *RedisClient) Ping(ctx context.Context) error {
	return rc.Client.Ping(ctx).Err()
}

// GetLeaderboard retrieves the leaderboard data from Redis.
func (rc *RedisClient) GetLeaderboard(ctx context.Context) (*model.LeaderboardResponse, error) {
	data, err := rc.Client.Get(ctx, "leaderboard").Result()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	} else if err != nil {
		return nil, err
	}

	leaderboard := &model.LeaderboardResponse{}
	err = json.Unmarshal([]byte(data), leaderboard)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling leaderboard data: %v", err)
	}

	return leaderboard, nil
}

// UpdateLeaderboard updates and structures the leaderboard data in Redis.
func (rc *RedisClient) UpdateLeaderboard(ctx context.Context, leaderboardData *model.LeaderboardResponse) error {
	// Begin a new Redis transaction.
	pipe := rc.Client.TxPipeline()

	// Store each player's stats in a separate hash.
	for _, player := range leaderboardData.Included {
		playerStatsJSON, err := json.Marshal(player.Attributes)
		if err != nil {
			return fmt.Errorf("error marshaling player stats: %v", err)
		}

		// Set the player stats hash.
		pipe.HSet(ctx, player.ID, "stats", playerStatsJSON)
	}

	// Optionally set an expiration time on each hash.
	// This is not necessary if your leaderboard doesn't expire or if you handle expiration differently.
	for _, player := range leaderboardData.Included {
		pipe.Expire(ctx, player.ID, 10*time.Minute)
	}

	// Execute the transaction.
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error updating leaderboard in Redis: %v", err)
	}

	return nil
}

// GetSeason retrieves the current season identifier from Redis.
func (rc *RedisClient) GetSeason(ctx context.Context) (*model.SeasonData, error) {
	data, err := rc.Client.Get(ctx, "current_season").Result()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	} else if err != nil {
		return nil, err
	}

	season := &model.SeasonData{}
	err = json.Unmarshal([]byte(data), season)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling season data: %v", err)
	}

	return season, nil
}

// UpdateSeason updates the current season data in Redis.
func (rc *RedisClient) UpdateSeason(ctx context.Context, season *model.SeasonData) error {
	data, err := json.Marshal(season)
	if err != nil {
		return fmt.Errorf("error marshalling season data: %v", err)
	}

	// This sets the season data with a 24-hour expiry, matching the daily season refresh requirement.
	return rc.Client.Set(ctx, "current_season", data, 24*time.Hour).Err()
}
