package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sideDesert/shiba/internal/server/dto"
	"sideDesert/shiba/internal/server/lib"
	"strings"

	"github.com/go-gst/go-gst/gst"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/pion/webrtc/v4"
)

func (c *Controller) handleWebsocket(w http.ResponseWriter, r *http.Request) error {
	// Extract user ID from request context
	userId, ok := r.Context().Value("userId").(string)
	chatroomId := r.URL.Query().Get("cid")

	if chatroomId == "" {
		log.Println("Nothing provided as chatroomId")
		return fmt.Errorf("invalid chatroomId")
	}

	log.Println("Chatroom ID:", chatroomId)

	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return fmt.Errorf("invalid user ID")
	}

	// Fetch user chat rooms
	chatrooms, err := c.s.GetUserChatRooms(userId)
	if err != nil {
		log.Println("Error in handleChatWebsocket[GetUserChatRooms]:", err)
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
		log.Println("Error in handleChatWebsocket[upgrader]:", err)
		return err
	}

	// Store client connection - this creates the peer connection as well
	connsVal, err := lib.NewConnMap(userId)
	if err != nil {
		log.Println("Error in handleChatWebsocket[upgrader]", err)
	}
	c.conns[conn] = connsVal

	userTag := strings.Split(userId, "-")[0]
	log.Println("✅ Established WebSocket connection with", userTag)

	// CLEANUP : DO NOT REMOVE
	defer func() {
		log.Println("🚀 Starting cleanup for", userTag)

		c.mu.Lock()
		delete(c.conns, conn)
		c.mu.Unlock()

		conn.Close()
		log.Println("❌ Connection closed with", userTag)
	}()

	if err != nil {
		log.Println("❌ Error creating participant:", err)
		return fmt.Errorf("Error creating participant")
	}

	log.Println("🫂 Total active connections:", len(c.conns))

	// Subscribe to chat rooms
	for _, room := range chatrooms {
		chatroomId := room.Id
		_, err := c.nats.Subscribe("chatrooms."+chatroomId, func(msg *nats.Msg) {
			log.Println("Received Message:", string(msg.Data))
			err := conn.WriteMessage(websocket.TextMessage, msg.Data)

			if err != nil {
				log.Println("❌ Error writing WebSocket message:", err)
				// Remove connection from cache safely
				c.mu.Lock()
				delete(c.conns, conn)
				c.mu.Unlock()

				conn.Close()
				return
			}
		})

		if err != nil {
			log.Println("❌ Error subscribing to NATS[chatrooms.*]:", err)
			continue
		}

		_, err = c.nats.Subscribe("webrtc.*."+chatroomId, func(msg *nats.Msg) {
			err := conn.WriteMessage(websocket.TextMessage, msg.Data)
			if err != nil {
				log.Println("❌ Error writing WebSocket Webrtc Message:", err)
				// Remove connection from cache safely
				c.mu.Lock()
				delete(c.conns, conn)
				c.mu.Unlock()

				conn.Close()
				return
			}
		})
		if err != nil {
			log.Println("❌ Error subscribing to NATS[webrtc.*]:", err)
			continue
		}
	}

	// Listen for messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("❌ Client disconnected or read error:", err)
			break
		}

		// Parse message
		var initMsgObj dto.Message[any]
		err = json.Unmarshal(msg, &initMsgObj)
		fmt.Println("📥 Received:", string(initMsgObj.Subject))

		if err != nil {
			log.Println("Error Unmarshalling initMsgObj: ", err)
		}

		if strings.HasPrefix(initMsgObj.Subject, "chat") {
			msgObj := dto.Message[dto.ChatMessagePayload]{}
			s := strings.Split(initMsgObj.Subject, ".")

			if len(s) < 2 {
				log.Println("❌ Error chat message is not correct format, got:", string(msg))
			}

			chatroomId := s[1]
			json.Unmarshal(msg, &msgObj)

			if !ok {
				log.Println("❌ Failed to assert Payload as dto.ChatMessagePayload")
				log.Println("Payload", string(msg))
				return fmt.Errorf("invalid payload type")
			}

			c.nats.Publish("chatrooms."+chatroomId, msg)

			// Store message
			if err := c.s.StoreChatMessage(initMsgObj.Sender, chatroomId, msgObj.Payload); err != nil {
				log.Println("❌ Error storing chat message:", err)
				break
			}
		}

		// Type - webrtc.[offer].[id]
		if strings.HasPrefix(initMsgObj.Subject, "webrtc") {
			s := strings.Split(initMsgObj.Subject, ".")
			if len(s) != 3 {
				log.Println("❌ Error in msg[webrtc] type:")
				break
			}
			chatroomId := s[2]
			c.nats.Publish("chatrooms."+chatroomId, msg)
		}

		if strings.HasPrefix(initMsgObj.Subject, "stream") {
			// The message form will be - stream.[type].[chatroomId]
			// DEBUG
			// log.Println("✅Message received in stream socket subject")

			sp := strings.Split(initMsgObj.Subject, ".")
			if len(sp) != 3 {
				log.Println("❌ Error in msg[subject] length:(not 3)")
				break
			}
			msgType := sp[1]
			chatroomId := sp[2]
			userId := initMsgObj.Sender
			if userId == "" {
				log.Println("🔴 No senderId provided in message")
				log.Println("Sender", string(msg))
				break
			}

			if msgType == "answer" {
				payloadMap, ok := initMsgObj.Payload.(map[string]any)
				if !ok {
					log.Println("🔴 Payload is not a valid map[string]interface{}")
					break
				}

				jsonBytes, err := json.Marshal(payloadMap)
				if err != nil {
					log.Println("🔴 Failed to marshal payload map to JSON:", err)
					break
				}

				var desc webrtc.SessionDescription
				err = json.Unmarshal(jsonBytes, &desc)
				if err != nil {
					log.Println("🔴Failed to unmarshal JSON to SessionDescription:", err)
					break
				}

				c.mu.Lock()
				err = c.conns[conn].StreamConfig.PeerConnection.SetRemoteDescription(desc)
				c.mu.Unlock()
				if err != nil {
					log.Println("🔴Error setting remote description:", err)
					break
				}
				log.Println("✅ Remote description set for", userId, ":", chatroomId)
				log.Println("🔥 Webrtc Connection Established with", userId, ":", chatroomId)
			}
			if msgType == "ice" {
				payloadMap, ok := initMsgObj.Payload.(map[string]any)
				if !ok {
					log.Println("Ice Candidate payload is not of type map[string]any")
					break
				}

				jsonBytes, err := json.Marshal(payloadMap)
				if err != nil {
					log.Println("🔴 Failed to marshal payload map to JSON:", err)
					break
				}

				var _candidate webrtc.ICECandidateInit
				err = json.Unmarshal(jsonBytes, &_candidate)
				if err != nil {
					log.Println("🔴Failed to unmarshal JSON to ICECandidateInit:", err)
					break
				}

				c.mu.Lock()
				c.conns[conn].StreamConfig.PeerConnection.AddICECandidate(_candidate)
				c.mu.Unlock()
				// log.Println("✅ ICE candidate added for", userId, ":", chatroomId)
			}

			if msgType == "disconnected" {
				log.Println("⭕User", userId, "disconnected from", chatroomId)
				c.mu.Lock()
				c.conns[conn].StreamConfig.PeerConnection.Close()
				delete(c.conns, conn)
				c.mu.Unlock()
			}

			if msgType == "stop-stream" {
				log.Println("⛔ Stopping Stream")

				err := c.browserManager.Pipeline.SetState(gst.StateNull)
				if err != nil {
					log.Println("Error stopping stream[Pipeline.SetState(gst.StateNull)]", err)
					return err
				}

				userIds, err := c.s.Store.GetUsersByChatroomId(c.s.Ctx, chatroomId)
				if err != nil {
					log.Println("Error getting users by chatroom id:", err)
					return err
				}

				for _, conn := range c.conns {
					if lib.Contains(userIds, conn.UserId) {
						conn.StreamConfig.PeerConnection.Close()
					}
				}

				c.chatroomCtx[chatroomId].cancel()
				delete(c.chatroomCtx, chatroomId)
				log.Println("⛔👍Stream Ended")
			}
		}
	}

	log.Println("👋 Client Disconnected")
	return nil
}
