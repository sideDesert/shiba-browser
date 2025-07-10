package controller

import (
	"context"
	"sideDesert/shiba/internal/logger"
	"sideDesert/shiba/internal/vbrowser"
	vb "sideDesert/shiba/internal/vbrowser"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/ion-sfu/pkg/sfu"
)

type ChatroomCtx struct {
	Users              map[string]*websocket.Conn
	BrowserManager     *vbrowser.VbrowserManager
	CallParticipants   map[string]*sfu.PeerLocal
	StreamParticipants map[string]*sfu.PeerLocal
	Session            *sfu.Session
	onCall             bool
	streaming          bool
	Port               int
	ctx                context.Context
	cancel             context.CancelFunc
	mu                 *sync.Mutex
}

func (c *Controller) NewChatroomCtx(ctx context.Context, cid string, port int) *ChatroomCtx {
	ctx, cancel := context.WithCancel(ctx)

	session, _ := c.sfu.GetSession(cid)

	return &ChatroomCtx{
		ctx:                ctx,
		mu:                 &sync.Mutex{},
		cancel:             cancel,
		onCall:             false,
		streaming:          false,
		Port:               port,
		BrowserManager:     vb.NewVbManager(port),
		Users:              make(map[string]*websocket.Conn),
		CallParticipants:   make(map[string]*sfu.PeerLocal),
		StreamParticipants: make(map[string]*sfu.PeerLocal),
		Session:            &session,
	}
}

func (c *ChatroomCtx) Streaming() bool {
	return c.streaming
}

func (c *ChatroomCtx) SetStreaming(val bool) {
	c.mu.Lock()
	c.streaming = val
	c.mu.Unlock()
}
func (c *ChatroomCtx) SetOnCall(val bool) {
	c.mu.Lock()
	c.onCall = val
	c.mu.Unlock()
}
func (c *ChatroomCtx) SetPort(port int) {
	c.mu.Lock()
	c.Port = port
	c.mu.Unlock()
}

func (c *ChatroomCtx) AddUser(userId string, conn *websocket.Conn) {
	c.mu.Lock()
	c.Users[userId] = conn
	c.mu.Unlock()
}

func (c *ChatroomCtx) AddCallParticipant(userId string, peer *sfu.PeerLocal) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if oldPeer, ok := c.CallParticipants[userId]; ok {
		logger.NewLogger(logger.Console, "AddCallParticipant").Info("Closing old SFU peer for user", userId)
		oldPeer.Close()
		delete(c.CallParticipants, userId)
	}
	logger.NewLogger(logger.Console, "AddCallParticipant").Info("Adding new SFU peer for user", userId)
	c.CallParticipants[userId] = peer
}

func (c *ChatroomCtx) EndVideoCall() {
	c.mu.Lock()
	c.SetOnCall(false)
	c.mu.Unlock()

	go func() {
		time.Sleep(300 * time.Millisecond)
		c.mu.Lock()
		defer c.mu.Unlock()
		for k, p := range c.CallParticipants {
			p.Close()
			delete(c.CallParticipants, k)
		}
	}()
}

func (c *ChatroomCtx) RemoveCallParticipant(userId string) {
	c.mu.Lock()
	peer := c.CallParticipants[userId]
	if peer.Subscriber() != nil {
		peer.Subscriber().Close()
	}
	if peer.Publisher() != nil {
		peer.Publisher().Close()
	}
	peer.Close()
	delete(c.CallParticipants, userId)
	c.mu.Unlock()
}

func (ctx *ChatroomCtx) RemoveUserAtomic(userId string) {
	ctx.mu.Lock()
	if conn, ok := ctx.Users[userId]; ok {
		conn.Close()
		delete(ctx.Users, userId)
	}
	ctx.mu.Unlock()
}

func (ctx *ChatroomCtx) HasUser(userId string) bool {
	_, ok := ctx.Users[userId]
	return ok
}

func (c *ChatroomCtx) _removeCallParticipant(userId string) {
	delete(c.CallParticipants, userId)
}
func (c *ChatroomCtx) _removeUser(userId string) {
	delete(c.Users, userId)
}

func (c *ChatroomCtx) CleanupUser(userId string) {
	c.mu.Lock()
	c._removeUser(userId)
	c._removeCallParticipant(userId)
	c.mu.Unlock()
}
