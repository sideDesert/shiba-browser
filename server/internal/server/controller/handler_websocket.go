package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sideDesert/shiba/internal/logger"
	"sideDesert/shiba/internal/server/lib"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/pion/ion-sfu/pkg/sfu"
)

type WsWriter struct {
	conn *websocket.Conn
	send chan []byte
}

func NewWsWriter(conn *websocket.Conn, c *Controller) *WsWriter {
	ws := &WsWriter{
		conn: conn,
		send: make(chan []byte, 256),
	}
	go ws.writePump(c)
	return ws
}

func (ws *WsWriter) writePump(c *Controller) {
	log := logger.NewLogger(logger.Console, "writePump")
	for msg := range ws.send {
		err := ws.conn.WriteMessage(websocket.TextMessage, msg)

		if err != nil {
			log.Error("writing WebSocket[chatrooms.sfu.*.*.] Webrtc Message:", err)
			// Remove connection from cache safely
			c.mu.Lock()
			delete(c.conns, ws.conn)
			c.mu.Unlock()
			ws.conn.Close()
			return
		}
	}
}

func (ws *WsWriter) Send(msg []byte) {
	ws.send <- msg
}

func (c *Controller) handleWebsocket(w http.ResponseWriter, r *http.Request) error {
	// Extract user ID from request context
	log := logger.NewLogger(logger.Console, "handleWebsocket")

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

	wsWriter := NewWsWriter(conn, c)

	if c.GetChatroomCtx(chatroomId).HasUser(userId) {
		c.GetChatroomCtx(chatroomId).RemoveUserAtomic(userId)
	}
	c.GetChatroomCtx(chatroomId).AddUser(userId, conn)

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
		}

		c.mu.Unlock()

		c.GetChatroomCtx(chatroomId).RemoveUserAtomic(userId)

		if len(c.GetChatroomCtx(chatroomId).Users) == 0 {
			log.Info("Chatroom will be removed - ", chatroomId)
			// TODO: Fix this
		}

		conn.Close()
		log.Info("Connection closed with", userTag)
		log.Info("Total active connections:", len(c.conns))
	}()

	log.Info("Total active connections:", len(c.conns))

	// Subscribe to chat rooms
	// This is the part of code taking care of SENDING MESSAGES TO ALL PEERS

	c.subscribeToMsgChannels(wsWriter, userId, userChatrooms)
	c.incomingMessageHandler(wsWriter, userId, chatroomId)

	log.Info("Client Disconnected")
	return nil
}

func (c *Controller) subscribeToMsgChannels(ws *WsWriter, userId string, userChatrooms []lib.Chatroom) {
	log := logger.NewLogger(logger.Console, "natsSubscriptionHandler")
	currentConn := c.conns[ws.conn]
	for _, room := range userChatrooms {
		chatroomId := room.Id
		// Chat subscription
		sub, err := c.msgChannel.Subscribe("chatroom.chat."+chatroomId, func(msg *nats.Msg) {
			log.Info("Received Message:", string(msg.Data))
			ws.Send(msg.Data)
		})
		if err != nil {
			log.Error("NATSubscription[chatrooms.chat.*]:", err)
			continue
		}

		// Signalling
		currentConn.Subscriptions = append(currentConn.Subscriptions, sub)
		sub, err = c.msgChannel.Subscribe("chatrooms.signal.*."+chatroomId, func(msg *nats.Msg) {
			data := msg.Data
			var socketMsg lib.SocketMessage[any]
			err := json.Unmarshal(data, &socketMsg)
			if err != nil {
				log.Error("json.Unmarshal", err)
			}
			if socketMsg.Sender == userId {
				return
			}

			log.Info("Received Signal:", string(msg.Data))
			ws.Send(msg.Data)

			log.Info("Signal sent to chatroom - ", socketMsg.Subj(), string(msg.Data))
			if err != nil {
				log.Error("conn.WriteMessage[chatrooms.signal.*.]", err)
				c.mu.Lock()
				delete(c.conns, ws.conn)
				c.mu.Unlock()
				ws.conn.Close()
				c.GetChatroomCtx(chatroomId).RemoveUserAtomic(userId)
				return
			}
		})

		if err != nil {
			log.Error("c.nats.Subscribe(chatrooms.signal.*)", err)
			continue
		}

		// SFU Signalling
		// chatrooms.sfu.sdp.server-request.
		sub, err = c.msgChannel.Subscribe("chatrooms.sfu.*.*."+userId, func(msg *nats.Msg) {
			// Do not send the message back to the sender
			data := msg.Data
			var socketMsg lib.SocketMessage[any]
			err := json.Unmarshal(data, &socketMsg)
			if err != nil {
				log.Error("json.Unmarshal", err)
			}
			if socketMsg.Sender == userId {
				return
			}

			ws.Send(msg.Data)
		})
		if err != nil {
			log.Error("Error subscribing to NATS[webrtc.*]:", err)
			continue
		}
		currentConn.Subscriptions = append(currentConn.Subscriptions, sub)
	}
}

