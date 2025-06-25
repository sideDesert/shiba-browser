import { type SocketMessage } from "./message";

/*
 * CLIENT SIDE TO SERVER SIDE SIGNALS
 */

export const prefix = "chatrooms.signal.";

// 1. Init Indiv Calls  (signal.init-ic.<cid>)
export const signal0 = "init-ic.<cid>";
export const subject0 = prefix + signal0;
export type Signal0Payload = {};
export type Signal0Message = SocketMessage<Signal0Payload>;

// 2. Init Group Calls  (signal.init-gc.<cid>)
export const signal1 = "init-gc.<cid>";
export const subject1 = prefix + signal1;
export type Signal1Payload = {};
export type Signal1Message = SocketMessage<Signal1Payload>;

// 3. Indiv Call Answer (signal.ans-ic.<cid>)
export const signal2 = "ans-ic.<cid>";
export const subject2 = prefix + signal2;
export type Signal2Payload = {
  answer: "accept" | "decline";
  sdp: string;
};
export type Signal2Message = SocketMessage<Signal2Payload>;

// 12. Ack Indiv Call (signal.ack-ic.<cid>)
export const signal11 = "ack-ic.<cid>";
export const subject11 = prefix + signal11;
export type Signal11Payload = {
  sdp: string;
};
export type Signal11Message = SocketMessage<Signal11Payload>;

// 4. Group Call Answer (signal.end-call.<cid>)
export const signal3 = "end-call.<cid>";
export const subject3 = prefix + signal3;
export type Signal3Payload = {};
export type Signal3Message = SocketMessage<Signal3Payload>;

// 5. Join Call (signal.ans-gc.<cid>)
export const signal4 = "ans-gc.<cid>";
export const subject4 = prefix + signal4;
export type Signal4Payload = {
  answer: "join" | "leave";
  sdp?: string;
};
export type Signal4Message = SocketMessage<Signal4Payload>;

// 6. Start Stream (signal.start-stream.<cid>)
export const signal5 = "start-stream.<cid>";
export const subject5 = prefix + signal5;
export type Signal5Payload = {};
export type Signal5Message = SocketMessage<Signal5Payload>;

// 8. Join Stream (signal.join-stream.<cid>)
export const signal7 = "join-stream.<cid>";
export const subject7 = prefix + signal7;
export type Signal7Payload = {};
export type Signal7Message = SocketMessage<Signal7Payload>;

// 9. Leave Stream (signal.leave-stream.<cid>)
export const signal8 = "leave-stream.<cid>";
export const subject8 = prefix + signal8;
export type Signal8Payload = {};
export type Signal8Message = SocketMessage<Signal8Payload>;

// 10. End Stream (signal.end-stream.<cid>)
export const signal9 = "end-stream.<cid>";
export const subject9 = prefix + signal9;
export type Signal9Payload = {};
export type Signal9Message = SocketMessage<Signal9Payload>;

/*
 * SERVER TO CLIENT SIGNALS
 */

// 11. Remote Action (signal.remote.<cid>)
export const signal10 = "remote.<cid>";
export const subject10 = prefix + signal10;
export type Signal10Payload = {
  action: "send" | "keys" | "left-click" | "right-click";
  value: string;
};
export type Signal10Message = SocketMessage<Signal10Payload>;

export const signalSubject = [
  subject0,
  subject1,
  subject2,
  subject3,
  subject4,
  subject5,
  subject7,
  subject8,
  subject9,
  subject10,
  subject11,
];

export const signals = [
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
];

export const sigtable = signalSubject.reduce<Record<string, number>>(
  (acc, curr, idx) => {
    acc[curr] = idx;
    return acc;
  },
  {}
);
