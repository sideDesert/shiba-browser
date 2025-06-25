package lib

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/pion/webrtc/v4"
)

type ApiError struct {
	Error string `json:"error"`
}

type SocketMessage[T any] struct {
	// Subject can be - chatroom.chat.<cid>, chatroom.sfu.<cid>, chatroom.signal.<cid>
	Subject string `json:"subject"`
	// Sender is always userId
	Sender string `json:"sender"`
	// Payload depends on different type of messages
	Payload T `json:"payload"`
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
