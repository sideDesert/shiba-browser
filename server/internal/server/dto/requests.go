package dto

import (
	"fmt"
	"log"
	"regexp"
	"sideDesert/shiba/internal/server/lib"
	"strconv"
)

type SignupUserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChatHistoryRequest struct {
	Sender     string `json:"sender"`
	ChatroomId string `json:"chatroom_id"`
	Offset     int    `json:"offset"`
}

type CreateChatRoomRequest struct {
	Name           string   `json:"name"`
	ProfilePicture string   `json:"profile_picture"`
	DirectMessage  bool     `json:"direct_message"`
	Participants   []string `json:"participants"`
}

type Message[T any] struct {
	Sender  string `json:"sender"`
	Subject string `json:"subject"`
	Payload T      `json:"payload"`
}

type ChatMessagePayload struct {
	Id         string `json:"id"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

type RemoteMessagePayload struct {
	Type  string `json:"type"`  // cursor.move, cursor.left_click, cursor.right_click, key
	Value string `json:"value"` // [aA-zZ], "(x, y)"
}

type FriendStatusRequest struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

type SendFriendRequest struct {
	FriendId string `json:"friend_id"`
}

type GetChatroomRemote struct {
	ChatroomId string `json:"chatroom_id"`
}

type PatchChatroomRemoteRequest struct {
	ChatroomId string `json:"chatroom_id"`
	UserId     string `json:"user_id"`
}

type ChangeChatroomRemoteRequest = PatchChatroomRemoteRequest

func (p *RemoteMessagePayload) ExtractKeys(code int) (string, error) {
	defErr := fmt.Errorf("code is not of type Keys")
	switch code {
	case lib.Key:
		return p.Value, nil
	default:
		return "", defErr
	}
}

func (payload *RemoteMessagePayload) ValidateRemotePayload() (int, bool) {
	if payload.Type == "key" {
		return lib.Key, true
	}

	re := regexp.MustCompile(lib.CURSOR_Value_REGEX)
	if payload.Type == "cursor.move" {
		if !re.MatchString(payload.Value) {
			log.Println("🚨Error in ValidateRemotePayload[cursor.move.Regex]: Not of form (x,y)")
			return lib.CursorMove, false
		}
		return lib.CursorMove, true
	}

	if payload.Type == "cursor.click" {
		if !re.MatchString(payload.Value) {
			log.Println("🚨Error in ValidateRemotePayload[cursor.click.Regex]: Not of form (x,y)")
			return lib.CursorClick, false
		}
		return lib.CursorClick, true
	}

	return lib.Undefined, false
}

func (p *RemoteMessagePayload) ExtractCursorPos(code int) (float32, float32, error) {
	re := regexp.MustCompile(lib.CURSOR_Value_REGEX)
	matches := re.FindStringSubmatch(p.Value)
	defErr := fmt.Errorf("code is not of type Cursor")
	switch code {
	case lib.CursorMove:
		if re.MatchString(p.Value) {
			x, err := strconv.ParseFloat(matches[1], 32)
			if err != nil {
				log.Println("🚨Error[ExtractCursorPos.X]:", err)
				return 0.0, 0.0, err
			}
			y, err := strconv.ParseFloat(matches[2], 32)
			if err != nil {
				log.Println("🚨Error[ExtractCursorPos.Y]:", err)
				return 0.0, 0.0, err
			}
			return float32(x), float32(y), nil
		}
	case lib.CursorClick:
		if re.MatchString(p.Value) {
			x, err := strconv.ParseFloat(matches[1], 32)
			if err != nil {
				log.Println("🚨Error[ExtractCursorPos.X]:", err)
				return 0.0, 0.0, err
			}
			y, err := strconv.ParseFloat(matches[2], 32)
			if err != nil {
				log.Println("🚨Error[ExtractCursorPos.Y]:", err)
				return 0.0, 0.0, err
			}
			return float32(x), float32(y), nil
		}
	default:
		return 0, 0, defErr
	}

	return 0, 0, nil
}
