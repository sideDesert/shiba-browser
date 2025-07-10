package types

type JanusServerInfo struct {
	Janus                 string            `json:"janus"`
	Transaction           string            `json:"transaction"`
	Name                  string            `json:"name"`
	Version               int               `json:"version"`
	VersionString         string            `json:"version_string"`
	Author                string            `json:"author"`
	CommitHash            string            `json:"commit-hash"`
	CompileTime           string            `json:"compile-time"`
	LogToStdout           bool              `json:"log-to-stdout"`
	LogToFile             bool              `json:"log-to-file"`
	DataChannels          bool              `json:"data_channels"`
	AcceptingNewSessions  bool              `json:"accepting-new-sessions"`
	SessionTimeout        int               `json:"session-timeout"`
	ReclaimSessionTimeout int               `json:"reclaim-session-timeout"`
	CandidatesTimeout     int               `json:"candidates-timeout"`
	ServerName            string            `json:"server-name"`
	LocalIP               string            `json:"local-ip"`
	IPv6                  bool              `json:"ipv6"`
	ICELite               bool              `json:"ice-lite"`
	ICETCP                bool              `json:"ice-tcp"`
	ICENomination         string            `json:"ice-nomination"`
	ICEConsentFreshness   bool              `json:"ice-consent-freshness"`
	ICEKeepaliveConncheck bool              `json:"ice-keepalive-conncheck"`
	HangupOnFailed        bool              `json:"hangup-on-failed"`
	FullTrickle           bool              `json:"full-trickle"`
	MDNSEnabled           bool              `json:"mdns-enabled"`
	MinNackQueue          int               `json:"min-nack-queue"`
	NackOptimizations     bool              `json:"nack-optimizations"`
	TwccPeriod            int               `json:"twcc-period"`
	DTLSMTU               int               `json:"dtls-mtu"`
	StaticEventLoops      int               `json:"static-event-loops"`
	APISecret             bool              `json:"api_secret"`
	AuthToken             bool              `json:"auth_token"`
	EventHandlers         bool              `json:"event_handlers"`
	OpaqueIDInAPI         bool              `json:"opaqueid_in_api"`
	Dependencies          Dependencies      `json:"dependencies"`
	Transports            map[string]Plugin `json:"transports"`
	Events                map[string]any    `json:"events"`
	Loggers               map[string]any    `json:"loggers"`
	Plugins               map[string]Plugin `json:"plugins"`
}

type Dependencies struct {
	GLib2   string `json:"glib2"`
	Jansson string `json:"jansson"`
	LibNice string `json:"libnice"`
	LibSRTP string `json:"libsrtp"`
	LibCurl string `json:"libcurl"`
	Crypto  string `json:"crypto"`
}

type Plugin struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Version     int    `json:"version"`
	VersionStr  string `json:"version_string"`
}
