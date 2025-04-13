package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sideDesert/shiba/internal/server/dto"
	"sideDesert/shiba/internal/server/lib"
)

func (c *Controller) handleRemote(w http.ResponseWriter, r *http.Request) error {
	userId := r.Context().Value("userId").(string)
	chatroomId := r.URL.Query().Get("cid")

	if r.Method == http.MethodGet {
		if chatroomId == "" {
			return fmt.Errorf("Query Params Missing chatroom id")
		}
		remote, err := c.s.GetChatroomRemote(chatroomId)
		if err != nil {
			log.Println("Error in handleRemote[GET]:", err)
			return fmt.Errorf("Error in GET Remote")
		}
		return lib.WriteJSON(w, r, http.StatusOK, remote)
	}

	if r.Method == http.MethodPut {
		body := dto.ChangeChatroomRemoteRequest{}
		json.NewDecoder(r.Body).Decode(&body)

		isRemote := c.s.CheckUserIsRemoteForChatroom(userId, body.ChatroomId)
		if !isRemote {
			return fmt.Errorf("User is not remote")
		}
		err := c.s.ChangeChatroomRemote(body.ChatroomId, body.UserId)
		if err != nil {
			log.Println("Error in handleRemote[PUT]:", err)
			return fmt.Errorf("Could not change chatroom remote")
		}

		return lib.WriteJSON(w, r, http.StatusOK, dto.PatchOKResponse{Status: "Success"})
	}

	return fmt.Errorf("Method not allowed: %s", r.Method)
}
