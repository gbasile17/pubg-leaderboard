package model

// SeasonsResponse represents the response structure for the list of seasons from the PUBG API.
type SeasonsResponse struct {
	Data  []SeasonData `json:"data"`
	Links Links        `json:"links"`
	Meta  interface{}  `json:"meta"` // Meta can be an empty interface as it's often empty or varies
}

// SeasonData holds individual season data.
type SeasonData struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Attributes SeasonAttribute `json:"attributes"`
}

// SeasonAttribute contains attributes of a season.
type SeasonAttribute struct {
	IsCurrentSeason bool `json:"isCurrentSeason"`
	IsOffseason     bool `json:"isOffseason"`
}

// Links holds link information in the API response.
type Links struct {
	Self string `json:"self"`
}
