package main

import (
	"fmt"
	"net"
	"strings"

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

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	logger.Info("Starting PUBG Leaderboard service")

	logger.Infof("Redis Cluster Service: %s", cfg.RedisAddr)

	// Extract the hostname without the port
	hostname := strings.Split(cfg.RedisAddr, ":")[0]

	// Perform DNS lookup
	addresses, err := net.LookupHost(hostname)
	if err != nil {
		logger.Errorf("DNS Lookup error for Redis address '%s': %v", hostname, err)
	} else {
		for i, addr := range addresses {
			addresses[i] = fmt.Sprintf("%s:6379", addr)
			logger.Infof("Resolved Redis address to IP: %s", addr)
		}
	}

	redisClient, err := store.NewRedisClient(addresses, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		logger.Fatalf("Error initializing Redis cluster client: %v", err)
	}

	minioClient, err := store.NewMinioClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, false)
	if err != nil {
		logger.Fatalf("Error initializing Minio client: %v", err)
	}

	// Initialize the service layer with Redis client and Resty client
	restyClient := client.NewPUBGClient(cfg, logger) // Assuming you have a Resty client setup for PUBG API
	leaderboardService := service.NewLeaderboardService(redisClient, restyClient, minioClient, logger)

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
