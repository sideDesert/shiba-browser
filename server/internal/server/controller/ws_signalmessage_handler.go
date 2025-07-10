package controller

import (
	"errors"
	"sideDesert/shiba/internal/logger"
	"sideDesert/shiba/internal/server/lib"
	"strings"

	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
)

func SetupPeer(peer *sfu.PeerLocal, msgChannel *MsgChannel, userId string) {
	log := logger.NewLogger(logger.Console, "SetupPeer")
	peer.OnOffer = func(sdp *webrtc.SessionDescription) {
		sfuPayload := lib.SdpPayload{
			Type:  "sub",
			Offer: sdp,
		}
		serverOffer, err := lib.NewSFUSocketMessage("server.sfu", lib.ServerOffer+"."+userId, sfuPayload)
		if err != nil {
			log.Error("peerB.OnOffer.NewSFUSocketMessage", err)
			return
		}
		err = msgChannel.Publish(&serverOffer)
		if err != nil {
			log.Error("peerB.msgChannel.Publish", err)
			return
		}
	}

	peer.Publisher().OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		sfuPayload := lib.IcePayload{
			Type:    "pub",
			Trickle: candidate.ToJSON(),
		}
		serverTrickle, err := lib.NewSFUSocketMessage("server.sfu", lib.ServerTrickle+"."+userId, sfuPayload)
		if err != nil {
			log.Error("peerB.Publisher.OnICECandidate.NewSFUSocketMessage", err)
			return
		}
		msgChannel.Publish(&serverTrickle)
	})

	peer.Subscriber().OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		sfuPayload := lib.IcePayload{
			Type:    "sub",
			Trickle: candidate.ToJSON(),
		}
		serverTrickle, err := lib.NewSFUSocketMessage("server.sfu", lib.ServerTrickle+"."+userId, sfuPayload)
		if err != nil {
			log.Error("peerB.Subscriber.OnICECandidate.NewSFUSocketMessage", err)
			return
		}
		msgChannel.Publish(&serverTrickle)
	})
}

// A3
func (c *Controller) AckMsgHandler(v lib.Signal11Message, chatroomId string, userId string) error {
	log := logger.NewLogger(logger.Console, "AckMsgHandler")
	sdp := v.Payload.Sdp
	if !strings.Contains(sdp.SDP, "m=video") {
		log.Error("SDP does not contain a video m-line, rejecting")
		return errors.New("invalid SDP: no video")
	}
	peerA := sfu.NewPeer(c.sfu)
	// For the subscriber
	err := peerA.Join(chatroomId, userId)
	if err != nil {
		log.Error("peerA.join", err)
	}
	SetupPeer(peerA, c.msgChannel, userId)
	answer, err := peerA.Answer(sdp)
	if err != nil {
		log.Error("peerA.Answer", err)
		return err
	}

	sfuPayload := lib.SdpPayload{
		Offer: answer,
		Type:  "pub",
	}
	serverAnswer, err := lib.NewSFUSocketMessage("server.sfu", lib.ServerAnswer+"."+userId, sfuPayload)
	if err != nil {
		log.Error("NewSFUSocketMessage.server-ans", err)
	}
	err = c.msgChannel.Publish(&serverAnswer)
	if err != nil {
		log.Error("c.msgChannel.Publish", err)
	}
	c.GetChatroomCtx(chatroomId).AddCallParticipant(userId, peerA)
	return nil
}

// B1
func (c *Controller) AnsIcMsgHandler(v lib.Signal2Message, chatroomId string, userId string) error {
	log := logger.NewLogger(logger.Console, "AnsIcMsgHandler")
	payload := v.Payload
	if payload.Answer == "decline" {
		c.msgChannel.Publish(&v)
	}

	if payload.Answer == "accept" {
		callerId := payload.CallerId
		sdp := payload.Sdp
		if !strings.Contains(sdp.SDP, "m=video") {
			log.Error("SDP does not contain a video m-line, rejecting")
			return errors.New("invalid SDP: no video")
		}
		peerB := sfu.NewPeer(c.sfu)
		err := peerB.Join(chatroomId, userId)
		if err != nil {
			log.Error("peerB.Join", err)
		}
		SetupPeer(peerB, c.msgChannel, userId)
		answer, err := peerB.Answer(sdp)
		if err != nil {
			log.Error("peerB.Answer", err)
			return err
		}

		sfuPayload := lib.SdpPayload{
			Offer: answer,
			Type:  "pub",
		}
		serverAnswer, err := lib.NewSFUSocketMessage("server.sfu", lib.ServerAnswer+"."+userId, sfuPayload)
		if err != nil {
			log.Error("NewSFUSocketMessage.server-answer", err)
			return err
		}
		c.msgChannel.Publish(&serverAnswer)

		request, err := lib.NewSFUSocketMessage("server.sfu", callerId, lib.RequestPayload{})
		if err != nil {
			log.Error("NewSFUSocketMessage.callerId", err)
			return err
		}
		c.msgChannel.Publish(&request)

		c.GetChatroomCtx(chatroomId).AddCallParticipant(userId, peerB)
		c.GetChatroomCtx(chatroomId).SetOnCall(true)

	}
	return nil
}
