import { type SocketMessage } from "./message";

const iceSubject = "chatrooms.sfu.ice.<cid>";
const sdpSubject = "chatrooms.sfu.sdp.<cid>";
const sender = "<uid>";

type IcePayload = Record<string, string>;
type SdpPayload = RTCSessionDescriptionInit;

type IceSfuMessage = SocketMessage<IcePayload>;
type SdpSfuMessage = SocketMessage<SdpPayload>;
