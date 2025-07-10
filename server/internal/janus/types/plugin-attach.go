package types

type JanusAttachRequest struct {
	Janus       string `json:"janus"`       // e.g., "attach"
	SessionID   int64  `json:"session_id"`  // The session identifier (assuming it's an integer)
	Plugin      string `json:"plugin"`      // The plugin's unique package name
	Transaction string `json:"transaction"` // Random string for transaction tracking
}
