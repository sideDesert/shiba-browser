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
		log.Println("🚨Error in handleChatWebsocket[GetUserChatRooms]:", err)
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
		log.Println("🚨Error in handleChatWebsocket[upgrader]:", err)
		return err
	}

	// Store client connection - this creates the peer connection as well
	connsVal, err := lib.NewConnMap(userId)
	if err != nil {
		log.Println("🚨Error in handleChatWebsocket[upgrader]", err)
	} else {
		c.conns[conn] = connsVal
	}

	userTag := strings.Split(userId, "-")[0]
	log.Println("✅ Established WebSocket connection with", userTag)

	// CLEANUP : DO NOT REMOVE
	defer func() {
		log.Println("🚀 Starting cleanup for", userTag)

		c.mu.Lock()
		connVal, ok := c.conns[conn]
		if ok {
			for _, sub := range connVal.Subscriptions {
				sub.Unsubscribe()
			}
			connVal.StreamConfig.PeerConnection.Close()
			delete(c.conns, conn)
		}
		c.mu.Unlock()

		conn.Close()
		log.Println("❗Connection closed with", userTag)
	}()

	log.Println("🫂 Total active connections:", len(c.conns))

	// Subscribe to chat rooms
	currentConn := c.conns[conn]
	for _, room := range chatrooms {
		chatroomId := room.Id
		sub, err := c.nats.Subscribe("chatrooms."+chatroomId, func(msg *nats.Msg) {
			log.Println("Received Message:", string(msg.Data))
			err := conn.WriteMessage(websocket.TextMessage, msg.Data)

			if err != nil {
				log.Println("🚨 Error writing WebSocket message:", err)
				// Remove connection from cache safely
				c.mu.Lock()
				delete(c.conns, conn)
				c.mu.Unlock()

				conn.Close()
				return
			}
		})
		if err != nil {
			log.Println("🚨 Error subscribing to NATS[chatrooms.*]:", err)
			continue
		}
		currentConn.Subscriptions = append(currentConn.Subscriptions, sub)

		sub, err = c.nats.Subscribe("webrtc.*."+chatroomId, func(msg *nats.Msg) {
			err := conn.WriteMessage(websocket.TextMessage, msg.Data)
			if err != nil {
				log.Println("🚨 Error writing WebSocket Webrtc Message:", err)
				// Remove connection from cache safely
				c.mu.Lock()
				delete(c.conns, conn)
				c.mu.Unlock()

				conn.Close()
				return
			}
		})
		if err != nil {
			log.Println("🚨 Error subscribing to NATS[webrtc.*]:", err)
			continue
		}
		currentConn.Subscriptions = append(currentConn.Subscriptions, sub)
	}

	// Listen for messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("🚨 Client disconnected or read error:", err)
			continue
		}

		// Parse message
		var initMsgObj dto.Message[any]
		err = json.Unmarshal(msg, &initMsgObj)
		if err != nil {
			log.Println("🚨Error[json.Unmarshal(initMsgObj)]:", err)
		}
		fmt.Println("📥 Received:", string(initMsgObj.Subject))

		if err != nil {
			log.Println("🚨Error Unmarshalling initMsgObj: ", err)
		}

		if strings.HasPrefix(initMsgObj.Subject, "chat") {
			msgObj := dto.Message[dto.ChatMessagePayload]{}
			s := strings.Split(initMsgObj.Subject, ".")

			if len(s) < 2 {
				log.Println("🚨 Error chat message is not correct format, got:", string(msg))
				continue
			}

			chatroomId := s[1]
			json.Unmarshal(msg, &msgObj)

			if !ok {
				log.Println("🚨 Failed to assert Payload as dto.ChatMessagePayload")
				log.Println("Payload", string(msg))
				continue
			}

			c.nats.Publish("chatrooms."+chatroomId, msg)

			// Store message
			if err := c.s.StoreChatMessage(initMsgObj.Sender, chatroomId, msgObj.Payload); err != nil {
				log.Println("🚨 Error storing chat message:", err)
				continue
			}
		}

		// Type - webrtc.[offer].[id]
		if strings.HasPrefix(initMsgObj.Subject, "webrtc") {
			s := strings.Split(initMsgObj.Subject, ".")
			if len(s) != 3 {
				log.Println("❌ Error in msg[webrtc] type:")
				continue
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
				continue
			}
			msgType := sp[1]
			chatroomId := sp[2]
			userId := initMsgObj.Sender
			if userId == "" {
				log.Println("🔴 No senderId provided in message")
				log.Println("Sender", string(msg))
				continue
			}

			if msgType == "answer" {
				payloadMap, ok := initMsgObj.Payload.(map[string]any)
				if !ok {
					log.Println("🔴 Payload is not a valid map[string]interface{}")
					continue
				}

				jsonBytes, err := json.Marshal(payloadMap)
				if err != nil {
					log.Println("🔴 Failed to marshal payload map to JSON:", err)
					continue
				}

				var desc webrtc.SessionDescription
				err = json.Unmarshal(jsonBytes, &desc)
				if err != nil {
					log.Println("🔴Failed to unmarshal JSON to SessionDescription:", err)
					continue
				}

				c.mu.Lock()
				connVal, ok := c.conns[conn]
				c.mu.Unlock()
				if !ok {
					log.Println("🔴Error: reding connVal c.conns[conn]:")
					continue
				}
				err = connVal.StreamConfig.PeerConnection.SetRemoteDescription(desc)
				if err != nil {
					log.Println("🔴Error setting remote description:", err)
					continue
				}
				log.Println("✅ Remote description set for", userId, ":", chatroomId)
				log.Println("🔥 Webrtc Connection Established with", userId, ":", chatroomId)
			}
			if msgType == "ice" {
				payloadMap, ok := initMsgObj.Payload.(map[string]any)
				if !ok {
					log.Println("Ice Candidate payload is not of type map[string]any")
					continue
				}

				jsonBytes, err := json.Marshal(payloadMap)
				if err != nil {
					log.Println("🔴 Failed to marshal payload map to JSON:", err)
					continue
				}

				var _candidate webrtc.ICECandidateInit
				err = json.Unmarshal(jsonBytes, &_candidate)
				if err != nil {
					log.Println("🔴Failed to unmarshal JSON to ICECandidateInit:", err)
					continue
				}

				c.mu.Lock()
				c.conns[conn].StreamConfig.PeerConnection.AddICECandidate(_candidate)
				c.mu.Unlock()
				// log.Println("✅ ICE candidate added for", userId, ":", chatroomId)
			}

			if msgType == "disconnected" {
				log.Println("⭕User", userId, "disconnected from", chatroomId)
				break
			}

			if msgType == "remote" {
				log.Println("🕹️Remote", chatroomId)
				payload, ok := initMsgObj.Payload.(dto.RemoteMessagePayload)
				if !ok {
					log.Println("🚨Error[msgType:remote]:", err)
					continue
				}

				payloadType, ok := payload.ValidateRemotePayload()
				if !ok {
					log.Println("🚨Error[payloadTypeValidation:remote]: payload form not valid")
					continue
				}

				if payloadType == lib.CursorClick || payloadType == lib.CursorMove {
					x, y, err := payload.ExtractCursorPos(payloadType)
					if err != nil {
						log.Println("🚨Error[Cursor:payload]: Couldn't extract payload")
						continue
					}

					if payloadType == lib.CursorClick {
						err := c.browserManager.Cursor.Move(x, y)
						if err != nil {
							log.Println("🚨Error[CursorClick.Move]:", err)
							continue
						}
						err = c.browserManager.Cursor.Click()
						if err != nil {
							log.Println("Error[CursorClick.Click]:", err)
							continue
						}
					}

					if payloadType == lib.CursorMove {
						err = c.browserManager.Cursor.Move(x, y)
						if err != nil {
							log.Println("🚨Error[CursorMove.Move]:", err)
							continue
						}
					}
				}

				if payloadType == lib.Key {
					keys, err := payload.ExtractKeys(payloadType)
					if err != nil {
						log.Println("🚨Error[Key:payload]: Couldn't extract payload")
						continue
					}

					err = c.browserManager.Keyboard.SendKeys(keys)
					if err != nil {
						log.Println("🚨Error[Keys]:", err)
						continue
					}
				}
				if payloadType == lib.Undefined {
					return nil
				}
			}

			if msgType == "stop-stream" {
				log.Println("⛔ Stopping Stream")

				err := c.browserManager.Pipeline.SetState(gst.StateNull)
				if err != nil {
					log.Println("Error stopping stream[Pipeline.SetState(gst.StateNull)]", err)
				}

				userIds, err := c.s.Store.GetUsersByChatroomId(c.s.Ctx, chatroomId)
				if err != nil {
					log.Println("Error getting users by chatroom id:", err)
				}

				c.mu.Lock()
				for _, conn := range c.conns {
					if lib.Contains(userIds, conn.UserId) {
						conn.StreamConfig.PeerConnection.Close()
					}
				}
				c.mu.Unlock()

				c.chatroomCtx[chatroomId].cancel()
				delete(c.chatroomCtx, chatroomId)
				log.Println("⛔👍Stream Ended")
				break
			}
		}
	}

	log.Println("👋 Client Disconnected")
	return nil
}
