package zenc

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// SignalingClient handles WebSocket signaling
type SignalingClient struct {
	url               string
	token             string
	conn              *websocket.Conn
	reconnectEnabled  bool
	reconnectAttempts int
	reconnectDelay    time.Duration
	currentAttempt    int
	connectionState   ConnectionState
	messageQueue      []WSMessage
	handlers          map[string][]func(interface{})

	mu     sync.RWMutex
	connMu sync.Mutex
	done   chan struct{}
}

// NewSignalingClient creates a new signaling client
func NewSignalingClient(opts ClientOptions) *SignalingClient {
	reconnectDelay := opts.ReconnectDelay
	if reconnectDelay == 0 {
		reconnectDelay = 1 * time.Second
	}

	reconnectAttempts := opts.ReconnectAttempts
	if reconnectAttempts == 0 {
		reconnectAttempts = 5
	}

	return &SignalingClient{
		url:               opts.URL,
		token:             opts.Token,
		reconnectEnabled:  opts.Reconnect,
		reconnectAttempts: reconnectAttempts,
		reconnectDelay:    reconnectDelay,
		connectionState:   ConnectionStateDisconnected,
		messageQueue:      make([]WSMessage, 0),
		handlers:          make(map[string][]func(interface{})),
		done:              make(chan struct{}),
	}
}

// Connect establishes WebSocket connection
func (s *SignalingClient) Connect() error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	wsURL := s.url
	if s.token != "" {
		u, err := url.Parse(s.url)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		q := u.Query()
		q.Set("token", s.token)
		u.RawQuery = q.Encode()
		wsURL = u.String()
	}

	s.setConnectionState(ConnectionStateConnecting)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		s.setConnectionState(ConnectionStateFailed)
		return fmt.Errorf("failed to connect: %w", err)
	}

	s.conn = conn
	s.setConnectionState(ConnectionStateConnected)
	s.currentAttempt = 0

	// Flush message queue
	s.flushMessageQueue()

	// Start read loop
	go s.readLoop()

	s.emit("connected", nil)

	return nil
}

// Disconnect closes the WebSocket connection
func (s *SignalingClient) Disconnect() {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	s.reconnectEnabled = false
	close(s.done)

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	s.setConnectionState(ConnectionStateDisconnected)
}

// Send sends a message through the WebSocket
func (s *SignalingClient) Send(msg WSMessage) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if s.connectionState != ConnectionStateConnected || s.conn == nil {
		s.messageQueue = append(s.messageQueue, msg)
		return nil
	}

	return s.conn.WriteJSON(msg)
}

// On registers an event handler
func (s *SignalingClient) On(event string, handler func(interface{})) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handlers[event] == nil {
		s.handlers[event] = make([]func(interface{}), 0)
	}
	s.handlers[event] = append(s.handlers[event], handler)
}

// Once registers a one-time event handler
func (s *SignalingClient) Once(event string, handler func(interface{})) {
	var wrapper func(interface{})
	wrapper = func(data interface{}) {
		handler(data)
		// Remove this handler after execution
		s.mu.Lock()
		defer s.mu.Unlock()
		if handlers, ok := s.handlers[event]; ok {
			for i, h := range handlers {
				// Compare function pointers (this is a simplification)
				_ = h
				s.handlers[event] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
	s.On(event, wrapper)
}

// GetConnectionState returns the current connection state
func (s *SignalingClient) GetConnectionState() ConnectionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectionState
}

func (s *SignalingClient) setConnectionState(state ConnectionState) {
	s.mu.Lock()
	oldState := s.connectionState
	s.connectionState = state
	s.mu.Unlock()

	if oldState != state {
		s.emit("connectionStateChanged", state)
	}
}

func (s *SignalingClient) emit(event string, data interface{}) {
	s.mu.RLock()
	handlers := s.handlers[event]
	s.mu.RUnlock()

	for _, handler := range handlers {
		go handler(data)
	}
}

func (s *SignalingClient) readLoop() {
	defer func() {
		s.handleDisconnect()
	}()

	for {
		select {
		case <-s.done:
			return
		default:
			var msg WSMessage
			err := s.conn.ReadJSON(&msg)
			if err != nil {
				return
			}

			s.emit("message", msg)
			s.emit(msg.Type, msg.Data)
		}
	}
}

func (s *SignalingClient) handleDisconnect() {
	s.setConnectionState(ConnectionStateDisconnected)
	s.emit("disconnected", map[string]interface{}{"reason": "connection closed"})

	if s.reconnectEnabled && s.currentAttempt < s.reconnectAttempts {
		s.attemptReconnect()
	} else if s.currentAttempt >= s.reconnectAttempts {
		s.setConnectionState(ConnectionStateFailed)
		s.emit("reconnectFailed", nil)
	}
}

func (s *SignalingClient) attemptReconnect() {
	s.currentAttempt++
	s.setConnectionState(ConnectionStateReconnecting)

	delay := s.reconnectDelay * time.Duration(1<<uint(s.currentAttempt-1))
	s.emit("reconnecting", map[string]interface{}{
		"attempt": s.currentAttempt,
		"delay":   delay,
	})

	time.Sleep(delay)

	if err := s.Connect(); err != nil {
		// Will trigger handleDisconnect again
	}
}

func (s *SignalingClient) flushMessageQueue() {
	s.mu.Lock()
	queue := s.messageQueue
	s.messageQueue = make([]WSMessage, 0)
	s.mu.Unlock()

	for _, msg := range queue {
		s.Send(msg)
	}
}
