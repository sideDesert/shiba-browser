package lib

import (
	"database/sql"
	"time"
)

/*
CREATE TABLE users (

	id SERIAL PRIMARY KEY,
	user_id VARCHAR(255) UNIQUE NOT NULL,
	name VARCHAR(255) NOT NULL,
	email VARCHAR(255) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	username VARCHAR(255) NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	profile_picture VARCHAR(255) NULL,
	status VARCHAR(50) NULL

);
*/
type User struct {
	Id             int32          `json:"id"`
	Name           string         `json:"first_name"`
	UserId         string         `json:"user_id"`
	Username       string         `json:"username"`
	Email          string         `json:"email"`
	PasswordHash   string         `json:"password_hash"`
	CreatedAt      time.Time      `json:"created_at"`
	Status         sql.NullString `json:"status"`
	ProfilePicture sql.NullString `json:"profile_picture"`
}

/*
CREATE TABLE session (

	id SERIAL PRIMARY KEY,
	user_id VARCHAR(255) NOT NULL,
	token VARCHAR(255) NOT NULL,
	expires_at TIMESTAMP NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE

);
*/
type Session struct {
	Id        int32     `json:"id"`
	UserId    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_ai"`
}

type Chatroom struct {
	Id             string         `json:"id"`
	Name           string         `json:"name"`
	ProfilePicture sql.NullString `json:"profile_picture"`
	DirectMessage  bool           `json:"direct_message"`
	CreatedAt      time.Time      `json:"created_at"`
}

type Message struct {
	Id         string         `json:"id"`
	Sender     string         `json:"sender"`
	SenderName string         `json:"sender_name"`
	ChatroomId string         `json:"chatroom_id"`
	Content    string         `json:"content"`
	Status     sql.NullString `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
}

type UserChatroom struct {
	UserId     string `json:"user_id"`
	ChatroomId string `json:"chatroom_id"`
}

type UserChatrooms = []UserChatroom

type UserId = string

type ReadReceipt struct {
	Id        string    `json:"id"`
	MessageId string    `json:"message_id"`
	UserId    string    `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

type FriendRelations struct {
	Id        string         `json:"id"`
	UserId1   string         `json:"user_id1"`
	UserId2   string         `json:"user_id2"`
	Status    sql.NullString `json:"status"`
	UpdatedAt time.Time      `json:"updated_at"` // Pointer to allow NULL
	CreatedAt time.Time      `json:"created_at"` // Pointer to allow NULL
}

type Remote struct {
	ChatroomId string `json:"chatroom_id"`
	UserId     string `json:"user_id"`
}
