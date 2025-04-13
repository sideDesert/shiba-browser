package dto

import (
	"database/sql"
	"time"
)

type UserResponse struct {
	UserId         string    `json:"user_id"`
	Name           string    `json:"name"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
}

type SignupUserResponse struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	UserId string `json:"user_id"`
}

type OAuthGoogleUserDataResponse struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type CreateChatRoomResponse struct {
	ChatRoomId string `json:"chatroom_id"`
}

type SearchUserResponse struct {
	Name           string         `json:"first_name"`
	UserId         string         `json:"user_id"`
	Username       string         `json:"username"`
	Email          string         `json:"email"`
	Status         sql.NullString `json:"status"`
	ProfilePicture sql.NullString `json:"profile_picture"`
}

type FriendResponse struct {
	Status string `json:"status"`
}

type FriendRequestResponse struct {
	Name           string         `json:"name"`
	Username       string         `json:"username"`
	ProfilePicture sql.NullString `json:"profile_picture"`
	RequestId      string         `json:"request_id"`
	UserId         string         `json:"user_id"`
	CreatedAt      time.Time      `json:"created_at"`
	Status         string         `json:"status"`
}

type RemoteResponse struct {
	Name     string         `json:"name"`
	Username string         `json:"username"`
	UserId   string         `json:"user_id"`
	Status   sql.NullString `json:"status"`
}

type PatchOKResponse struct {
	Status string `json:"status"`
}
