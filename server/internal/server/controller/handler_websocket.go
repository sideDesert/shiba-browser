package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sideDesert/shiba/internal/logger"
	"sideDesert/shiba/internal/server/lib"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
)

func (c *Controller) handleWebsocket(w http.ResponseWriter, r *http.Request) error {
	// Extract user ID from request context
	log := logger.NewLogger(logger.File, "handleWebsocket")

	userId, ok := r.Context().Value("userId").(string)
	chatroomId := r.URL.Query().Get("cid")

	if chatroomId == "" {
		log.Info("Nothing provided as chatroomId")
		return fmt.Errorf("invalid chatroomId")
	}
	log.Info("Chatroom ID:", chatroomId)

	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return fmt.Errorf("invalid user ID")
	}

	// Fetch user chat rooms
	userChatrooms, err := c.s.GetUserChatRooms(userId)
	if err != nil {
		log.Error("GetUserChatRooms", err)
		return err
	}

	// Upgrade HTTP to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrader", err)
		return err
	}

	c.mu.Lock()
	chatroomCtx, ok := c.chatroomCtx[chatroomId]
	c.mu.Unlock()

	if !ok {
		c.mu.Lock()
		c.chatroomCtx[chatroomId] = c.NewChatroomCtx(context.Background(), chatroomId, -1)
		c.mu.Lock()
	}

	c.mu.Lock()
	c.chatroomCtx[chatroomId].Users[userId] = conn
	chatroomCtx = c.chatroomCtx[chatroomId]
	c.mu.Unlock()

	// Store client connection - this creates the peer connection as well
	connsVal, err := lib.NewConnMap(userId)
	if err != nil {
		log.Error("upgrader", err)
	} else {
		c.conns[conn] = connsVal
	}

	userTag := strings.Split(userId, "-")[0]
	log.Success("Established WebSocket connection with", userTag)

	// CLEANUP : DO NOT REMOVE
	defer func() {
		log.Info("Starting cleanup for", userTag)

		c.mu.Lock()
		connVal, ok := c.conns[conn]
		if ok {
			for _, sub := range connVal.Subscriptions {
				sub.Unsubscribe()
			}
			connVal.StreamConfig.PeerConnection.Close()

			delete(c.conns, conn)
			delete(c.chatroomCtx[chatroomId].Users, userId)

			if len(c.chatroomCtx[chatroomId].Users) == 0 {
				log.Info("Chatroom will be removed - ", chatroomId)
			}
		}
		c.mu.Unlock()

		conn.Close()
		log.Info("Connection closed with", userTag)
	}()

	log.Info("Total active connections:", len(c.conns))

	// Subscribe to chat rooms
	// This is the part of code taking care of SENDING MESSAGES TO ALL PEERS
	currentConn := c.conns[conn]

	for _, room := range userChatrooms {
		chatroomId := room.Id
		sub, err := c.nats.Subscribe("chatroom.chat."+chatroomId, func(msg *nats.Msg) {
			log.Info("Received Message:", string(msg.Data))
			err := conn.WriteMessage(websocket.TextMessage, msg.Data)

			if err != nil {
				log.Error("writingWs[chatroom.chat.*]", err)
				// Remove connection from cache safely
				c.mu.Lock()
				delete(c.conns, conn)
				c.mu.Unlock()

				conn.Close()
				return
			}
		})
		if err != nil {
			log.Error("NATSubscription[chatrooms.chat.*]:", err)
			continue
		}

		// Signalling
		currentConn.Subscriptions = append(currentConn.Subscriptions, sub)
		sub, err = c.nats.Subscribe("chatroom.signal.*."+chatroomId, func(msg *nats.Msg) {
			log.Info("Received Signal:", string(msg.Data))
			err := conn.WriteMessage(websocket.TextMessage, msg.Data)

			if err != nil {
				log.Error("Writing WebSocket[chatroom.signal.*] message:", err)
				return
			}
		})

		if err != nil {
			log.Error("NATS[chatrooms.signal.*]:", err)
			continue
		}

		// Webrtc Signalling
		sub, err = c.nats.Subscribe("chatroom.sfu.*."+chatroomId, func(msg *nats.Msg) {
			err := conn.WriteMessage(websocket.TextMessage, msg.Data)
			if err != nil {
				log.Error("writing WebSocket[webrtc.*] Webrtc Message:", err)
				// Remove connection from cache safely
				c.mu.Lock()
				delete(c.conns, conn)
				c.mu.Unlock()

				conn.Close()
				return
			}
		})
		if err != nil {
			log.Error("Error subscribing to NATS[webrtc.*]:", err)
			continue
		}
		currentConn.Subscriptions = append(currentConn.Subscriptions, sub)
	}

	// Listen for messages
	// This is the part of code taking care of HANDLING MESSAGES received from the websocket

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Error("Client disconnected or read error:", err)
			continue
		}

		// Parse message
		var initMsgObj lib.SocketMessage[any]
		err = json.Unmarshal(msg, &initMsgObj)
		if err != nil {
			log.Error("Error[json.Unmarshal(initMsgObj)]:", err)
		}
		fmt.Println("📥 Received:", string(initMsgObj.Subject))

		if err != nil {
			log.Error("Error Unmarshalling initMsgObj: ", err)
		}

		// TODO: Take care of this
		// Signal can be of alot of types
		if strings.HasPrefix(initMsgObj.Subject, "chatrooms.signal") {
			// Init Indiv Calls  (chatrooms.signal.init-ic.<cid>)
			// Init Group Calls  (chatrooms.signal.init-gc.<cid>)
			// Indiv Call Answer (chatrooms.signal.ans-ic.<cid>)
			// End Call          (chatrooms.signal.end-call.<cid>)
			// Join Call         (chatrooms.signal.join-call.<cid>)
			// Leave Call        (chatrooms.signal.leave-call.<cid>)
			// Start Stream      (chatrooms.signal.start-stream.<cid>)
			// Join Stream       (chatrooms.signal.join-stream.<cid>)
			// Leave Stream      (chatrooms.signal.leave-stream.<cid>)
			// End Stream        (chatrooms.signal.end-stream.<cid>)
			// Remote Actions    (chatrooms.signal.change-remote.<cid>)

			s := strings.Split(initMsgObj.Subject, ".")
			if len(s) != 3 {
				log.Error("Error in msg[signal] type:")
				continue
			}
			//TODO: Handle all cases of signalling
			chatroomId := s[3]
			signalType := s[2]

			switch signalType {
			case "init-ic":
				// Forward to nats
				continue
			case "init-gc":
				// Forward to nats
				continue
			case "join-ic":

			case "ans-ic":
				if userId == initMsgObj.Sender {
					// Join the call
					var sigAns lib.Signal2Payload
					err := json.Unmarshal(msg, &sigAns)
					if err != nil {
						log.Error("Error in SignalHandler[ans-ic]->", err)
						continue
					}
					if sigAns.Answer == "accept" {
						sdpString := sigAns.Sdp
						if sdpString == "" {
							log.Error("Error in SignalHandler[ans-ic].accecpt->", fmt.Errorf("No SDP sent"))
							continue
						}
						var sdp webrtc.SessionDescription
						err := json.Unmarshal([]byte(sdpString), &sdp)
						if err != nil {
							log.Error("Error in SignalHandler[ans-ic.json-parse]->", err)
							continue
						}

						sesh := (*c.chatroomCtx[chatroomId].Session)
						localSession := sesh.(*sfu.SessionLocal)
						peer := sfu.NewPeer(c.sfu)
						peer.Join(chatroomId, userId)

						answer, _ := peer.Answer(webrtc.SessionDescription{
							Type: webrtc.SDPTypeOffer,
							SDP:  sdp.SDP,
						})

						localSession.AddPeer(peer)

					}
					continue
				}
			case "ans-gc":
			case "end-call":
			case "start-stream":
			case "leave-stream":
			case "end-stream":
			case "remote":
			}

			c.nats.Publish("chatrooms."+chatroomId, msg)
		}

		if strings.HasPrefix(initMsgObj.Subject, "chatrooms.chat") {
			msgObj := lib.ChatMessage{}
			s := strings.Split(initMsgObj.Subject, ".")

			if len(s) < 2 {
				log.Error("Error chat message is not correct format, got:", string(msg))
				continue
			}

			chatroomId := s[1]
			json.Unmarshal(msg, &msgObj)

			if !ok {
				log.Error("Failed to assert Payload as dto.ChatMessagePayload")
				log.Error("Payload", string(msg))
				continue
			}

			c.nats.Publish("chatrooms.chat."+chatroomId, msg)

			// Store message
			if err := c.s.StoreChatMessage(initMsgObj.Sender, chatroomId, msgObj.Payload); err != nil {
				log.Error("Error storing chat message:", err)
				continue
			}
		}

		// Type - chatrooms.sfu.[offer].[id]
		if strings.HasPrefix(initMsgObj.Subject, "chatrooms.sfu.") {
			s := strings.Split(initMsgObj.Subject, ".")
			if len(s) != 3 {
				log.Error("Error in msg[webrtc] type:")
				continue
			}
			chatroomId := s[2]
			c.nats.Publish("chatrooms."+chatroomId, msg)
		}

		if strings.HasPrefix(initMsgObj.Subject, "") {
			// The message form will be - stream.[type].[chatroomId]
			// DEBUG
			// log.Info("Message received in stream socket subject")

			sp := strings.Split(initMsgObj.Subject, ".")
			if len(sp) != 3 {
				log.Error("Error in msg[subject] length:(not 3)")
				continue
			}
			msgType := sp[1]
			chatroomId := sp[2]
			userId := initMsgObj.Sender
			if userId == "" {
				log.Error("No senderId provided in message")
				log.Error("Sender", string(msg))
				continue
			}

			if msgType == "answer" {
				payloadMap, ok := initMsgObj.Payload.(map[string]any)
				if !ok {
					log.Error("Payload is not a valid map[string]interface{}")
					continue
				}

				jsonBytes, err := json.Marshal(payloadMap)
				if err != nil {
					log.Error("Failed to marshal payload map to JSON:", err)
					continue
				}

				var desc webrtc.SessionDescription
				err = json.Unmarshal(jsonBytes, &desc)
				if err != nil {
					log.Error("Failed to unmarshal JSON to SessionDescription:", err)
					continue
				}

				c.mu.Lock()
				connVal, ok := c.conns[conn]
				c.mu.Unlock()
				if !ok {
					log.Error("Error: reading connVal c.conns[conn]:")
					continue
				}
				err = connVal.StreamConfig.PeerConnection.SetRemoteDescription(desc)
				if err != nil {
					log.Error("Error setting remote description:", err)
					continue
				}
				log.Info("Remote description set for", userId, ":", chatroomId)
				log.Info("Webrtc Connection Established with", userId, ":", chatroomId)
			}
			if msgType == "ice" {
				payloadMap, ok := initMsgObj.Payload.(map[string]any)
				if !ok {
					log.Error("Ice Candidate payload is not of type map[string]any")
					continue
				}

				jsonBytes, err := json.Marshal(payloadMap)
				if err != nil {
					log.Error("Failed to marshal payload map to JSON:", err)
					continue
				}

				var _candidate webrtc.ICECandidateInit
				err = json.Unmarshal(jsonBytes, &_candidate)
				if err != nil {
					log.Error("Failed to unmarshal JSON to ICECandidateInit:", err)
					continue
				}

				c.mu.Lock()
				c.conns[conn].StreamConfig.PeerConnection.AddICECandidate(_candidate)
				c.mu.Unlock()
				// log.Info("ICE candidate added for", userId, ":", chatroomId)
			}

			if msgType == "disconnected" {
				log.Info("User", userId, "disconnected from", chatroomId)
				break
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
		}
	}

	log.Info("Client Disconnected")
	return nil
}
