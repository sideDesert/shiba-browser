package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sideDesert/shiba/internal/server/dto"
	server "sideDesert/shiba/internal/server/lib"
)

func (c *Controller) handleChatRoom(w http.ResponseWriter, r *http.Request) error {
	// This is for GET Requeests - We Get all chatrooms
	if r.Method == http.MethodGet {

		userId := r.Context().Value("userId").(string)
		if userId == "" {
			log.Println("Error in handleCreateChatRoom[userId] No userId")
			return fmt.Errorf("No User ID in request context")
		}

		cs, err := c.s.GetUserChatRooms(userId)
		if err != nil {
			log.Println("Error in handleCreateChatRoom[GetUserChatRooms]", err)
			return fmt.Errorf("Could not get User Chat Rooms")
		}
		return server.WriteJSON(w, r, http.StatusOK, struct {
			Chatrooms []server.Chatroom `json:"chatrooms"`
		}{
			Chatrooms: cs,
		})
	}

	// This is for POST Requests - We Create
	if r.Method == http.MethodPost {
		createChatRequest := dto.CreateChatRoomRequest{}
		err := json.NewDecoder(r.Body).Decode(&createChatRequest)
		if err != nil {
			log.Println("Error in handleCreateChatRoom[decodeChatRequest]", err)
			return err
		}

		chatRoomId, err := c.s.CreateChatRoom(createChatRequest)

		if err != nil {
			log.Println("Error in handleCreateChatRoom[createChatRoom]", err)
			return err
		}

		return server.WriteJSON(w, r, http.StatusOK, dto.CreateChatRoomResponse{
			ChatRoomId: chatRoomId,
		})
	}

	return fmt.Errorf("Error: Method not allowed %s", r.Method)
}
