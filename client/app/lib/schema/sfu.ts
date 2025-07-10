import { type SocketMessage } from "./message";

export const SfuType = {
  ice: "ice",
  sdp: "sdp",
} as const;

export const SfuSignal = {
  serverAnswer: "server-answer",
  serverOffer: "server-offer",
  serverRequest: "server-request",
  serverTrickle: "server-trickle",
  ClientTrickle: "client-trickle",
} as const;

//_type => server-answer, server-offer, sever-request
const iceSubject = "chatrooms.sfu.ice.<sigcode>.<uid>";
const sdpSubject = "chatrooms.sfu.sdp.<sigcode>.<uid>";
const sender = "<uid>";

export type IcePayload = {
  trickle: RTCIceCandidateInit;
  type: "pub" | "sub";
};
export type SdpPayload = {
  offer: RTCSessionDescriptionInit;
  type: "pub" | "sub";
};
type b = RTCSessionDescription;

export type RequestPayload = {};

export type IceSfuMessage = SocketMessage<IcePayload>;
export type SdpSfuMessage = SocketMessage<SdpPayload>;
export type ReqSfuMessage = SocketMessage<RequestPayload>;
