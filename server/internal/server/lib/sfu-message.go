package lib

import "github.com/pion/webrtc/v4"

const iceSubject = "chatrooms.sfu.ice.<cid>"
const sdpSubject = "chatrooms.sfu.sdp.<cid>"
const sender = "<uid>"

type IcePayload = webrtc.ICECandidateInit
type SdpPayload = string

type SFUIceMessage = SocketMessage[IcePayload]
type SFUSdpuMessage = SocketMessage[SdpPayload]
