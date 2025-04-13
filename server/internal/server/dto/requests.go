package dto

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
