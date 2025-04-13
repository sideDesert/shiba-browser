package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"log"
	"os"
	"sideDesert/shiba/internal/server/dto"
	"sideDesert/shiba/internal/server/lib"
	"sideDesert/shiba/internal/server/store"

	"golang.org/x/oauth2"
)

type ServerConfig struct {
	DbUrl string
}

type Service struct {
	Store       *store.Store
	Ctx         context.Context
	oauthConfig *oauth2.Config
	config      *ServerConfig
}

func NewService(ctx context.Context, config *ServerConfig) (*Service, error) {
	store, err := store.NewStore(ctx, config.DbUrl)
	if err != nil {
		log.Println("Error in NewService[NewStore()]:", err)
		return nil, err
	}

	return &Service{
		Ctx:    ctx,
		Store:  store,
		config: config,
		oauthConfig: createGoogleOAuthConfig(
			os.Getenv("CLIENT_ID"),
			os.Getenv("CLIENT_SECRET"),
			// Note: Redirect URL is the same is CALLBACK_URL
			os.Getenv("CALLBACK_URL"),
		),
	}, nil
}

func (s *Service) RegisterUser(user *dto.SignupUserRequest) (*dto.SignupUserResponse, error) {
	userId, err := s.Store.CreateUser(s.Ctx, user)

	if err != nil {
		log.Print("Error in Registering User:", err)
		return nil, err
	}

	return &dto.SignupUserResponse{
		Name:   user.Name,
		Email:  user.Email,
		UserId: *userId,
	}, nil
}

func (s *Service) GenerateOAuthLoginURL(provider string) string {
	config := s.oauthConfig
	var oauthStateString = "state"

	url := config.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	return url
}

func (s *Service) Logout() error {
	return nil
}

func (s *Service) Health() (string, error) {
	return "Healthy!", nil
}

func (s *Service) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return s.oauthConfig.Exchange(s.Ctx, code)
}

