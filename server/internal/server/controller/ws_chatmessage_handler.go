package controller

import (
	"sideDesert/shiba/internal/logger"
	"sideDesert/shiba/internal/server/lib"
	"strings"
)

func (c *Controller) handleSocketChatMessage(msg lib.ChatMessage) {
	s := strings.Split(msg.Subject, ".")
	log := logger.NewLogger(logger.Console, "handleSocketChatMessage")

	if len(s) < 2 {
		log.Error("Subject", "Error chat message is not correct format, got:", msg.Subject)
		return
	}

	chatroomId := s[2]
	c.msgChannel.Publish(&msg)

	if err := c.s.StoreChatMessage(msg.Sender, chatroomId, msg.Payload); err != nil {
		log.Error("Error storing chat message:", err)
		return
	}
}
