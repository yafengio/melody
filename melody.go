// Copyright (c) 2015 Ola Holmström. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package melody

import (
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

// Close codes defined in RFC 6455, section 11.7.
// Duplicate of codes from gorilla/websocket for convenience.
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseServiceRestart          = 1012
	CloseTryAgainLater           = 1013
	CloseTLSHandshake            = 1015
)

type (
	handleMessageFunc func(*Session, []byte)
	handleErrorFunc   func(*Session, error)
	handleCloseFunc   func(*Session, int, string) error
	handleSessionFunc func(*Session)
	filterFunc        func(*Session) bool
)

// Melody implements a websocket manager.
type Melody struct {
	Config                   *Config
	Upgrader                 *websocket.FastHTTPUpgrader
	messageHandler           handleMessageFunc
	messageHandlerBinary     handleMessageFunc
	messageSentHandler       handleMessageFunc
	messageSentHandlerBinary handleMessageFunc
	errorHandler             handleErrorFunc
	closeHandler             handleCloseFunc
	connectHandler           handleSessionFunc
	disconnectHandler        handleSessionFunc
	pongHandler              handleSessionFunc
	hub                      *hub
}

// New creates a new melody instance with default Upgrader and Config.
func New() *Melody {
	upgrader := &websocket.FastHTTPUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(ctx *fasthttp.RequestCtx) bool { return true },
	}

	hub := newHub()

	go hub.run()

	return &Melody{
		Config:                   newConfig(),
		Upgrader:                 upgrader,
		messageHandler:           func(*Session, []byte) {},
		messageHandlerBinary:     func(*Session, []byte) {},
		messageSentHandler:       func(*Session, []byte) {},
		messageSentHandlerBinary: func(*Session, []byte) {},
		errorHandler:             func(*Session, error) {},
		closeHandler:             nil,
		connectHandler:           func(*Session) {},
		disconnectHandler:        func(*Session) {},
		pongHandler:              func(*Session) {},
		hub:                      hub,
	}
}

// HandleConnect fires fn when a session connects.
func (m *Melody) HandleConnect(fn func(*Session)) {
	m.connectHandler = fn
}

// HandleDisconnect fires fn when a session disconnects.
func (m *Melody) HandleDisconnect(fn func(*Session)) {
	m.disconnectHandler = fn
}

// HandlePong fires fn when a pong is received from a session.
func (m *Melody) HandlePong(fn func(*Session)) {
	m.pongHandler = fn
}

// HandleMessage fires fn when a text message comes in.
// NOTE: by default melody handles messages sequentially for each
// session. This has the effect that a message handler exceeding the
// read deadline (Config.PongWait, by default 1 minute) will time out
// the session. Concurrent message handling can be turned on by setting
// Config.ConcurrentMessageHandling to true.
func (m *Melody) HandleMessage(fn func(*Session, []byte)) {
	m.messageHandler = fn
}

// HandleMessageBinary fires fn when a binary message comes in.
func (m *Melody) HandleMessageBinary(fn func(*Session, []byte)) {
	m.messageHandlerBinary = fn
}

// HandleSentMessage fires fn when a text message is successfully sent.
func (m *Melody) HandleSentMessage(fn func(*Session, []byte)) {
	m.messageSentHandler = fn
}

// HandleSentMessageBinary fires fn when a binary message is successfully sent.
func (m *Melody) HandleSentMessageBinary(fn func(*Session, []byte)) {
	m.messageSentHandlerBinary = fn
}

// HandleError fires fn when a session has an error.
func (m *Melody) HandleError(fn func(*Session, error)) {
	m.errorHandler = fn
}

// HandleClose sets the handler for close messages received from the session.
// The code argument to h is the received close code or CloseNoStatusReceived
// if the close message is empty. The default close handler sends a close frame
// back to the session.
//
// The application must read the connection to process close messages as
// described in the section on Control Frames above.
//
// The connection read methods return a CloseError when a close frame is
// received. Most applications should handle close messages as part of their
// normal error handling. Applications should only set a close handler when the
// application must perform some action before sending a close frame back to
// the session.
func (m *Melody) HandleClose(fn func(*Session, int, string) error) {
	if fn != nil {
		m.closeHandler = fn
	}
}

// HandleRequest upgrades http requests to websocket connections and dispatches them to be handled by the melody instance.
func (m *Melody) HandleRequest(ctx *fasthttp.RequestCtx) error {
	return m.HandleRequestWithKeys(ctx, nil)
}

