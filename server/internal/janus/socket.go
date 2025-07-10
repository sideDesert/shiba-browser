package janus

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sideDesert/shiba/internal/logger"

	"github.com/gorilla/websocket"
)

type Socket struct {
	sendChannel chan []byte
	ReadChannel chan []byte
	conn        *websocket.Conn
}

func NewSocket() *Socket {
	log := logger.NewLogger(logger.Console, "NewSocket")
	a := os.Getenv("JANUS_URL")
	re := regexp.MustCompile("(^[a-zA-Z]*)://(.*)")
	substr := re.FindStringSubmatch(a)
	if len(substr) != 3 {
		panic(fmt.Errorf("JANUS_URL is malformed"))
	}
	scheme := substr[1]
	if scheme != "wss" {
		panic(fmt.Errorf("Only wss is supported for Janus"))
	}

	host := substr[2]

	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   "/janus",
	}

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Subprotocols: []string{"janus-protocol"},
	}

	conn, _, err := dialer.Dial(u.String(), http.Header{})
	if err != nil {
		log.Error("NewSocket", err)
	}

	socket := &Socket{
		sendChannel: make(chan []byte, 128),
		ReadChannel: make(chan []byte, 128),
		conn:        conn,
	}

	go socket.handleSend()
	go socket.handleRead()

	return socket
}

func (s *Socket) handleSend() {
	log := logger.NewLogger(logger.Console, "janus.Socket.handleSend")
	for msg := range s.sendChannel {
		err := s.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Error("s.conn.WriteMessage", err)
			return
		}
	}
}

func (s *Socket) handleRead() {
	log := logger.NewLogger(logger.Console, "janus.Socket.handleRead")
	for {
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			s.Close()
			return
		}
		select {
		case s.ReadChannel <- msg:
		default:
			log.Warning("readChannel full, dropping message")
		}
	}
}

func (s *Socket) Send(data []byte) {
	log := logger.NewLogger(logger.Console, "janus.Socket.Send")
	select {
	case s.sendChannel <- data:
	default:
		log.Warning("send channel full, dropping message")
	}
}

func (s *Socket) Close() {
	close(s.sendChannel)
	close(s.ReadChannel)
	s.conn.Close()
}
