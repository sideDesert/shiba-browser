package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sideDesert/shiba/internal/server/controller/common"
	"sideDesert/shiba/internal/server/lib"
	"sideDesert/shiba/internal/server/services"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/pion/ion-sfu/pkg/sfu"
)

type MsgChannel struct {
	nats *nats.Conn
}

func (p *MsgChannel) Publish(msg lib.Msg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.nats.Publish(msg.Subj(), data)
}

func (p *MsgChannel) Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error) {
	return p.nats.Subscribe(subj, cb)
}

type Controller struct {
	s              *services.Service
	msgChannel     *MsgChannel
	conns          map[*websocket.Conn]*lib.ConnMap
	chatroomCtxMap map[string]*ChatroomCtx
	mu             sync.Mutex
	sfu            *sfu.SFU
}

func (c *Controller) GetChatroomCtx(chatroomId string) *ChatroomCtx {
	c.mu.Lock()
	a, ok := c.chatroomCtxMap[chatroomId]

	if !ok {
		chatroomCtx := c.NewChatroomCtx(context.Background(), chatroomId, -1)
		c.chatroomCtxMap[chatroomId] = chatroomCtx
	}
	c.mu.Unlock()
	return a
}

// TODO: IMplement this
func (c *Controller) NewVideoPort() (int, error) {
	return 99, nil
}

func (c *Controller) CloseDbConn(ctx context.Context) {
	c.s.Store.Close(ctx)
}

func NewController(s *services.Service, nats *nats.Conn) *Controller {
	return &Controller{
		s:              s,
		msgChannel:     &MsgChannel{nats: nats},
		conns:          make(map[*websocket.Conn]*lib.ConnMap),
		chatroomCtxMap: make(map[string]*ChatroomCtx),
		sfu:            sfu.NewSFU(sfu.Config{}),
	}
}

func (c *Controller) Run(port string) {
	router := mux.NewRouter()
	router.Use(lib.AllowCors)

	controllerMap := map[string]*common.ControllerMapValue{
		"health":         common.NewCMV(c.handleHealth, false),
		"signup":         common.NewCMV(c.handleSignup, false),
		"logout":         common.NewCMV(c.handleLogout, false),
		"user":           common.NewCMV(c.handleUser, false),
		"login/oauth":    common.NewCMV(c.handleLogin, false),
		"oauth/callback": common.NewCMV(c.handleOAuthCallback, false),

		// These are protected
		"chat":             common.NewCMV(c.handleWebsocket, true),
		"chatroom":         common.NewCMV(c.handleChatRoom, true),
		"chatroom/history": common.NewCMV(c.handleChatHistory, true),
		"friends":          common.NewCMV(c.handleFriends, true),
		"notifications":    common.NewCMV(c.handleNotifications, true),
		"search":           common.NewCMV(c.handleSearch, true),
		// "stream":           common.NewCMV(c.handleStream, true),
		"remote": common.NewCMV(c.handleRemote, true),
	}

	for key, value := range controllerMap {
		ep := "/" + key
		handler := lib.CreateHTTPHandleFunc(value.Handler)

		if value.Protected {
			// Wrap handler with middleware properly
			wrappedHandler := lib.AuthenticateMiddleware(http.HandlerFunc(handler))
			handler = func(w http.ResponseWriter, r *http.Request) {
				wrappedHandler.ServeHTTP(w, r)
			}
		}
		router.HandleFunc(ep, handler)
	}

	log.Println("API Server Running on port", port)
	err := http.ListenAndServe(port, router)

	if err != nil {
		log.Panic("Error in http.ListenAndServe: ", err)
		panic(err)
	}
}
