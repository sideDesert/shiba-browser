import type { SocketMessage } from "@/lib/schema/message";
import { HandlerStrategy } from "./base";
import type { SFUParserContext } from "../parser-context";
import type { IcePayload, RequestPayload, SdpPayload } from "@/lib/schema/sfu";
import {
  SocketMessageBuilder,
  SocketMessageHeader,
} from "@/lib/socket-message-builder";
import { signal11, type Signal11Payload } from "@/lib/schema/signal";
import { getMediaStream } from "@/lib/utils";

export class ClientTrickleStrategy extends HandlerStrategy {
  handle(ctx: SFUParserContext, msg: SocketMessage<IcePayload>) {}
}
export class ServerTrickleStrategy extends HandlerStrategy {
  handle(ctx: SFUParserContext, msg: SocketMessage<IcePayload>) {
    const payload = msg.payload;
    if (payload.type === "pub") {
      if (this.getManager().conns.vc.pubRemoteDescriptionSet) {
        this.getManager().conns.vc.pub.addIceCandidate(payload.trickle);
      } else {
        this.getManager().conns.vc.pubIceCandidates.push(payload.trickle);
      }
    }

    if (payload.type === "sub") {
      if (this.getManager().conns.vc.subRemoteDescriptionSet) {
        this.getManager().conns.vc.subIceCandidates.push(payload.trickle);
      } else {
        this.getManager().conns.vc.sub.addIceCandidate(payload.trickle);
      }
    }
  }
}
export class ServerRequestStrategy extends HandlerStrategy {
  async handle(ctx: SFUParserContext, msg: SocketMessage<RequestPayload>) {
    const manager = this.getManager();

    manager.sideEffects.setIsOutgoingCall(false);
    const peerVideoRTCConn = manager.conns.vc;
    const userId = manager.userId;

    try {
      const stream = await getMediaStream();
      if (!stream) {
        console.error("No media stream available");
        return;
      }

      stream.getTracks().forEach((track) => {
        peerVideoRTCConn.pub.addTrack(track, stream);
      });

      const offer = await peerVideoRTCConn.pub.createOffer();
      await peerVideoRTCConn.pub.setLocalDescription(offer);
      const offerMsg = new SocketMessageBuilder();
      offerMsg.setSender(userId);
      offerMsg.setHeader(SocketMessageHeader.signal);
      offerMsg.setUid(userId);
      offerMsg.setType(signal11);
      const offerMsgPayload: Signal11Payload = {
        sdp: offer,
      };
      offerMsg.setPayload(offerMsgPayload);
      manager.send(offerMsg.build());

      manager.sideEffects.setLocalVideoStream(stream);
    } catch (err) {
      console.error("Error in ServerRequestStrategy", err);
    }
  }
}

export class ServerOfferStrategy extends HandlerStrategy {
  async handle(ctx: SFUParserContext, msg: SocketMessage<SdpPayload>) {
    const manager = this.getManager();
    const peerVideoRTCConn = manager.conns.vc;
    const payload = msg.payload;
    if (payload.type === "sub") {
      peerVideoRTCConn.sub.setRemoteDescription(payload.offer);
      peerVideoRTCConn.subRemoteDescriptionSet = true;
      for (const candidate of peerVideoRTCConn.subIceCandidates) {
        await peerVideoRTCConn.sub.addIceCandidate(candidate);
      }
      peerVideoRTCConn.subIceCandidates = [];

      const clientAnswer = await peerVideoRTCConn.sub.createAnswer();

      await peerVideoRTCConn.sub.setLocalDescription(clientAnswer);

      const msg = new SocketMessageBuilder();

      //TODO: Handle signalling back the answer to the server
    }
  }
}

export class ServerAnswerStrategy extends HandlerStrategy {
  async handle(ctx: SFUParserContext, msg: SocketMessage<SdpPayload>) {
    const payload = msg.payload;
    const manager = this.getManager();
    const peerVideoRTCConn = manager.conns.vc;
    if (payload.type === "pub") {
      peerVideoRTCConn.pub.setRemoteDescription(payload.offer);
      peerVideoRTCConn.pubRemoteDescriptionSet = true;
      for (const candidate of peerVideoRTCConn.pubIceCandidates) {
        await peerVideoRTCConn.pub.addIceCandidate(candidate);
      }
      peerVideoRTCConn.pubIceCandidates = [];
    }
  }
}
