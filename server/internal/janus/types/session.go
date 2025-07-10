package types

type JanusSessionResponse struct {
	Janus       string     `json:"janus"`
	Transaction string     `json:"transaction"`
	Data        *JanusData `json:"data,omitempty"` // pointer to make it optional
}

type JanusData struct {
	ID int64 `json:"id"`
}
