package lib

import (
	"log"

	"github.com/pion/webrtc/v4"
)

type PeerConnection struct {
	pc    *webrtc.PeerConnection
	video *webrtc.TrackLocalStaticSample
	audio *webrtc.TrackLocalStaticSample
}

func NewRTCPeerConnection(streamId string) (*PeerConnection, error) {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Println("Error in NewParticipant[PeerConnection]:", err)
		return nil, err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "v-"+streamId)
	if err != nil {
		log.Println("Error in NewParticipant[VideoTrack]:", err)
		return nil, err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "a-"+streamId)
	if err != nil {
		log.Println("Error in NewParticipant[AudioTrack]:", err)
		return nil, err
	}
	_, err = pc.AddTransceiverFromTrack(videoTrack, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})
	if err != nil {
		log.Println("Error adding video transceiver:", err)
		return nil, err
	}

	_, err = pc.AddTransceiverFromTrack(audioTrack, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})
	if err != nil {
		log.Println("Error adding audio transceiver:", err)
		return nil, err
	}

	return &PeerConnection{pc: pc, video: videoTrack, audio: audioTrack}, nil
}