// HandleRequestWithKeys does the same as HandleRequest but populates session.Keys with keys.
func (m *Melody) HandleRequestWithKeys(ctx *fasthttp.RequestCtx, keys map[string]any) error {
	if m.hub.closed() {
		return ErrClosed
	}

	return m.Upgrader.Upgrade(ctx, m.fastHTTPHandler(ctx, keys))
}

func (m *Melody) fastHTTPHandler(ctx *fasthttp.RequestCtx, keys map[string]any) websocket.FastHTTPHandler {
	return func(conn *websocket.Conn) {
		session := &Session{
			RequestCtx: ctx,
			Keys:       keys,
			conn:       conn,
			output:     make(chan envelope, m.Config.MessageBufferSize),
			outputDone: make(chan struct{}),
			melody:     m,
			open:       true,
			rwmutex:    &sync.RWMutex{},
		}

		m.hub.register <- session

		m.connectHandler(session)

		go session.writePump()

		session.readPump()

		if !m.hub.closed() {
			m.hub.unregister <- session
		}

		session.close()

		m.disconnectHandler(session)
	}
}

// Broadcast broadcasts a text message to all sessions.
func (m *Melody) Broadcast(msg []byte) error {
	if m.hub.closed() {
		return ErrClosed
	}

	message := envelope{t: websocket.TextMessage, msg: msg}
	m.hub.broadcast <- message

	return nil
}

// BroadcastFilter broadcasts a text message to all sessions that fn returns true for.
func (m *Melody) BroadcastFilter(msg []byte, fn func(*Session) bool) error {
	if m.hub.closed() {
		return ErrClosed
	}

	message := envelope{t: websocket.TextMessage, msg: msg, filter: fn}
	m.hub.broadcast <- message

	return nil
}

// BroadcastOthers broadcasts a text message to all sessions except session s.
func (m *Melody) BroadcastOthers(msg []byte, s *Session) error {
	return m.BroadcastFilter(msg, func(q *Session) bool {
		return s != q
	})
}

// BroadcastMultiple broadcasts a text message to multiple sessions given in the sessions slice.
func (m *Melody) BroadcastMultiple(msg []byte, sessions []*Session) error {
	for _, sess := range sessions {
		if writeErr := sess.Write(msg); writeErr != nil {
			return writeErr
		}
	}
	return nil
}

// BroadcastBinary broadcasts a binary message to all sessions.
func (m *Melody) BroadcastBinary(msg []byte) error {
	if m.hub.closed() {
		return ErrClosed
	}

	message := envelope{t: websocket.BinaryMessage, msg: msg}
	m.hub.broadcast <- message

	return nil
}

// BroadcastBinaryFilter broadcasts a binary message to all sessions that fn returns true for.
func (m *Melody) BroadcastBinaryFilter(msg []byte, fn func(*Session) bool) error {
	if m.hub.closed() {
		return ErrClosed
	}

	message := envelope{t: websocket.BinaryMessage, msg: msg, filter: fn}
	m.hub.broadcast <- message

	return nil
}

// BroadcastBinaryOthers broadcasts a binary message to all sessions except session s.
func (m *Melody) BroadcastBinaryOthers(msg []byte, s *Session) error {
	return m.BroadcastBinaryFilter(msg, func(q *Session) bool {
		return s != q
	})
}

// Sessions returns all sessions. An error is returned if the melody session is closed.
func (m *Melody) Sessions() ([]*Session, error) {
	if m.hub.closed() {
		return nil, ErrClosed
	}
	return m.hub.all(), nil
}

// Close closes the melody instance and all connected sessions.
func (m *Melody) Close() error {
	if m.hub.closed() {
		return ErrClosed
	}

	m.hub.exit <- envelope{t: websocket.CloseMessage, msg: []byte{}}

	return nil
}

// CloseWithMsg closes the melody instance with the given close payload and all connected sessions.
// Use the FormatCloseMessage function to format a proper close message payload.
func (m *Melody) CloseWithMsg(msg []byte) error {
	if m.hub.closed() {
		return ErrClosed
	}

	m.hub.exit <- envelope{t: websocket.CloseMessage, msg: msg}

	return nil
}

// Len return the number of connected sessions.
func (m *Melody) Len() int {
	return m.hub.len()
}

// IsClosed returns the status of the melody instance.
func (m *Melody) IsClosed() bool {
	return m.hub.closed()
}

// FormatCloseMessage formats closeCode and text as a WebSocket close message.
func FormatCloseMessage(closeCode int, text string) []byte {
	return websocket.FormatCloseMessage(closeCode, text)
}
