package controller

import (
	"net/http"
	"sideDesert/shiba/internal/server/lib"
	"time"
)

func (c *Controller) handleLogin(w http.ResponseWriter, r *http.Request) error {
	oauthUrl := c.s.GenerateOAuthLoginURL("google")
	csrfToken := lib.CreateCSRFToken()

	http.SetCookie(w, &http.Cookie{
		Name: "shiba-state-token",
		// TODO: Change this
		Value:    csrfToken,
		Expires:  time.Now().Add(10 * time.Hour),
		HttpOnly: true,  // Prevent JavaScript access
		Secure:   false, // Only send over HTTPS
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, oauthUrl, http.StatusFound)
	return nil
}
