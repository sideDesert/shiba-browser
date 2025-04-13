package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	server "sideDesert/shiba/internal/server/lib"

	"sideDesert/shiba/internal/server/dto"
)

func (c *Controller) handleOAuthCallback(w http.ResponseWriter, r *http.Request) error {
	// TODO: Check CSRF Token Validity
	log.Println("handleOAuthCallback[URL]:", r.URL)

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	stateCookie, err := r.Cookie("shiba-state-token")

	if err != nil || state != stateCookie.Value {
		log.Println("Error in handleOAuthCallback[Cookie]:", err)
		return err
	}

	if code == "" {
		err := fmt.Errorf("No code in URL, code recieved is \"\"")
		log.Println("Error in handleOAuthCallback[code]:", err)
		return err
	}

	// For Debugging
	fmt.Println("Code for token", code)

	token, err := c.s.ExchangeCodeForToken(code)
	if err != nil {
		log.Println("Error in handleOAuthCallback[ExchangeCodeForToken]:", err)
		return err
	}

	// Used for logs
	userData, err := c.s.GetOAuthUserData(token)

	if err != nil {
		log.Println("Error in handleOAuthCallback[GetOAuthUserData]:", err)
		return err
	}

	// Check if Email is empty
	user, isNewUser, err := c.s.Store.GetOrCreateUser(c.s.Ctx, userData)
	clientUrl := os.Getenv("CLIENT_URL")

	if err != nil {
		log.Println("Error in handleOAuthCallback[GetOrCreateUserByEmail]:", err)
		return err
	}

	jwtToken, err := server.CreateToken(user.UserId)

	if err != nil {
		log.Println("Error in handleOAuthCallback[createToken]:", err)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "shiba-auth-token",
		Value:    jwtToken,
		Expires:  time.Now().Add(5 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	if isNewUser {
		http.Redirect(w, r, clientUrl+"/signup", http.StatusAccepted)
	}

	http.Redirect(w, r, clientUrl+"/dashboard", http.StatusFound)

	return nil
}

func (c *Controller) handleSignup(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("Method not allowed")
	}

	userReq := dto.SignupUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		log.Println("Error in controller.handleSignup[Decode()]:", err)
		return err
	}

	userDetails, err := c.s.RegisterUser(&userReq)

	if err != nil {
		log.Println("Error in controller.handleSignup[RegisterUser()]", err)
		return err
	}

	log.Printf("Registered New User: %s %s", userReq.Name, userReq.Email)
	return server.WriteJSON(w, r, http.StatusOK, userDetails)
}

func (c *Controller) handleLogout(w http.ResponseWriter, r *http.Request) error {
	return nil
}
