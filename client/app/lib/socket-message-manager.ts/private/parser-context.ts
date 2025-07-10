import {
  SocketMessageHeader,
  type SocketMessageHeaderType,
} from "@/lib/socket-message-builder";

export interface ParserContext {
  header: SocketMessageHeaderType;
  rid: string;
  type?: string;
  sigcode?: string;
}

export class SFUParserContext implements ParserContext {
  header: SocketMessageHeaderType;
  constructor(public rid: string, public type: string, public sigcode: string) {
    this.header = SocketMessageHeader.sfu;
  }
}

export class ChatParserContext implements ParserContext {
  header: SocketMessageHeaderType;
  constructor(public rid: string) {
    this.header = SocketMessageHeader.chat;
  }
}

export class SignalParserContext implements ParserContext {
  header: SocketMessageHeaderType;
  constructor(public rid: string, public type: string) {
    this.header = SocketMessageHeader.signal;
  }
}
