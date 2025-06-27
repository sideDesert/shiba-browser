package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/pion/webrtc/v4"
)

type ApiError struct {
	Error string `json:"error"`
}

type Msg interface {
	Parse() (any, error)
	Subj() string
}

type SocketMessage[T any] struct {
	Subject string `json:"subject"`
	Sender  string `json:"sender"`
	Payload T      `json:"payload"`
}

func (msg *SocketMessage[any]) Subj() string {
	return msg.Subject
}

func (msg *SocketMessage[any]) Parse() (interface{}, error) {
	payloadBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	switch {
	case strings.HasPrefix(msg.Subject, "chatroom.chat"):
		var chatMsg ChatMessage
		if err := json.Unmarshal(payloadBytes, &chatMsg); err != nil {
			return nil, fmt.Errorf("unmarshal ChatMessage: %w", err)
		}
		return chatMsg, nil

	case strings.HasPrefix(msg.Subject, "chatroom.signal"):
		parts := strings.Split(msg.Subject, ".")
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid signal subject: %s", msg.Subject)
		}
		code := parts[2]

		constructor, ok := SignalMessageTypes[code]
		if !ok {
			return nil, fmt.Errorf("unknown signal code: %s", code)
		}

		instance := constructor()
		if err := json.Unmarshal(payloadBytes, instance); err != nil {
			return nil, fmt.Errorf("unmarshal signal message: %w", err)
		}
		return instance, nil

	case strings.HasPrefix(msg.Subject, "chatroom.sfu.ice"):
		var sfuMsg SFUIceMessage
		if err := json.Unmarshal(payloadBytes, &sfuMsg); err != nil {
			return nil, fmt.Errorf("unmarshal SFUIceMessage: %w", err)
		}
		if sfuMsg.Payload.Type != "pub" && sfuMsg.Payload.Type != "sub" {
			return nil, fmt.Errorf("sfuMsg type is invalid")
		}
		return sfuMsg, nil

	case strings.HasPrefix(msg.Subject, "chatroom.sfu.sdp"):
		var sfuMsg SFUSdpMessage
		if err := json.Unmarshal(payloadBytes, &sfuMsg); err != nil {
			return nil, fmt.Errorf("unmarshal SFUSdpMessage: %w", err)
		}
		if sfuMsg.Payload.Type != "pub" && sfuMsg.Payload.Type != "sub" {
			return nil, fmt.Errorf("sfuMsg type is invalid")
		}
		return sfuMsg, nil

	default:
		return nil, fmt.Errorf("unknown subject: %s", msg.Subject)
	}
}

type StreamConfig struct {
	PeerConnection *webrtc.PeerConnection         `json:"peer_connection"`
	IceCandidates  []*webrtc.ICECandidate         `json:"ice_candidates"`
	VideoTrack     *webrtc.TrackLocalStaticSample `json:"video_track"`
	AudioTrack     *webrtc.TrackLocalStaticSample `json:"audio_track"`
}

type ConnMap struct {
	UserId        string        `json:"user_id"`
	StreamConfig  *StreamConfig `json:"stream_config"`
	Chatrooms     []Chatroom    `json:"chatrooms"`
	Subscriptions []*nats.Subscription
}

func NewConnMap(userId string) (*ConnMap, error) {
	config, err := NewStreamConfig(userId)
	if err != nil {
		log.Println("Error in NewConnMap[StreamConfig]:", err)
		return nil, err
	}
	return &ConnMap{
		UserId:       userId,
		StreamConfig: config,
		Chatrooms:    []Chatroom{},
	}, nil
}

func NewStreamConfig(streamId string) (*StreamConfig, error) {
	peerConn, err := NewRTCPeerConnection(streamId)
	if err != nil {
		log.Println("Error in NewStreamConfig[PeerConnection]:", err)
		return nil, err
	}

	return &StreamConfig{
		PeerConnection: peerConn.pc,
		VideoTrack:     peerConn.video,
		AudioTrack:     peerConn.audio,
		IceCandidates:  make([]*webrtc.ICECandidate, 0),
	}, nil
}
