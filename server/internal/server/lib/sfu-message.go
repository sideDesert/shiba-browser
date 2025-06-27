package lib

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pion/webrtc/v3"
)

// _type => server-answer, server-offer, server-request
const (
	ServerAnswer  = "server-answer"
	ServerOffer   = "server-offer"
	ServerRequest = "server-request"
	ServerTrickle = "server-trickle"
	ClientTrickle = "client-trickle"
)
const iceSubject = "chatrooms.sfu.ice.<sigcode>.<uid>"
const sdpSubject = "chatrooms.sfu.sdp.<sigcode>.<uid>"
const sender = "<uid>"

type SdpPayload struct {
	Offer *webrtc.SessionDescription `json:"offer"`
	Type  string                     `json:"type"`
}
type RequestPayload = struct{}

type IcePayload struct {
	Trickle webrtc.ICECandidateInit `json:"trickle"`
	Type    string                  `json:"type"`
}

type SFUIceMessage = SocketMessage[IcePayload]
type SFUSdpMessage = SocketMessage[SdpPayload]
type SFUSdpReqmessage = SocketMessage[RequestPayload]

func (msg *SocketMessage[IcePayload]) UserId() (string, error) {
	split := strings.Split(msg.Subject, ".")
	if len(split) != 5 {
		return "", fmt.Errorf("msg subject is not length 4, got length %d", len(split))
	}
	return split[len(split)-1], nil

}

func (msg *SocketMessage[IcePayload]) SignalCode() (string, error) {
	split := strings.Split(msg.Subject, ".")
	if len(split) != 5 {
		return "", fmt.Errorf("msg subject is not length 4, got length %d", len(split))
	}
	if split[3] == ServerAnswer {
		return ServerAnswer, nil
	}
	if split[3] == ServerOffer {
		return ServerOffer, nil
	}
	if split[3] == ServerRequest {
		return ServerRequest, nil
	}
	if split[3] == ServerTrickle {
		return ServerTrickle, nil
	}
	if split[3] == ClientTrickle {
		return ClientTrickle, nil
	}
	return "", fmt.Errorf("Invalid Signal Code")
}

func NewSFUSocketMessage[T IcePayload | SdpPayload | RequestPayload](sender string, _subject string, payload T) (SocketMessage[T], error) {
	var subject string

	switch any(payload).(type) {
	case IcePayload:
		subject = "chatrooms.sfu.ice." + _subject
	case SdpPayload:
		subject = "chatrooms.sfu.sdp." + _subject
	case RequestPayload:
		subject = "chatrooms.sfu.sdp.server-request." + _subject
	default:
		return SocketMessage[T]{}, errors.New("unsupported payload type")
	}

	msg := SocketMessage[T]{
		Subject: subject,
		Sender:  sender,
		Payload: payload,
	}

	return msg, nil
}
