package ws

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/httpx"
	"github.com/gotosky/gotosky/internal/logger"
)

type client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	mu      sync.Mutex
	clients map[*client]bool
	origins []string
	host    string
}

func NewHub(origins []string, publicHost string) *Hub {
	return &Hub{clients: map[*client]bool{}, origins: origins, host: publicHost}
}

func (h *Hub) Publish(msg domain.Telemetry) {
	b, _ := json.Marshal(map[string]any{"type": "telemetry", "data": msg})
	h.mu.Lock()
	list := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		list = append(list, c)
	}
	h.mu.Unlock()
	for _, c := range list {
		select {
		case c.send <- b:
		default:
		}
	}
}

func (h *Hub) checkOrigin(r *http.Request) bool {
	o := r.Header.Get("Origin")
	if o == "" {
		return true
	}
	u, err := url.Parse(o)
	if err != nil {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	for _, a := range h.origins {
		if a == o {
			return true
		}
	}
	if strings.HasPrefix(u.Hostname(), "127.0.0.1") || strings.HasPrefix(u.Hostname(), "localhost") {
		return true
	}
	return false
}

func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
	up := websocket.Upgrader{CheckOrigin: h.checkOrigin, ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := up.Upgrade(httpx.WrapHijack(w), r, nil)
	if err != nil {
		logger.L().Error("ws upgrade", "err", err)
		return
	}
	c := &client{conn: conn, send: make(chan []byte, 16)}
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
	hello, _ := json.Marshal(map[string]any{"type": "hello", "data": map[string]any{"ok": true}})
	c.send <- hello
	go c.write()
	go c.read(func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
	})
}

func (c *client) write() {
	tick := time.NewTicker(20 * time.Second)
	defer func() {
		tick.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-tick.C:
			_ = c.conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(3*time.Second))
		}
	}
}

func (c *client) read(leave func()) {
	defer leave()
	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
