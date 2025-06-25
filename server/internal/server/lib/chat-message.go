package lib

type ChatPayload struct {
	Id         string `json:"id"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

type ChatMessage = SocketMessage[ChatPayload]

// const subject = "chatrooms.chat.<cid>"
// const sender = "<uid>"
