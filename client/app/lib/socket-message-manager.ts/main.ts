import { type SocketMessage } from "../schema/message";
import {
  ChatParserStrategy,
  SFUParserStrategy,
  SignalParserStrategy,
  type ParserStrategy,
} from "./private/parser-strategy";
import { Handler } from "./private/handler";
import { SFUConnection } from "../sfu-manager";
import type { ChatPayload } from "../schema/chat";
import {
  SocketMessageBuilder,
  SocketMessageHeader,
} from "../socket-message-builder";

const parserMap = {
  chat: new ChatParserStrategy(),
  sfu: new SFUParserStrategy(),
  signal: new SignalParserStrategy(),
};

export type SideEffects = {
  addChatMessage: (msg: SocketMessage<ChatPayload>) => void;
  setLocalVideoStream: (stream: MediaStream) => void;
  setIsIncomingCall: (val: boolean) => void;
  setIsOutgoingCall: (val: boolean) => void;
  setRemoteVideoStream: (stream: MediaStream) => void;
};

export class ShibaMessageManager {
  private parser;
  private handler;
  private ws;
  private _resolveCall: (val: boolean) => void;
  public conns;

  constructor(
    public wsUrl: string,
    public userId: string,
    public chatroomId: string,
    public sideEffects: SideEffects
  ) {
    this._resolveCall = (val: boolean) => {};
    this.sideEffects = sideEffects;
    this.ws = new WebSocket(wsUrl);
    this.setupWs();
    this.conns = {
      vc: new SFUConnection(this.ws, userId),
    };
    this.handler = new Handler(this.ws, this);
    this.parser = new Parser(parserMap);
  }

  close() {
    this.conns.vc.close();
    this.ws.close();
  }

  send(data: SocketMessage<unknown>) {
    this.ws.send(JSON.stringify(data));
  }

  public waitForUserDecision(): Promise<boolean> {
    return new Promise((resolve) => {
      this._resolveCall = resolve;
    });
  }

  public acceptCall() {
    this._resolveCall(true);
  }

  public rejectCall() {
    this._resolveCall(false);
    // Send declined call signal
  }

  private handle(msg: string) {
    const msgJson: SocketMessage<unknown> = JSON.parse(msg);
    const ctx = this.parser.parse(msgJson);
    if (!ctx) {
      console.error("ctx is undefined in SocketMessageManager.handle");
      return;
    }
    this.handler.handle(ctx, msgJson);
  }

  private setupWs() {
    this.ws.onopen = () => console.log("Connected to WebSocket");
    this.ws.onmessage = (e) => {
      try {
        this.handle(e.data);
      } catch (err) {
        console.error("Error handling e.data", err);
      }
    };
    this.ws.onerror = (err: Event) => console.error("Websocket Error:", err);
    this.ws.onclose = (event) =>
      console.log("Disconnected from WebSocket", event);
  }
}

class Parser {
  constructor(private parserMap: Record<string, ParserStrategy>) {}

  parse(msg: SocketMessage<unknown>) {
    const header = msg.subject.split(".")[1];
    const parser = this.parserMap?.[header];
    if (parser) {
      const ctx = parser.parse(msg);
      return ctx;
    }
  }
}
