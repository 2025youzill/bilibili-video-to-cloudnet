package socket

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"bvtc/log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 54 * time.Second
)

type Client struct {
	SessionID string
	Conn      *websocket.Conn
	Send      chan []byte

	closeOnce sync.Once
	sendMu    sync.RWMutex
	closed    bool
	done      chan struct{}
}

type manager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex

	closeOnce sync.Once
	done      chan struct{}
	stopped   chan struct{}
}

var (
	defaultManager = newManager()
	startOnce      sync.Once
	upgrader       = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}

			envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
			if envOrigins == "" {
				return false
			}

			for _, allowedOrigin := range strings.Split(envOrigins, ",") {
				if origin == strings.TrimSpace(allowedOrigin) {
					return true
				}
			}
			return false
		},
	}
)

func newManager() *manager {
	return &manager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}
}

func ensureStarted() {
	startOnce.Do(func() {
		go defaultManager.start()
	})
}

func (m *manager) start() {
	defer close(m.stopped)

	log.Logger.Info("netcloud websocket server started")
	for {
		select {
		case <-m.done:
			m.closeAllClients()
			log.Logger.Info("netcloud websocket server stopped")
			return
		case c := <-m.register:
			m.mutex.Lock()
			if oldClient, ok := m.clients[c.SessionID]; ok {
				oldClient.Close()
			}
			m.clients[c.SessionID] = c
			m.mutex.Unlock()
			log.Logger.Info("netcloud websocket client registered", log.String("sid", c.SessionID))
		case c := <-m.unregister:
			m.mutex.Lock()
			current, ok := m.clients[c.SessionID]
			if ok && current == c {
				delete(m.clients, c.SessionID)
			}
			m.mutex.Unlock()
			c.Close()
			log.Logger.Info("netcloud websocket client unregistered", log.String("sid", c.SessionID))
		}
	}
}

func (m *manager) registerClient(c *Client) bool {
	select {
	case <-m.done:
		return false
	case m.register <- c:
		return true
	}
}

func (m *manager) unregisterClient(c *Client) {
	select {
	case <-m.done:
		c.Close()
	case m.unregister <- c:
	}
}

func (m *manager) closeAllClients() {
	m.mutex.Lock()
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clients = make(map[string]*Client)
	m.mutex.Unlock()

	for _, client := range clients {
		client.Close()
	}
}

func (m *manager) shutdown() {
	m.closeOnce.Do(func() {
		close(m.done)
	})

	<-m.stopped
}

func Shutdown(ctx context.Context) error {
	defaultManager.shutdown()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func Upgrade(c *gin.Context, sessionID string) (*Client, error) {
	ensureStarted()

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("SessionId", sessionID, 60*60*24*7, "/", "", true, true)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		SessionID: sessionID,
		Conn:      conn,
		Send:      make(chan []byte, 8),
		done:      make(chan struct{}),
	}
	if ok := defaultManager.registerClient(client); !ok {
		_ = conn.Close()
		return nil, http.ErrServerClosed
	}

	go client.readPump(defaultManager)
	go client.writePump(defaultManager)

	return client, nil
}

func (c *Client) readPump(m *manager) {
	defer func() {
		m.unregisterClient(c)
	}()

	c.Conn.SetReadLimit(4096)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Logger.Error("netcloud websocket set read deadline error", log.Any("err", err), log.String("sid", c.SessionID))
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Logger.Error("netcloud websocket refresh read deadline error", log.Any("err", err), log.String("sid", c.SessionID))
			return err
		}
		return nil
	})

	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Logger.Error("netcloud websocket read error", log.Any("err", err), log.String("sid", c.SessionID))
			}
			return
		}
	}
}

func (c *Client) writePump(m *manager) {
	heartbeat := time.NewTicker(pingPeriod)
	defer func() {
		heartbeat.Stop()
		m.unregisterClient(c)
	}()

	for {
		select {
		case <-c.done:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Logger.Error("netcloud websocket set close deadline error", log.Any("err", err), log.String("sid", c.SessionID))
			}
			if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
				log.Logger.Error("netcloud websocket close message error", log.Any("err", err), log.String("sid", c.SessionID))
			}
			return
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Logger.Error("netcloud websocket set write deadline error", log.Any("err", err), log.String("sid", c.SessionID))
				return
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Logger.Error("netcloud websocket close message error", log.Any("err", err), log.String("sid", c.SessionID))
				}
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Logger.Error("netcloud websocket write error", log.Any("err", err), log.String("sid", c.SessionID))
				return
			}
		case <-heartbeat.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Logger.Error("netcloud websocket set ping deadline error", log.Any("err", err), log.String("sid", c.SessionID))
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Logger.Error("netcloud websocket ping error", log.Any("err", err), log.String("sid", c.SessionID))
				return
			}
		}
	}
}

func (c *Client) SendJSON(message any) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	c.sendMu.RLock()
	defer c.sendMu.RUnlock()

	if c.closed {
		return http.ErrServerClosed
	}

	select {
	case <-c.done:
		return http.ErrServerClosed
	case c.Send <- payload:
		return nil
	default:
		return http.ErrHandlerTimeout
	}
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.sendMu.Lock()
		c.closed = true
		close(c.Send)
		c.sendMu.Unlock()

		close(c.done)
		_ = c.Conn.Close()
	})
}
