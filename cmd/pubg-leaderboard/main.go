package main

import (
	"github.com/gbasileGP/pubg-leaderboard/internal/api"
	"github.com/gbasileGP/pubg-leaderboard/internal/client"
	"github.com/gbasileGP/pubg-leaderboard/internal/config"
	"github.com/gbasileGP/pubg-leaderboard/internal/store"
	"github.com/gbasileGP/pubg-leaderboard/service"
	"github.com/sirupsen/logrus"
)

func main() {
	// Configure the logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting PUBG Leaderboard service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	redisClient, err := store.NewRedisClient([]string{cfg.RedisAddr}, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		logger.Fatalf("Error initializing Redis cluster client: %v", err)
	}

	// Initialize the service layer with Redis client and Resty client
	restyClient := client.NewPUBGClient(cfg, logger) // Assuming you have a Resty client setup for PUBG API
	leaderboardService := service.NewLeaderboardService(redisClient, restyClient, logger)

	// Initialize the server with the Redis client and logger
	server := api.NewServer(redisClient, leaderboardService, logger)
	if err != nil {
		logger.Fatalf("Error initializing server: %v", err)
	}

	// Start the server
	if err := server.Run(":8080"); err != nil {
		logger.Fatalf("Error running server: %v", err)
	}
}
