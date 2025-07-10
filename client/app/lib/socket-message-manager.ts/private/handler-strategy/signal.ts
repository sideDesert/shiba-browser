import {
  SocketMessageBuilder,
  SocketMessageHeader,
} from "@/lib/socket-message-builder";
import { HandlerStrategy } from "./base";
import type { SignalParserContext } from "../parser-context";
import type { SocketMessage } from "@/lib/schema/message";
import {
  signal2,
  type Signal0Payload,
  type Signal2Message,
  type Signal2Payload,
} from "@/lib/schema/signal";
import { getMediaStream } from "@/lib/utils";

export class Signal0HandlerStrategy extends HandlerStrategy {
  async handle(ctx: SignalParserContext, msg: SocketMessage<Signal0Payload>) {
    try {
      console.log("New Incoming Call Request Received!!!", msg);
      const manager = this.getManager();
      manager.sideEffects.setIsIncomingCall(true);

      const ans = await manager.waitForUserDecision();
      const message = new SocketMessageBuilder();
      const userId = manager.userId;
      const chatroomId = manager.chatroomId;
      const vcConn = manager.conns.vc;

      message.setSender(userId);
      message.setCid(chatroomId);
      message.setHeader(SocketMessageHeader.signal);
      message.setType(signal2);

      if (ans) {
        //Init-ic
        const stream = await getMediaStream();
        if (!stream) {
          console.error("No media stream available");
          return;
        }
        stream.getTracks().forEach((track) => {
          vcConn.pub.addTrack(track, stream);
        });
        manager.sideEffects.setLocalVideoStream(stream);

        // the case is accept
        const offer = await vcConn.pub.createOffer();
        await vcConn.pub.setLocalDescription(offer);

        const payload: Signal2Payload = {
          answer: "accept",
          caller_id: ctx.rid,
          sdp: offer,
        };

        message.setPayload(payload);
        debugger;
        manager.send(message.build());
      } else {
        const payload: Signal2Payload = {
          answer: "decline",
          caller_id: message.getSender(),
        };
        message.setPayload(payload);
        manager.send(message.build());
      }
    } catch (err) {
      console.error("Errorr handling signal0", err);
    }
  }
}

export class Signal1HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {}
}

export class Signal2HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext, msg: Signal2Message) {
    console.log(msg);
  }
}

export class Signal3HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {
    const peerVideoRTCConn = this.getManager().conns.vc;
    peerVideoRTCConn.pub.close();
    peerVideoRTCConn.sub.close();
  }
}

export class Signal4HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {}
}

export class Signal5HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {}
}

export class Signal7HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {}
}

export class Signal8HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {}
}

export class Signal9HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {}
}

export class Signal10HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {}
}

export class Signal11HandlerStrategy extends HandlerStrategy {
  handle(ctx: SignalParserContext) {}
}