func (s *Service) GetOAuthUserData(token *oauth2.Token) (*dto.OAuthGoogleUserDataResponse, error) {
	client := s.oauthConfig.Client(s.Ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	defer resp.Body.Close()

	userDataResponse := dto.OAuthGoogleUserDataResponse{}

	err = json.NewDecoder(resp.Body).Decode(&userDataResponse)
	if err != nil {
		return nil, err
	}
	return &userDataResponse, nil
}

func (s *Service) GetUserResponse(w http.ResponseWriter, r *http.Request) (*dto.UserResponse, error) {
	if r.Method != http.MethodGet {
		return nil, fmt.Errorf("Method %s not allowed", r.Method)
	}

	tokenCookie, err := r.Cookie("shiba-auth-token")

	if err != nil {
		log.Println("Error in handleUser: No token in Cookies named shiba-auth-token")
		return nil, fmt.Errorf("No token in Cookies named shiba-auth-token")
	}

	cookieString := tokenCookie.String()
	tokenString := strings.Split(cookieString, "=")[1]
	token, err := lib.VerifyToken(tokenString)
	if err != nil {
		log.Println("Error in handleUser:", err)
		return nil, fmt.Errorf("Token Validation Error")
	}

	exp_time, err := token.Claims.GetExpirationTime()
	if err != nil {
		log.Println("Error in handleUser: No Expiration Time for token")
		return nil, fmt.Errorf("No Expiration Time for token")
	}

	if exp_time.Before(time.Now()) {
		log.Println("Error in handleUser: Token has expired")
		return nil, fmt.Errorf("Token Expired")
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		log.Println("Error in handleUser: No Subject in Token")
		return nil, err
	}

	if userID == "" {
		log.Println("Error in handleUser: UserID is empty string")
		return nil, fmt.Errorf("UserID is empty")
	}

	userDetails, err := s.Store.GetUserById(s.Ctx, userID)
	if err != nil {
		log.Println("Error in handleUser[GetUserById]", err)
		return nil, fmt.Errorf("User could not be fetched from DB")
	}

	return &dto.UserResponse{
		UserId:         userID,
		Email:          userDetails.Email,
		Username:       userDetails.Username,
		Name:           userDetails.Name,
		ProfilePicture: userDetails.ProfilePicture.String,
		CreatedAt:      userDetails.CreatedAt,
	}, nil
}

func (s *Service) GetUserChatRooms(userId string) ([]lib.Chatroom, error) {
	chatrooms, err := s.Store.GetChatRoomsByUserId(s.Ctx, userId)
	if err != nil {
		log.Println("Error in GetUserChatRooms[userId]", err)
		return nil, fmt.Errorf("Chatrooms could not be fetched from DB")
	}
	return chatrooms, nil
}

func (s *Service) CreateChatRoom(crr dto.CreateChatRoomRequest) (string, error) {
	return s.Store.CreateChatRoom(s.Ctx, crr)
}

func (s *Service) StoreChatMessage(senderId string, chatroomId string, msg dto.ChatMessagePayload) error {

	temp := store.StoreChatMessageDto{
		Sender:     senderId,
		ChatroomId: chatroomId,
		Id:         msg.Id,
		SenderName: msg.SenderName,
		Content:    msg.Content,
		CreatedAt:  msg.CreatedAt,
	}
	err := s.Store.StoreChatRoomMessage(s.Ctx, temp)
	if err != nil {
		log.Println("❌ Error in StoreChatMessage:", err)
		return err
	}
	return nil
}

func (s *Service) GetChatroomHistory(chatroomId string, offset int) ([]lib.Message, error) {
	messages, err := s.Store.GetLast50ChatRoomMessages(s.Ctx, chatroomId, offset)
	if err != nil {
		log.Println("❌ Error in GetChatroomHistory:", err)
		return nil, err
	}

	return messages, nil
}

func (s *Service) SearchUsers(userId string, query string, limit int) ([]dto.SearchUserResponse, error) {
	users, err := s.Store.SearchUsers(s.Ctx, userId, query, limit)
	if err != nil {
		log.Println("Error in SearchUsers:", err)
		return users, err
	}

	return users, nil
}

func (s *Service) GetFriends(userId string) ([]store.UserFriend, error) {
	users, err := s.Store.GetFriendsByUserId(s.Ctx, userId)
	if err != nil {
		log.Println("Error in SearchUsers:", err)
		return users, err
	}
	return users, nil
}

func (s *Service) HandleFriendRequest(userId string, reqId string, status string) error {
	// TODO: Add Check for - Can this userId even accept this friend request?
	err := s.Store.ChangeFriendStatus(s.Ctx, reqId, status)
	if err != nil {
		log.Println("Error in SearchUsers:", err)
		return err
	}
	return nil
}

func (s *Service) SendFriendRequest(userId string, friendId string) error {
	err := s.Store.InsertFriendStatus(s.Ctx, userId, friendId, "pending")

	if err != nil {
		log.Println("Error in SendFriendRequest:", err)
		return err
	}
	return nil
}

func (s *Service) GetFriendRequests(userId string) ([]dto.FriendRequestResponse, error) {
	friends, err := s.Store.GetFriendRequestsByUserId(s.Ctx, userId)

	if err != nil {
		log.Println("Error in SendFriendRequest:", err)
		return nil, err
	}

	return friends, nil
}

func (s *Service) GetChatroomRemote(chatroomId string) (*dto.RemoteResponse, error) {
	remote, err := s.Store.GetRemoteByChatroomId(s.Ctx, chatroomId)

	if err != nil {
		log.Println("Error in SendFriendRequest:", err)
		return nil, err
	}
	return remote, nil
}

func (s *Service) ChangeChatroomRemote(chatroomId string, userId string) error {
	err := s.Store.UpdateRemote(s.Ctx, chatroomId, userId)

	if err != nil {
		log.Println("Error in SendFriendRequest:", err)
		return err
	}
	return nil
}
func (s *Service) CheckUserIsRemoteForChatroom(userId string, chatroomId string) bool {
	remote, err := s.Store.GetRemoteByChatroomId(s.Ctx, chatroomId)

	if err != nil {
		log.Println("Error in SendFriendRequest:", err)
		return false
	}

	return remote.UserId == userId
}
