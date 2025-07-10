import { SocketMessageHeader } from "@/lib/socket-message-builder";
import type { ParserContext } from "./parser-context";
import { SfuSignal, SfuType } from "@/lib/schema/sfu";
import {
  signal0,
  signal1,
  signal2,
  signal3,
  signal4,
  signal5,
  signal7,
  signal8,
  signal9,
  signal10,
  signal11,
} from "@/lib/schema/signal";
import { ChatHandlerStrategy } from "./handler-strategy/chat";
import {
  ClientTrickleStrategy,
  ServerAnswerStrategy,
  ServerOfferStrategy,
  ServerRequestStrategy,
  ServerTrickleStrategy,
} from "./handler-strategy/sfu";
import {
  Signal0HandlerStrategy,
  Signal10HandlerStrategy,
  Signal11HandlerStrategy,
  Signal1HandlerStrategy,
  Signal2HandlerStrategy,
  Signal3HandlerStrategy,
  Signal4HandlerStrategy,
  Signal5HandlerStrategy,
  Signal7HandlerStrategy,
  Signal8HandlerStrategy,
  Signal9HandlerStrategy,
} from "./handler-strategy/signal";
import type { SocketMessage } from "@/lib/schema/message";
import type { ShibaMessageManager } from "../main";
import type { HandlerStrategy } from "./handler-strategy/base";

interface HandlerMap {
  [key: string]: HandlerStrategy | HandlerMap;
}

export class Handler {
  private handlerMap;
  private manager;

  constructor(private socket: WebSocket, manager: ShibaMessageManager) {
    this.manager = manager;
    this.handlerMap = {
      [SocketMessageHeader.chat]: new ChatHandlerStrategy(this.manager),
      [SocketMessageHeader.sfu]: {
        [SfuType.ice]: {
          [SfuSignal.ClientTrickle]: new ClientTrickleStrategy(this.manager),
          [SfuSignal.serverTrickle]: new ServerTrickleStrategy(this.manager),
        },
        [SfuType.sdp]: {
          [SfuSignal.serverRequest]: new ServerRequestStrategy(this.manager),
          [SfuSignal.serverOffer]: new ServerOfferStrategy(this.manager),
          [SfuSignal.serverAnswer]: new ServerAnswerStrategy(this.manager),
        },
      },
      [SocketMessageHeader.signal]: {
        [signal0]: new Signal0HandlerStrategy(this.manager),
        [signal1]: new Signal1HandlerStrategy(this.manager),
        [signal2]: new Signal2HandlerStrategy(this.manager),
        [signal3]: new Signal3HandlerStrategy(this.manager),
        [signal4]: new Signal4HandlerStrategy(this.manager),
        [signal5]: new Signal5HandlerStrategy(this.manager),
        [signal7]: new Signal7HandlerStrategy(this.manager),
        [signal8]: new Signal8HandlerStrategy(this.manager),
        [signal9]: new Signal9HandlerStrategy(this.manager),
        [signal10]: new Signal10HandlerStrategy(this.manager),
        [signal11]: new Signal11HandlerStrategy(this.manager),
      },
    } as const;
  }

  handle(ctx: ParserContext, msg: SocketMessage<unknown>) {
    let handler = this.handlerMap[ctx.header];
    if (ctx.type) {
      handler = (handler as Record<string, any>)[ctx.type];
    }
    if (ctx.sigcode) {
      handler = (handler as Record<string, any>)[ctx.sigcode];
    }
    handler = handler as HandlerStrategy;
    handler.handle(ctx, msg);
  }
}
