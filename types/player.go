package types

// Player Source server players
type Player struct {
	Index    int    `json:"index"`
	Score    int    `json:"score"`
	Name     string `json:"name"`
	SteamID  string `json:"steam_id"`
	Duration int    `json:"duration"`
}
