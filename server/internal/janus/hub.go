package janus

import (
	"encoding/json"
	"sideDesert/shiba/internal/janus/types"
	"sideDesert/shiba/internal/logger"
	"sideDesert/shiba/lib"
)

type Hub struct {
	session map[string]*Session
}

func NewHub() *Hub {
	return &Hub{
		session: make(map[string]*Session),
	}
}

type SessionConfig struct {
	userId         string
	chatroomId     string
	messageHandler func(chan []byte)
}

func (h *Hub) NewSession(config *SessionConfig) {
	log := logger.NewLogger(logger.Console, "hub.NewSession")
	log.Info("Creating new session for user %s", config.userId)

	socket := NewSocket()
	requestTransactionId := lib.RandomString(7)
	message := map[string]any{
		"janus":       "create",
		"transaction": requestTransactionId,
	}
	data, err := json.Marshal(message)
	if err != nil {
		log.Error("json.Marshal", err)
		return
	}
	socket.Send(data)

	msg := <-socket.ReadChannel
	var response types.JanusSessionResponse
	err = json.Unmarshal(msg, &response)
	if err != nil {
		log.Error("json.Unmarshal", err)
		return
	}
	var sessionId int64
	if response.Transaction == requestTransactionId {
		sessionId = response.Data.ID
	} else {
		log.Error("Txn Id", "Got - %d, expected - %d", response.Transaction, requestTransactionId)
		return
	}
	session := &Session{
		userId:       config.userId,
		chatroomId:   config.chatroomId,
		janusConn:    socket,
		sessionId:    sessionId,
		handlerRefId: make(map[string]int64),
	}
	go config.messageHandler(socket.ReadChannel)

	h.session[config.userId] = session
}

func (h *Hub) CloseSession(userId string) {
	session, ok := h.session[userId]
	if ok {
		session.janusConn.Close()
		delete(h.session, userId)
	}
}
