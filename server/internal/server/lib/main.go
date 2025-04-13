package lib

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Init() {
	mrand.Seed(time.Now().UnixNano())
}

func Contains[T comparable](slice []T, item T) bool {
	return slices.Contains(slice, item)
}

func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(v)
}

func CreateCSRFToken() string {
	return "state"
}

func CreateHTTPHandleFunc(f func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			v := ApiError{Error: err.Error()}
			WriteJSON(w, r, http.StatusBadRequest, v)
		}
	}
}

func GenerateSSRC() uint32 {
	return mrand.Uint32()
}

func GenerateSecureRandomID(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func HashPassword(pass string) (string, error) {
	bp := []byte(pass)
	bytes, err := bcrypt.GenerateFromPassword(bp, 10)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CreateToken(userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"iss": "shiba",
		"aud": "user",
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	// Parse the token with the secret key
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	// Check for verification errors
	if err != nil {
		fmt.Println("Error in verifyToken:", err)
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Return the verified token
	return token, nil
}

func extractTokenString(tokenString string) (string, bool) {
	parts := strings.Split(tokenString, "=")
	if len(parts) >= 2 {
		return parts[1], true
	}

	return "", false
}

func Client(path string) string {
	return os.Getenv("CLIENT_URL") + path
}

func StartXvfb() (*exec.Cmd, string, error) {
	display := ":99"
	fmt.Println("Starting Virtual Display...")

	cmd := exec.Command("Xvfb", display, "-screen", "0", "1920x1080x24")

	// Start Xvfb
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting virtual display: %s\n", err)
		return nil, "", fmt.Errorf("failed to start Xvfb:  %w", err)
	}

	fmt.Println("Xvfb started on Display", display)
	return cmd, display, nil
}

func StartBrowser(initUrl string) (*exec.Cmd, error) {
	display := os.Getenv("DISPLAY")
	url := initUrl
	commands := []string{
		"--user-data-dir=./tmp/chrome-display99",
		"--no-sandbox",
		"--disable-gpu",
		"--disable-dev-shm-usage",
		"--disable-software-rasterizer",
		"--window-size=1920,1080",
		"--window-position=0,0",
		"--start-maximized",
		url,
	}
	// Start browser
	cmd := exec.Command("google-chrome", commands...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	fmt.Println("STARTING:Google Chrome with Xvfb at port:" + display)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting browser: %s\n", err)
		return nil, err
	}
	fmt.Println("Started Chrome at port:", display)

	return cmd, nil

}
