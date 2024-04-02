package config

import (
	"os"
	"strconv"
)

// Config represents the configuration settings for the application.
type Config struct {
	AppPort         string // Port on which the app will run
	RedisAddr       string // Redis server address
	RedisPass       string // Redis password
	RedisDB         int    // Redis database number
	PubgAPIEndpoint string // PUBG API endpoint
	PubgAPIKey      string // PUBG API key
	MinioEndpoint   string // MinIO server endpoint
	MinioAccessKey  string // MinIO access key
	MinioSecretKey  string // MinIO secret key
	BackupsBucket   string // MinIO bucket for backups
	BackupsFile     string // Backup file name
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() (*Config, error) {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, err
	}

	return &Config{
		AppPort:         getEnv("APP_PORT", "8080"),
		RedisAddr:       getEnv("REDIS_ADDR", "redis-cluster:6379"),
		RedisPass:       getEnv("REDIS_PASS", "themagicword"),
		RedisDB:         redisDB,
		PubgAPIEndpoint: getEnv("PUBG_API_ENDPOINT", "https://api.pubg.com/shards/pc-na"),
		PubgAPIKey:      getEnv("PUBG_API_KEY", ""),
		MinioEndpoint:   getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinioAccessKey:  getEnv("MINIO_ACCESS", "minio"),
		MinioSecretKey:  getEnv("MINIO_SECRET", "minio123"),
		BackupsBucket:   getEnv("BACKUPS_BUCKET", "pubg-leaderboard"),
		BackupsFile:     getEnv("BACKUPS_FILE", "leaderboard-backup.json"),
	}, nil
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
