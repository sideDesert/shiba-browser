// import { type SocketMessage } from "./schema/message";
// import { SocketMessageBuilder } from "./socket-message-builder";
import { ShibaMessageManager } from "./socket-message-manager.ts/main";

// export class ShibaSocket {
//   private manager: SocketMessageManager;
//   private ws;

//   constructor(private wsUrl: string) {
//     this.ws = new WebSocket(this.wsUrl);
//     this.manager = new SocketMessageManager(this.ws);

//     this.ws.onopen = () => console.log("Connected to WebSocket");
//     this.ws.onmessage = (e) => {
//       this.manager.handle(e.data);
//     };
//     this.ws.onerror = (err: Event) => console.error("Error:", err);
//     this.ws.onclose = () => console.log("Disconnected from WebSocket");
//   }

//   getManager() {
//     return this.manager;
//   }

//   getSocket() {
//     return this.ws;
//   }

//   send(msg: string) {
//     this.ws.send(msg);
//   }

//   close() {
//     this.ws.close();
//   }
// }

// export function NewRemoteMessage(
//   senderId: string,
//   chatroomId: string,
//   payload: RemoteMessagePayload
// ) {
//   return {
//     sender: senderId,
//     subject: "stream.remote." + chatroomId,
//     payload
//   }
// }

// export function NewStreamMessage(
//   senderId: string,
//   chatroomId: string,
//   payload: Record<string, unknown>
// ) {
//   let offerType = "";
//   const p = payload;

//   if (Object.keys(p).includes("candidate")) {
//     offerType = "ice";
//   } else if (Object.keys(p).includes("sdp")) {
//     offerType = "sdp";
//   } else {
//     console.group("DEBUG");
//     console.log("Payload", p);
//     console.error("payload is not valid offerType, recieved", payload);
//     console.groupEnd();
//   }

//   const subject = `stream.${offerType}.${chatroomId}`;
//   return {
//     sender: senderId,
//     subject,
//     payload,
//   };
// }

// export function NewWebrtcMessage(
//   senderId: string,
//   chatroomId: string,
//   payload: Record<string, unknown>
// ) {
//   let offerType = "";
//   const p = payload;

//   if (Object.keys(p).includes("candidate")) {
//     offerType = "ice";
//   } else if (Object.keys(p).includes("sdp")) {
//     offerType = "sdp";
//   } else {
//     console.group("DEBUG");
//     console.log("Payload", p);
//     console.error("payload is not valid offerType, recieved", payload);
//     console.groupEnd();
//   }

//   const subject = `webrtc.${offerType}.${chatroomId}`;
//   return {
//     sender: senderId,
//     subject,
//     payload,
//   };
// }

// export function NewWsChatMessage(
//   sender: string,
//   senderName: string,
//   content: string,
//   chatroomId: string
// ): Message<ChatMessagePayload> {
//   return {
//     subject: `chat.${chatroomId}`,
//     sender: sender,
//     payload: {
//       id: crypto.randomUUID(),
//       sender_name: senderName,
//       content: content,
//       created_at: new Date().toISOString(),
//     },
//   };
// }

// export type ChatMessage = {
//   sender: string;
//   chatroom_id: string;
//   id: string;
//   sender_name: string;
//   content: string;
//   created_at: string;
// };

// export function NewChatMessage(
//   sender: string,
//   senderName: string,
//   content: string,
//   chatroomId: string
// ): ChatMessage {
//   return {
//     sender: sender,
//     chatroom_id: chatroomId,
//     id: crypto.randomUUID(),
//     sender_name: senderName,
//     content: content,
//     created_at: new Date().toISOString(),
//   };
// }
