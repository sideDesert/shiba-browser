package lib

import (
	"context"
	"log"
	"net/http"
	"os"
)

func AllowCors(next http.Handler) http.Handler {
	// log.Println("Allow Cors Enabled")
	clientUrl := os.Getenv("CLIENT_URL")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", clientUrl)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS") // Allowed HTTP methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")            // Allowed headers

		if r.Method == "OPTIONS" {
			// This is the key: reply directly with 200
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthenticateMiddleware(next http.Handler) http.Handler {
	// Retrieve the token from the cookie
	// log.Println("Authenticate Middleware Enabled")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenCoookie, err := r.Cookie("shiba-auth-token")
		tokenString := tokenCoookie.String()
		// fmt.Println("TOKEN STRING:", tokenString)

		tokenString, ok := extractTokenString(tokenString)
		if !ok {
			log.Println("Invalid token string format")
			http.Redirect(w, r, Client("/login"), http.StatusSeeOther)
			return
		}

		if err != nil {
			log.Println("Token missing in cookie")
			http.Redirect(w, r, Client("/login"), http.StatusUnauthorized)
			return
		}

		// Verify the token
		token, err := VerifyToken(tokenString)

		if err != nil {
			log.Printf("Token verification failed: %v\\n", err)
			http.Redirect(w, r, Client("/login"), http.StatusSeeOther)
			return
		}

		sub, err := token.Claims.GetSubject()
		if err != nil {
			log.Println("userId not found in token")
			http.Redirect(w, r, Client("/login"), http.StatusUnauthorized)
			return
		}
		// log.Println("User:", sub)

		ctx := context.WithValue(r.Context(), "userId", sub)
		// Print information about the verified token
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
