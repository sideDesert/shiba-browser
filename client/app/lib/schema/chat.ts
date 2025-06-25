import { type SocketMessage } from "./message";
type ChatPayload = {
  id: string;
  sender_name: string;
  content: string;
  created_at: string;
};

type ChatMessage = SocketMessage<ChatPayload>;

// const subject = "chatrooms.chat.<cid>";
// const sender = "<uid>";
