package controller

import (
	"sideDesert/shiba/internal/logger"
	"sideDesert/shiba/internal/server/lib"

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
	peerA := sfu.NewPeer(c.sfu)
	// For the subscriber
	SetupPeer(peerA, c.msgChannel, userId)
	peerA.Join(chatroomId, userId)
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
		peerB := sfu.NewPeer(c.sfu)
		SetupPeer(peerB, c.msgChannel, userId)
		peerB.Join(chatroomId, userId)
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

// This is the old version of stopping stream
// if msgType == "stop-stream" {
// 	log.Info("Stopping Stream")

// 	err := browserManager.Pipeline.SetState(gst.StateNull)
// 	if err != nil {
// 		log.Error("Error stopping stream[Pipeline.SetState(gst.StateNull)]", err)
// 	}

// 	userIds, err := c.s.Store.GetUsersByChatroomId(c.s.Ctx, chatroomId)
// 	if err != nil {
// 		log.Error("Error getting users by chatroom id:", err)
// 	}

// 	c.mu.Lock()
// 	for _, conn := range c.conns {
// 		if lib.Contains(userIds, conn.UserId) {
// 			conn.StreamConfig.PeerConnection.Close()
// 		}
// 	}
// 	c.mu.Unlock()

// 	c.chatroomCtx[chatroomId].cancel()
// 	delete(c.chatroomCtx, chatroomId)
// 	log.Info("Stream Ended")
// 	break
// }

/* */

// if strings.HasPrefix(socketMsg.Subject, "") {
// 			// The message form will be - stream.[type].[chatroomId]
// 			// DEBUG
// 			// log.Info("Message received in stream socket subject")

// 			sp := strings.Split(socketMsg.Subject, ".")
// 			if len(sp) != 3 {
// 				log.Error("Error in msg[subject] length:(not 3)")
// 				continue
// 			}
// 			msgType := sp[1]
// 			chatroomId := sp[2]
// 			userId := socketMsg.Sender
// 			if userId == "" {
// 				log.Error("No senderId provided in message")
// 				log.Error("Sender", string(msg))
// 				continue
// 			}

// 			if msgType == "answer" {
// 				payloadMap, ok := socketMsg.Payload.(map[string]any)
// 				if !ok {
// 					log.Error("Payload is not a valid map[string]interface{}")
// 					continue
// 				}

// 				jsonBytes, err := json.Marshal(payloadMap)
// 				if err != nil {
// 					log.Error("Failed to marshal payload map to JSON:", err)
// 					continue
// 				}

// 				var desc webrtc.SessionDescription
// 				err = json.Unmarshal(jsonBytes, &desc)
// 				if err != nil {
// 					log.Error("Failed to unmarshal JSON to SessionDescription:", err)
// 					continue
// 				}

// 				c.mu.Lock()
// 				connVal, ok := c.conns[conn]
// 				c.mu.Unlock()
// 				if !ok {
// 					log.Error("Error: reading connVal c.conns[conn]:")
// 					continue
// 				}
// 				err = connVal.StreamConfig.PeerConnection.SetRemoteDescription(desc)
// 				if err != nil {
// 					log.Error("Error setting remote description:", err)
// 					continue
// 				}
// 				log.Info("Remote description set for", userId, ":", chatroomId)
// 				log.Info("Webrtc Connection Established with", userId, ":", chatroomId)
// 			}
// 			if msgType == "ice" {
// 				payloadMap, ok := socketMsg.Payload.(map[string]any)
// 				if !ok {
// 					log.Error("Ice Candidate payload is not of type map[string]any")
// 					continue
// 				}

// 				jsonBytes, err := json.Marshal(payloadMap)
// 				if err != nil {
// 					log.Error("Failed to marshal payload map to JSON:", err)
// 					continue
// 				}

// 				var _candidate webrtc.ICECandidateInit
// 				err = json.Unmarshal(jsonBytes, &_candidate)
// 				if err != nil {
// 					log.Error("Failed to unmarshal JSON to ICECandidateInit:", err)
// 					continue
// 				}

// 				c.mu.Lock()
// 				c.conns[conn].StreamConfig.PeerConnection.AddICECandidate(_candidate)
// 				c.mu.Unlock()
// 				// log.Info("ICE candidate added for", userId, ":", chatroomId)
// 			}

// 			if msgType == "disconnected" {
// 				log.Info("User", userId, "disconnected from", chatroomId)
// 				break
// 			}

// 		}
