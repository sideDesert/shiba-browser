package controller

import (
	"net/http"
	"sideDesert/shiba/internal/server/lib"
)

func (c *Controller) handleUser(w http.ResponseWriter, r *http.Request) error {
	userResponse, err := c.s.GetUserResponse(w, r)
	if err != nil {
		return err
	}

	return lib.WriteJSON(w, r, http.StatusAccepted, userResponse)
}
