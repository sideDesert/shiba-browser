// General -> "chatrooms.<sub>.<type>.<sigcode>.<rid>"
import type { ChatPayload } from "./schema/chat";
import { type SocketMessage } from "./schema/message";
import type { ClientChatMessage } from "./types";

export const SocketMessageHeader = {
  chat: "chat",
  signal: "signal",
  sfu: "sfu",
} as const;

export type SocketMessageHeaderType =
  (typeof SocketMessageHeader)[keyof typeof SocketMessageHeader];

export const SfuSignal = {
  serverAnswer: "server-answer",
  serverRequest: "server-request",
  serverOffer: "server-offer",
  serverTrickle: "server-trickle",
  ClientTrickle: "client-trickle",
} as const;

export type SfuSignalType = (typeof SfuSignal)[keyof typeof SfuSignal];

export class SocketMessageBuilder {
  private rid: string;
  private type: string;
  private sigcode: string;
  private payload: Record<string, unknown>;

  private subject: string;
  private header: string;
  private sender: string;

  constructor() {
    this.sender = "";
    this.rid = "";
    this.subject = "";
    this.sigcode = "";
    this.type = "";
    this.payload = {};
    this.header = "";
  }

  setHeader(header: string) {
    this.header = header;
  }

  setUid(userId: string) {
    this.rid = userId;
  }

  setCid(chatroomId: string) {
    this.rid = chatroomId;
  }

  setSender(userId: string) {
    this.sender = userId;
  }
  getSender() {
    return this.sender;
  }

  setType(h: string) {
    if (this.header === "chat") return;
    this.type = h;
  }

  getType() {
    return this.type;
  }

  setSigcode(sigcode: string) {
    if (this.header === SocketMessageHeader.sfu) {
      this.sigcode = sigcode;
      this.subject.replace("<sigcode>", sigcode);
    }
  }

  getSigcode() {
    return this.sigcode;
  }

  setPayload(payload: any) {
    this.payload = payload;
  }

  getPayload<T>() {
    return this.payload as T;
  }

  build<T extends Record<string, unknown>>(): SocketMessage<
    typeof this.payload
  > {
    if (this.header == SocketMessageHeader.chat) {
      this.subject = `chatrooms.chat.${this.rid}`;
    }
    if (this.header == SocketMessageHeader.sfu) {
      this.subject = `chatrooms.sfu.${this.type}.${this.sigcode}.${this.sender}`;
    }
    if (this.header == SocketMessageHeader.signal) {
      this.subject = `chatrooms.signal.${this.type}.${this.rid}`;
    }

    return {
      subject: this.subject,
      sender: this.sender,
      payload: this.payload,
    } as SocketMessage<T>;
  }

  getHeader(): string {
    return this.header;
  }

  toChatMessage(chatroomId: string): ClientChatMessage {
    return {
      sender: this.getSender(),
      chatroom_id: chatroomId,
      id: crypto.randomUUID(),
      sender_name: this.getPayload<ChatPayload>().sender_name,
      content: this.getPayload<ChatPayload>().content,
      created_at: new Date().toISOString(),
    };
  }
}
