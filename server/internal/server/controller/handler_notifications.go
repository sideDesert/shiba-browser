package controller

import (
	"fmt"
	"log"
	"net/http"
	"sideDesert/shiba/internal/server/lib"
)

func (c *Controller) handleNotifications(w http.ResponseWriter, r *http.Request) error {
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return fmt.Errorf("Invalid user ID")
	}

	if r.Method != http.MethodGet {
		return fmt.Errorf("Method not allowed: %s", r.Method)
	}

	friends, err := c.s.GetFriendRequests(userId)
	if err != nil {
		log.Println("Error in handleNotifications:", err)
		return err
	}

	return lib.WriteJSON(w, r, http.StatusOK, friends)
}
