package model

// LeaderboardResponse represents the structure of the leaderboard response from the PUBG API.
type LeaderboardResponse struct {
	Data     LeaderboardData `json:"data"`
	Links    Links           `json:"links"`
	Meta     interface{}     `json:"meta"` // Meta can be an empty interface as it's often empty or varies
	Included []PlayerData    `json:"included"`
}

// LeaderboardData holds the data part of the leaderboard response.
type LeaderboardData struct {
	Type          string               `json:"type"`
	ID            string               `json:"id"`
	Attributes    LeaderboardAttribute `json:"attributes"`
	Relationships Relationships        `json:"relationships"`
}

// LeaderboardAttribute contains attributes of leaderboard data.
type LeaderboardAttribute struct {
	ShardId  string `json:"shardId"`
	GameMode string `json:"gameMode"`
	SeasonId string `json:"seasonId"`
}

// Relationships holds the relationships data in the leaderboard response.
type Relationships struct {
	Players PlayerRelationship `json:"players"`
}

// PlayerRelationship represents the data structure for player relationships.
type PlayerRelationship struct {
	Data []PlayerDataReference `json:"data"`
}

// PlayerDataReference holds a reference to player data.
type PlayerDataReference struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// PlayerData represents player details included in the leaderboard data.
type PlayerData struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Attributes PlayerAttribute `json:"attributes"`
}

// PlayerAttribute contains attributes related to the player.
type PlayerAttribute struct {
	Name  string      `json:"name"`
	Rank  int         `json:"rank"`
	Stats PlayerStats `json:"stats"`
}

// PlayerStats holds statistics related to the player's performance.
type PlayerStats struct {
	RankPoints     float64 `json:"rankPoints"`
	Wins           int     `json:"wins"`
	Games          int     `json:"games"`
	WinRatio       float64 `json:"winRatio"`
	AverageDamage  float64 `json:"averageDamage"`
	Kills          int     `json:"kills"`
	KillDeathRatio float64 `json:"killDeathRatio"`
	Kda            float64 `json:"kda"`
	AverageRank    float64 `json:"averageRank"`
	Tier           string  `json:"tier"`
	SubTier        string  `json:"subTier"`
}
