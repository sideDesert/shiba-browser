package controller

import (
	"context"
	"log"
	"net/http"
	"sideDesert/shiba/internal/server/controller/common"
	"sideDesert/shiba/internal/server/lib"
	"sideDesert/shiba/internal/server/services"
	vb "sideDesert/shiba/internal/vbrowser"

	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

type Controller struct {
	s              *services.Service
	nats           *nats.Conn
	conns          map[*websocket.Conn]*lib.ConnMap
	chatroomCtx    map[string]ChatroomCtx
	mu             sync.Mutex
	browserManager *vb.VbrowserManager
}

type ChatroomCtx struct {
	ctx       context.Context
	cancel    context.CancelFunc
	Streaming bool
}

func (c *Controller) CloseDbConn(ctx context.Context) {
	c.s.Store.Close(ctx)
}

func NewController(s *services.Service, nats *nats.Conn, browserManager *vb.VbrowserManager) *Controller {
	return &Controller{
		s:              s,
		nats:           nats,
		conns:          make(map[*websocket.Conn]*lib.ConnMap),
		chatroomCtx:    make(map[string]ChatroomCtx),
		browserManager: browserManager,
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
		"stream":           common.NewCMV(c.handleStream, true),
		"remote":           common.NewCMV(c.handleRemote, true),
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
