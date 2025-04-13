package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sideDesert/shiba/internal/server/dto"
	server "sideDesert/shiba/internal/server/lib"
)

func (c *Controller) handleFriends(w http.ResponseWriter, r *http.Request) error {
	userId := r.Context().Value("userId").(string)
	// GET Friends
	if r.Method == http.MethodGet {
		friends, err := c.s.GetFriends(userId)
		if err != nil {
			return err
		}

		return server.WriteJSON(w, r, http.StatusOK, friends)
	}

	if r.Method == http.MethodPost {
		body := dto.SendFriendRequest{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			log.Println("Error in handleFriendsPOST[Decode]:", err)
			return fmt.Errorf("Body Is not of correct format")
		}

		err = c.s.SendFriendRequest(userId, body.FriendId)

		if err != nil {
			log.Println("Error in handleFriendsPOST[AcceptFriendRequest]:", err)
			return fmt.Errorf("Friend Request Could not be sent")
		}

		return server.WriteJSON(w, r, http.StatusOK, dto.FriendResponse{Status: "Sent"})
	}

	if r.Method == http.MethodPatch {
		body := dto.FriendStatusRequest{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			log.Println("Error in handleFriendsPATCH[Decode]:", err)
			return fmt.Errorf("Body Is not of correct format")
		}
		err = c.s.HandleFriendRequest(userId, body.Id, body.Status)

		if err != nil {
			log.Println("Error in handleFriendsPATCH[AcceptFriendRequest]:", err)
			return fmt.Errorf("Friend Request Could not be accepted")
		}

		return server.WriteJSON(w, r, http.StatusOK, dto.FriendResponse{Status: "accepted"})
	}

	return nil
}
