package lib

/*
 * CLIENT SIDE TO SERVER SIDE SIGNALS
 */

const Prefix = "chatrooms.signal."

// 1. Init Indiv Calls  (signal.init-ic.<cid>)
const Signal0 = "init-ic.<cid>"
const Subject0 = Prefix + Signal0

type Signal0Payload struct{}
type Signal0Message = SocketMessage[Signal0Payload]

// 2. Init Group Calls  (signal.init-gc.<cid>)
const Signal1 = "init-gc.<cid>"
const Subject1 = Prefix + Signal1

type Signal1Payload struct{}
type Signal1Message = SocketMessage[Signal1Payload]

// 3. Indiv Call Answer (signal.ans-ic.<cid>)
const Signal2 = "ans-ic.<cid>"
const Subject2 = Prefix + Signal2

type Signal2Payload struct {
	Answer string `json:"answer"` // "accept" or "decline"
	Sdp    string `json:"sdp,omitempty"`
}
type Signal2Message = SocketMessage[Signal2Payload]

// 12. Ack Indiv Call (signal.ack-ic.<cid>)
const Signal11 = "ack-ic.<cid>"
const Subject11 = Prefix + Signal11

type Signal11Payload struct {
	Sdp string `json:"sdp"`
}
type Signal11Message = SocketMessage[Signal11Payload]

// 4. Group Call Answer (signal.end-call.<cid>)
const Signal3 = "end-call.<cid>"
const Subject3 = Prefix + Signal3

type Signal3Payload struct{}
type Signal3Message = SocketMessage[Signal3Payload]

// 5. Join Call (signal.ans-gc.<cid>)
const Signal4 = "ans-gc.<cid>"
const Subject4 = Prefix + Signal4

type Signal4Payload struct {
	Answer string `json:"answer"` // "join" or "leave"
	Sdp    string `json:"sdp,omitempty"`
}
type Signal4Message = SocketMessage[Signal4Payload]

// 6. Start Stream (signal.start-stream.<cid>)
const Signal5 = "start-stream.<cid>"
const Subject5 = Prefix + Signal5

type Signal5Payload struct{}
type Signal5Message = SocketMessage[Signal5Payload]

// 8. Join Stream (signal.join-stream.<cid>)
const Signal7 = "join-stream.<cid>"
const Subject7 = Prefix + Signal7

type Signal7Payload struct{}
type Signal7Message = SocketMessage[Signal7Payload]

// 9. Leave Stream (signal.leave-stream.<cid>)
const Signal8 = "leave-stream.<cid>"
const Subject8 = Prefix + Signal8

type Signal8Payload struct{}
type Signal8Message = SocketMessage[Signal8Payload]

// 10. End Stream (signal.end-stream.<cid>)
const Signal9 = "end-stream.<cid>"
const Subject9 = Prefix + Signal9

type Signal9Payload struct{}
type Signal9Message = SocketMessage[Signal9Payload]

/*
 * SERVER TO CLIENT SIGNALS
 */

// 11. Remote Action (signal.remote.<cid>)
const Signal10 = "remote.<cid>"
const Subject10 = Prefix + Signal10

type Signal10Payload struct {
	Action string `json:"action"`
	Value  string `json:"value"`
}
type Signal10Message = SocketMessage[Signal10Payload]

// Slices for all signals and subjects
var Signals = []string{
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

func Sigtable() map[string]int {
	table := make(map[string]int)
	for idx, curr := range SignalSubject {
		table[curr] = idx
	}
	return table
}
