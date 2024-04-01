package client

import (
	"fmt"

	"github.com/gbasileGP/pubg-leaderboard/internal/config"
	"github.com/gbasileGP/pubg-leaderboard/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type PUBGClient struct {
	client *resty.Client
	config *config.Config
	logger *logrus.Logger
}

func NewPUBGClient(cfg *config.Config, logger *logrus.Logger) *PUBGClient {
	client := resty.New()

	if logger == nil {
		logger = logrus.New()
	}

	return &PUBGClient{
		client: client,
		config: cfg,
		logger: logger,
	}
}

// GetCurrentSeason fetches the current PUBG season with retry logic and logging.
func (p *PUBGClient) GetCurrentSeason() (*model.SeasonData, error) {
	p.logger.Info("pubgclient - Fetching current PUBG season")
	var seasonsResp model.SeasonsResponse
	_, err := p.client.R().
		SetHeader("Accept", "application/vnd.api+json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.config.PubgAPIKey)).
		SetResult(&seasonsResp).
		Get(fmt.Sprintf("%s/seasons", p.config.PubgAPIEndpoint))

	if err != nil {
		p.logger.WithError(err).Error("pubgclient - PUBGClient request failed")
		return nil, err
	}

	for _, season := range seasonsResp.Data {
		if season.Attributes.IsCurrentSeason && !season.Attributes.IsOffseason {
			return &season, nil
		}
	}

	return nil, fmt.Errorf("pubgclient - current season not found")
}

// GetSeasonStats fetches leaderboard stats for the given season and game mode with logging.
func (p *PUBGClient) GetSeasonStats(seasonID, gameMode string) (*model.LeaderboardResponse, error) {
	// Log the attempt to fetch season stats
	p.logger.WithFields(logrus.Fields{
		"seasonID": seasonID,
		"gameMode": gameMode,
	}).Info("pubgclient - Fetching season stats from PUBG API")

	var leaderboardResp model.LeaderboardResponse
	resp, err := p.client.R().
		SetHeader("Accept", "application/vnd.api+json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.config.PubgAPIKey)).
		SetResult(&leaderboardResp).
		Get(fmt.Sprintf("%s/leaderboards/%s/%s", p.config.PubgAPIEndpoint, seasonID, gameMode))

	// Log if there was an error in the request
	if err != nil {
		p.logger.WithError(err).WithFields(logrus.Fields{
			"seasonID": seasonID,
			"gameMode": gameMode,
		}).Error("pubgclient - Error fetching season stats")
		return nil, err
	}

	// Log successful retrieval with response status
	p.logger.WithFields(logrus.Fields{
		"seasonID": seasonID,
		"gameMode": gameMode,
		"status":   resp.Status(),
	}).Info("pubgclient - Successfully fetched season stats")

	return &leaderboardResp, nil
}
