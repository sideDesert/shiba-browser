package janus

import (
	"encoding/json"
	"sideDesert/shiba/internal/janus/types"
	"sideDesert/shiba/internal/logger"
	"sideDesert/shiba/lib"
)

type Session struct {
	userId       string
	sessionId    int64
	chatroomId   string
	janusConn    *Socket
	handlerRefId map[string]int64
}

func (s *Session) AttachPlugin(pluginSignature string) {
	log := logger.NewLogger(logger.Console, "*Session.AttachPlugin")
	msg := types.JanusAttachRequest{
		Janus:       "attach",
		SessionID:   s.sessionId,
		Transaction: lib.RandomString(7),
		Plugin:      pluginSignature,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("json.Marshal", err)
		return
	}
	s.janusConn.Send(data)
}