func (c *Controller) incomingMessageHandler(ws *WsWriter, userId string, chatroomId string) {
	log := logger.NewLogger(logger.Console, "incomingMessageHandler")

	for {
		_, msg, err := ws.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("Unexpected WS Close:", err)
			}
			break
		}

		// Parse message
		var socketMsg lib.SocketMessage[any]
		err = json.Unmarshal(msg, &socketMsg)
		if err != nil {
			log.Error("json.Unmarshal(initMsgObj):", err)
		}
		log.Println("📥 Received:", string(socketMsg.Subject))

		parsedMsg, err := socketMsg.Parse()

		switch v := parsedMsg.(type) {
		case *lib.ChatMessage:
			c.handleSocketChatMessage(v)

		// SFU Messages are not really supposed to be sent from the client to the server
		case *lib.SFUSdpMessage:
		case *lib.SFUIceMessage:
			payload := v.Payload
			trickleType := payload.Type
			if trickleType != "pub" && trickleType != "sub" {
				log.Error("trickleType", "Invalid Trickle type, got %s", trickleType)
			}
			candidate := payload.Trickle
			sigCode, err := v.SignalCode()
			if err != nil {
				log.Error("GetSignalCode", err)
				continue
			}
			if sigCode == lib.ClientTrickle {
				session, _ := c.sfu.GetSession(chatroomId)
				peer := session.GetPeer(userId)
				if trickleType == "pub" {
					peer.Publisher().AddICECandidate(candidate)
				}
				if trickleType == "sub" {
					peer.Subscriber().AddICECandidate(candidate)
				}
			}

		case *lib.SFUSdpReqmessage:

		// init-ic (indiv call)
		case *lib.Signal0Message:
			err := c.msgChannel.Publish(v)
			if err != nil {
				log.Error("Signal0Message.msgChannel.Publish", err)
				continue
			}
			log.Info("Published Signal-0 to subject - ", v.Subject)

		// init-gc (goup call)
		case *lib.Signal1Message:
			c.GetChatroomCtx(chatroomId).SetOnCall(true)
			err := c.msgChannel.Publish(v)
			if err != nil {
				log.Error("Signal1Message.msgChannel.Publish", err)
				continue
			}
			peerA := sfu.NewPeer(c.sfu)
			err = peerA.Join(chatroomId, userId)
			if err != nil {
				log.Error("Signal1Message.peerA.Join", err)
				continue
			}
			SetupPeer(peerA, c.msgChannel, userId)
			payload := v.Payload
			answer, err := peerA.Answer(payload.Sdp)
			sfuPayload := lib.SdpPayload{
				Offer: answer,
				Type:  "pub",
			}
			serverAnsMsg, err := lib.NewSFUSocketMessage("server.sfu", lib.ServerAnswer+"."+userId, sfuPayload)
			if err != nil {
				log.Error("Signal1Message.NewSFUSocketMessage", err)
				continue
			}
			err = c.msgChannel.Publish(&serverAnsMsg)
			if err != nil {
				log.Error("msgChannel.Publish", err)
				continue
			}

		// ans-ic (answer call)
		case *lib.Signal2Message:
			if err := c.AnsIcMsgHandler(*v, chatroomId, userId); err != nil {
				log.Error("AnsIcMsgHandler", err)
				continue
			}

		// ack-ic (ack individual call)
		case *lib.Signal11Message:
			if err = c.AckMsgHandler(*v, chatroomId, userId); err != nil {
				log.Error("AckIcMsgHandler", err)
				continue
			}

		// end-call (end call)
		case *lib.Signal3Message:
			// Move to asynch cleaner (sort of like a garbage collector)
			c.GetChatroomCtx(chatroomId).EndVideoCall()

		// ans-gc (answer group call)
		case *lib.Signal4Message:
			if v.Payload.Answer == "join" {
				sdp := v.Payload.Sdp
				peerB := sfu.NewPeer(c.sfu)
				err := peerB.Join(chatroomId, userId)
				if err != nil {
					log.Error("peerB.Join", err)
					continue
				}
				SetupPeer(peerB, c.msgChannel, userId)
				answer, err := peerB.Answer(sdp)
				if err != nil {
					log.Error("peerB.Answer", err)
					continue
				}
				sfuPayload := lib.SdpPayload{
					Type:  "pub",
					Offer: answer,
				}
				ansMsg, err := lib.NewSFUSocketMessage("server.sfu", lib.ServerAnswer+"."+userId, sfuPayload)
				if err != nil {
					log.Error("Signal4Message.NewSFUSocketMessage", err)
					continue
				}
				err = c.msgChannel.Publish(&ansMsg)
				if err != nil {
					log.Error("Signal4Message.msgChannel.Publish", err)
					continue
				}
				c.GetChatroomCtx(chatroomId).AddCallParticipant(userId, peerB)
			}

			if v.Payload.Answer == "leave" {
				c.GetChatroomCtx(chatroomId).RemoveCallParticipant(userId)
			}

		// start-stream (start stream)
		case *lib.Signal5Message:

		// join-stream (join stream)
		case *lib.Signal7Message:

		// leave-stream (leave stream)
		case *lib.Signal8Message:

		// end-stream (end stream)
		case *lib.Signal9Message:

		// remote (remote action)
		case *lib.Signal10Message:
		default:
			log.Warning("No Concrete type for message could be determined :/")
		}

		if err != nil {
			log.Error("socketMsg.Parse", "Could not parse socket message ", err)
			continue
		}

	}
}
