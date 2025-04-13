package controller

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"sideDesert/shiba/internal/server/dto"
	server "sideDesert/shiba/internal/server/lib"
)

func (c *Controller) handleChatHistory(w http.ResponseWriter, r *http.Request) error {
	log.Println("Fetching Chat History...")
	if r.Method != http.MethodGet {
		return fmt.Errorf("Method Not Allowed, Method: %s", r.Method)
	}

	query := r.URL.Query()
	sender := query.Get("sender")
	if sender == "" {
		return fmt.Errorf("Please Include sender - s")
	}

	chatroomId := query.Get("cid")
	if chatroomId == "" {
		return fmt.Errorf("Please Include chatroomId - cid")
	}

	offset := query.Get("page")
	if chatroomId == "" {
		return fmt.Errorf("Please Include offset - page")
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		log.Println("Error in handleChatHistory:", err)
		return fmt.Errorf("page is not valid number")
	}

	// DEBUGGING
	log.Println("sender", sender, "chatroomId", chatroomId, "offset", offset)

	history := dto.ChatHistoryRequest{
		Sender:     sender,
		ChatroomId: chatroomId,
		Offset:     offsetInt,
	}

	chat, err := c.s.GetChatroomHistory(history.ChatroomId, history.Offset)

	if err != nil {
		log.Println("Error in handleChatHistory", err)
		return err
	}

	return server.WriteJSON(w, r, http.StatusOK, chat)
}
