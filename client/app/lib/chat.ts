import type { ChatMessage, ChatMessagePayload } from "./types";
import type { Message } from "./types";

export const createSocket = (
  wsUrl: string,
  messageHandler: (msg: Message<unknown>) => void
) => {
  const ws = new WebSocket(wsUrl);
  ws.onopen = () => console.log("Connected to WebSocket");
  ws.onmessage = (e: { data: string }) => {
    const newMsg: Message<unknown> = JSON.parse(e.data);
    messageHandler(newMsg);
  };
  ws.onerror = (err: Event) => console.error("Error:", err);
  ws.onclose = () => console.log("Disconnected from WebSocket");

  return ws;
};

export function NewStreamMessage(
  senderId: string,
  chatroomId: string,
  payload: Record<string, unknown>
) {
  let offerType = "";
  const p = payload;

  if (Object.keys(p).includes("candidate")) {
    offerType = "ice";
  } else if (Object.keys(p).includes("sdp")) {
    offerType = "sdp";
  } else {
    console.group("DEBUG");
    console.log("Payload", p);
    console.error("payload is not valid offerType, recieved", payload);
    console.groupEnd();
  }

  const subject = `stream.${offerType}.${chatroomId}`;
  return {
    sender: senderId,
    subject,
    payload,
  };
}

export function NewWebrtcMessage(
  senderId: string,
  chatroomId: string,
  payload: Record<string, unknown>
) {
  let offerType = "";
  const p = payload;

  if (Object.keys(p).includes("candidate")) {
    offerType = "ice";
  } else if (Object.keys(p).includes("sdp")) {
    offerType = "sdp";
  } else {
    console.group("DEBUG");
    console.log("Payload", p);
    console.error("payload is not valid offerType, recieved", payload);
    console.groupEnd();
  }

  const subject = `webrtc.${offerType}.${chatroomId}`;
  return {
    sender: senderId,
    subject,
    payload,
  };
}

export function NewWsChatMessage(
  sender: string,
  senderName: string,
  content: string,
  chatroomId: string
): Message<ChatMessagePayload> {
  return {
    subject: `chat.${chatroomId}`,
    sender: sender,
    payload: {
      id: crypto.randomUUID(),
      sender_name: senderName,
      content: content,
      created_at: new Date().toISOString(),
    },
  };
}

export function NewChatMessage(
  sender: string,
  senderName: string,
  content: string,
  chatroomId: string
): ChatMessage {
  return {
    sender: sender,
    chatroom_id: chatroomId,
    id: crypto.randomUUID(),
    sender_name: senderName,
    content: content,
    created_at: new Date().toISOString(),
  };
}
