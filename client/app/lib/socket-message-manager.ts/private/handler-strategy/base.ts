import { ShibaMessageManager } from "../../main";
import { type ParserContext } from "../parser-context";
import { type SocketMessage } from "@/lib/schema/message";

export abstract class HandlerStrategy {
  constructor(private manager: ShibaMessageManager) {}
  getManager() {
    return this.manager;
  }
  abstract handle(ctx: ParserContext, msg: SocketMessage<unknown>): void;
}
