import type { SocketMessage } from "@/lib/schema/message";
import { HandlerStrategy } from "./base";
import type { ChatParserContext } from "../parser-context";
import type { ChatMessage } from "@/lib/schema/chat";

export class ChatHandlerStrategy extends HandlerStrategy {
  handle(ctx: ChatParserContext, msg: SocketMessage<unknown>) {
    if (msg.sender == this.getManager().userId) return;
    this.getManager().sideEffects.addChatMessage(msg as ChatMessage);
  }
}
