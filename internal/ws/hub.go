package ws

import (
    "log"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

type Hub struct {
    register   chan *Client
    unregister chan *Client
    clients    map[*Client]bool
    broadcast  chan []byte
}

func NewHub() *Hub { return &Hub{
    register:   make(chan *Client),
    unregister: make(chan *Client),
    clients:    make(map[*Client]bool),
    broadcast:  make(chan []byte, 256),
}}

func (h *Hub) Run() {
    for {
        select {
        case c := <-h.register:
            h.clients[c] = true
        case c := <-h.unregister:
            if _, ok := h.clients[c]; ok { delete(h.clients, c); c.conn.Close() }
        case msg := <-h.broadcast:
            for c := range h.clients {
                c.send <- msg
            }
        }
    }
}

func (h *Hub) Broadcast(msg string) { h.broadcast <- []byte(msg) }

type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

var upgrader = websocket.Upgrader{ CheckOrigin: func(r *http.Request) bool { return true } }

func ServeWS(h *Hub, w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil { log.Print("ws upgrade:", err); return }
    c := &Client{hub: h, conn: conn, send: make(chan []byte, 256)}
    h.register <- c

    go c.writePump()
}

func (c *Client) writePump() {
    ticker := time.NewTicker(30 * time.Second)
    defer func() { ticker.Stop(); c.hub.unregister <- c }()
    for {
        select {
        case msg, ok := <-c.send:
            if !ok { _ = c.conn.WriteMessage(websocket.CloseMessage, []byte{}); return }
            if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil { return }
        case <-ticker.C:
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil { return }
        }
    }
}

