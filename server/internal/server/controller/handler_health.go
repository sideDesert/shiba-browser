package controller

import (
	"fmt"
	"net/http"
	"sideDesert/shiba/internal/server/lib"
)

func (c *Controller) handleHealth(w http.ResponseWriter, r *http.Request) error {
	str, _ := c.s.Health()
	fmt.Println("Health Check!")
	return lib.WriteJSON(w, r, http.StatusOK, struct {
		Msg string `json:"msg"`
	}{Msg: str})
}
