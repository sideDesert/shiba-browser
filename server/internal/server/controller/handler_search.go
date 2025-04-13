package controller

import (
	"fmt"
	"log"
	"net/http"
	"sideDesert/shiba/internal/server/lib"
)

func (c *Controller) handleSearch(w http.ResponseWriter, r *http.Request) error {
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return fmt.Errorf("Invalid user ID")
	}

	query := r.URL.Query().Get("q")
	users, err := c.s.SearchUsers(userId, query, 20)
	if err != nil {
		log.Println("Error in handleSearch:", err)
		return err
	}

	return lib.WriteJSON(w, r, http.StatusOK, users)
}
