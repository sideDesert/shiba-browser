import type { SocketMessage } from "@/lib/schema/message";
import {
  ChatParserContext,
  SFUParserContext,
  SignalParserContext,
  type ParserContext,
} from "./parser-context";
import { SocketMessageHeader } from "@/lib/socket-message-builder";

export abstract class ParserStrategy {
  abstract parse(msg: SocketMessage<unknown>): ParserContext | undefined;
}

export class ChatParserStrategy implements ParserStrategy {
  parse(msg: SocketMessage<unknown>) {
    const split = msg.subject.split(".");
    const [_, header, rid] = split;
    if (split.length !== 3) {
      return undefined;
    }
    if (header !== SocketMessageHeader.chat) return undefined;
    return new ChatParserContext(rid);
  }
}

export class SFUParserStrategy implements ParserStrategy {
  parse(msg: SocketMessage<unknown>) {
    const split = msg.subject.split(".");
    if (split.length !== 5) {
      return undefined;
    }
    const [_, header, type, sigcode, rid] = split;
    if (header !== SocketMessageHeader.sfu) return undefined;
    return new SFUParserContext(rid, type, sigcode);
  }
}

export class SignalParserStrategy implements ParserStrategy {
  parse(msg: SocketMessage<unknown>) {
    const split = msg.subject.split(".");
    if (split.length !== 4) {
      return undefined;
    }
    const [_, header, type, rid] = split;
    if (header !== SocketMessageHeader.signal) return undefined;
    return new SignalParserContext(rid, type);
  }
}
