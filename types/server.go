package types

type (
	SourceEngine uint8
	ServerType   uint8
	ServerOs     uint8
	AppId        int32
)

const (
	AppTheShip AppId = 2400

	GoldSource SourceEngine = iota + 1
	Source
)

const (
	ServerTypeUnknown      ServerType = iota
	ServerTypeDedicated               // 1
	ServerTypeNonDedicated            // 2
	ServerTypeSourceTV                // 3
)

const (
	ServerOsUnknown ServerOs = iota
	ServerOsLinux            // 1
	ServerOsWindows          // 2
	ServerOsMac              // 3
)

// Official versions of the A2S_INFO reply.
const (
	GoldSourceInfoResponseType = 0x6D
	SourceInfoResponseType     = 0x49
)

func (t ServerType) String() string {
	switch t {
	case ServerTypeDedicated:
		return "Dedicated"
	case ServerTypeNonDedicated:
		return "NonDedicated"
	case ServerTypeSourceTV:
		return "SourceTN"
	default:
		return "Unknown"
	}
}

func (t ServerOs) String() string {
	switch t {
	case ServerOsLinux:
		return "Linux"
	case ServerOsWindows:
		return "Windows"
	case ServerOsMac:
		return "Mac"
	default:
		return "Unknown"
	}
}

// Server Information returned by an A2S_INFO query. Most of this is returned as-is
// from the wire, except where otherwise noted.
type Server struct {
	Address    string       `json:"address"`
	Version    SourceEngine `json:"version"`
	Protocol   uint8        `json:"protocol"`
	Name       string       `json:"name"`
	Map        string       `json:"map"`
	Folder     string       `json:"folder"`
	Game       string       `json:"game"`
	Players    uint8        `json:"players"`
	MaxPlayers uint8        `json:"max_players"`
	Bots       uint8        `json:"bots"`
	Type       ServerType   `json:"type"`
	OS         ServerOs     `json:"os"`
	Visibility uint8        `json:"visibility"`
	Vac        uint8        `json:"vac"`
	Mod        *ModInfo     `json:"mod"`
	TheShip    *TheShip     `json:"ship"`
	SourceTV   *SourceTV    `json:"source_tv"`
	Extended   *Extended    `json:"extended"`
}

// ModInfo Optional mod information returned by S2A_INFO_GOLDSRC.
type ModInfo struct {
	Url         string `json:"url"`
	DownloadUrl string `json:"download_url"`
	Version     uint32 `json:"version"`
	Size        uint32 `json:"size"`
	Type        uint8  `json:"type"`
	Dll         uint8  `json:"dll"`
}

// TheShip Optional information returned by App_TheShip.
type TheShip struct {
	Mode      uint8 `json:"mode"`
	Witnesses uint8 `json:"witnesses"`
	Duration  uint8 `json:"duration"`
}

// SourceTV Optional information available with S2A_INFO_SOURCE.
type SourceTV struct {
	Port uint16
	Name string
}

// Extended Optional information available with S2A_INFO_SOURCE. This is a grab-bag
// of various optional bits. If some are not present they are left as 0.
// In the future this may change to distinguish from being present as 0.
type Extended struct {
	AppId               AppId
	GameVersion         string
	Port                uint16 // 0 if not present.
	SteamId             uint64 // 0 if not present.
	GameModeDescription string // "" if not present.
	GameId              uint64 // 0 if not present.
}

// Engine Attempt to guess the game engine version.
func (s *Server) Engine() SourceEngine {
	if s.Version == GoldSourceInfoResponseType || s.Extended == nil {
		return GoldSource
	}

	if uint32(s.Extended.AppId) < 80 {
		return GoldSource
	}

	return Source
}
