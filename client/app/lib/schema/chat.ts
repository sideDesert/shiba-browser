import { type SocketMessage } from "./message";
export type ChatPayload = {
  id: string;
  sender_name: string;
  content: string;
  created_at: string;
};

export type ChatMessage = SocketMessage<ChatPayload>;

// const subject = "chatrooms.chat.<cid>";
// const sender = "<uid>";
