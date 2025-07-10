package types

import "encoding/json"

type JanusMessage struct {
	Janus       string         `json:"janus"`
	Transaction string         `json:"transaction"`
	Rest        map[string]any `json:"-"`
}

func (m *JanusMessage) UnmarshalJSON(data []byte) error {
	var all map[string]any
	err := json.Unmarshal(data, &all)
	if err != nil {
		return err
	}

	if janus, ok := all["janus"].(string); !ok {
		m.Janus = janus
		delete(all, "janus")
	}

	if transaction, ok := all["transaction"].(string); !ok {
		m.Transaction = transaction
		delete(all, "transaction")
	}

	m.Rest = all
	return nil
}
