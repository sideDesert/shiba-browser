package lib

import "github.com/pion/webrtc/v3"

/*
 * CLIENT SIDE TO SERVER SIDE SIGNALS
 */

const Prefix = "chatrooms.signal."

// 1. Init Indiv Calls  (signal.init-ic)
const Signal0 = "init-ic"
const SignalInitIndivCall = Signal0
const Subject0 = Prefix + Signal0

type Signal0Payload struct{}
type Signal0Message = SocketMessage[Signal0Payload]

// 2. Init Group Calls  (signal.init-gc)
const Signal1 = "init-gc"
const SignalInitGroupCall = Signal1
const Subject1 = Prefix + Signal1

type Signal1Payload struct {
	Sdp webrtc.SessionDescription `json:"sdp"`
}
type Signal1Message = SocketMessage[Signal1Payload]

// 3. Indiv Call Answer (signal.ans-ic)
const Signal2 = "ans-ic"
const SignalIndivCallAnswer = Signal2
const Subject2 = Prefix + Signal2

type Signal2Payload struct {
	Answer   string                    `json:"answer"` // "accept" or "decline"
	CallerId string                    `json:"caller_id"`
	Sdp      webrtc.SessionDescription `json:"sdp,omitempty"`
}
type Signal2Message = SocketMessage[Signal2Payload]

// 12. Ack Indiv Call (signal.ack-ic)
const Signal11 = "ack-ic"
const SignalAckIndivCall = Signal11
const Subject11 = Prefix + Signal11

type Signal11Payload struct {
	Sdp webrtc.SessionDescription `json:"sdp"`
}
type Signal11Message = SocketMessage[Signal11Payload]

// 4. Group Call Answer (signal.end-call)
const Signal3 = "end-call"
const SignalGroupCallAnswer = Signal3
const Subject3 = Prefix + Signal3

type Signal3Payload struct{}
type Signal3Message = SocketMessage[Signal3Payload]

// 5. Join Call (signal.ans-gc)
const Signal4 = "ans-gc"
const SignalJoinCall = Signal4
const Subject4 = Prefix + Signal4

type Signal4Payload struct {
	Answer string                    `json:"answer"` // "join" or "leave"
	Sdp    webrtc.SessionDescription `json:"sdp,omitempty"`
}
type Signal4Message = SocketMessage[Signal4Payload]

// 6. Start Stream (signal.start-stream)
const Signal5 = "start-stream"
const SignalStartStream = Signal5
const Subject5 = Prefix + Signal5

type Signal5Payload struct{}
type Signal5Message = SocketMessage[Signal5Payload]

// 8. Join Stream (signal.join-stream)
const Signal7 = "join-stream"
const SignalJoinStream = Signal7
const Subject7 = Prefix + Signal7

type Signal7Payload struct{}
type Signal7Message = SocketMessage[Signal7Payload]

// 9. Leave Stream (signal.leave-stream)
const Signal8 = "leave-stream"
const SignalLeaveStream = Signal8
const Subject8 = Prefix + Signal8

type Signal8Payload struct{}
type Signal8Message = SocketMessage[Signal8Payload]

// 10. End Stream (signal.end-stream)
const Signal9 = "end-stream"
const SignalEndStream = Signal9
const Subject9 = Prefix + Signal9

type Signal9Payload struct{}
type Signal9Message = SocketMessage[Signal9Payload]

/*
 * SERVER TO CLIENT SIGNALS
 */

// 11. Remote Action (signal.remote)
const Signal10 = "remote"
const SignalRemoteAction = Signal10
const Subject10 = Prefix + Signal10

type Signal10Payload struct {
	Action string `json:"action"`
	Value  string `json:"value"`
}
type Signal10Message = SocketMessage[Signal10Payload]

// Slices for all signals and subjects
var Signal = []string{
	Signal0,
	Signal1,
	Signal2,
	Signal3,
	Signal4,
	Signal5,
	Signal7,
	Signal8,
	Signal9,
	Signal10,
	Signal11,
}

var SignalSubject = []string{
	Subject0,
	Subject1,
	Subject2,
	Subject3,
	Subject4,
	Subject5,
	Subject7,
	Subject8,
	Subject9,
	Subject10,
	Subject11,
}

var SignalMessageTypes = map[string]func() interface{}{
	Signal0:  func() any { return &Signal0Message{} },
	Signal1:  func() any { return &Signal1Message{} },
	Signal2:  func() any { return &Signal2Message{} },
	Signal3:  func() any { return &Signal3Message{} },
	Signal4:  func() any { return &Signal4Message{} },
	Signal5:  func() any { return &Signal5Message{} },
	Signal7:  func() any { return &Signal7Message{} },
	Signal8:  func() any { return &Signal8Message{} },
	Signal9:  func() any { return &Signal9Message{} },
	Signal10: func() any { return &Signal10Message{} },
	Signal11: func() any { return &Signal11Message{} },
}
