import { SocketMessageBuilder } from "./socket-message-builder";
import { SocketMessageHeader } from "./socket-message-builder";
import { SfuType } from "./schema/sfu";
import { SfuSignal } from "./schema/sfu";
import { type IcePayload } from "./schema/sfu";

export class SFUConnection {
  sub: RTCPeerConnection;
  pub: RTCPeerConnection;

  subRemoteDescriptionSet: boolean;
  pubRemoteDescriptionSet: boolean;
  subIceCandidates: Array<RTCIceCandidateInit>;
  pubIceCandidates: Array<RTCIceCandidateInit>;

  constructor(ws: WebSocket, userId: string) {
    this.sub = new RTCPeerConnection();
    this.pub = new RTCPeerConnection();
    this.subRemoteDescriptionSet = false;
    this.pubRemoteDescriptionSet = false;
    this.subIceCandidates = new Array();
    this.pubIceCandidates = new Array();
    this.setupIceCandidateHandler(ws, userId);
  }

  close() {
    this.sub.close();
    this.pub.close();
  }

  private setupIceCandidateHandler(socket: WebSocket, userId: string) {
    this.sub.onicecandidate = (event) => {
      if (event.candidate && !!socket) {
        const msg = new SocketMessageBuilder();
        msg.setSender(userId);
        msg.setHeader(SocketMessageHeader.sfu);
        msg.setUid(userId);
        msg.setType(SfuType.ice);
        msg.setSigcode(SfuSignal.ClientTrickle);
        const payload: IcePayload = {
          type: "sub",
          trickle: event.candidate,
        };
        msg.setPayload(payload);
        socket.send(JSON.stringify(msg.build()));
      }
    };

    this.pub.onicecandidate = (event) => {
      if (event.candidate && !!socket) {
        const msg = new SocketMessageBuilder();
        msg.setSender(userId);
        msg.setHeader(SocketMessageHeader.sfu);
        msg.setUid(userId);
        msg.setType(SfuType.ice);
        msg.setSigcode(SfuSignal.ClientTrickle);
        const payload: IcePayload = {
          type: "pub",
          trickle: event.candidate,
        };
        msg.setPayload(payload);
        socket.send(JSON.stringify(msg.build()));
      }
    };
  }
}
