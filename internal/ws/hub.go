package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pusher/pusher-http-go"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local and intranet connections
	},
}

// WSMessage represents a message exchanged over the native websocket
type WSMessage struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Data    interface{} `json:"data"`
}

// Client represents a connected websocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	channels map[string]bool
	mu       sync.Mutex
}

// Hub maintains the set of active clients and broadcasts messages to channels
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan WSMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new websocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan WSMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			payload, err := json.Marshal(message)
			if err != nil {
				log.Println("Error marshaling ws message:", err)
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				client.mu.Lock()
				subscribed := client.channels[message.Channel]
				client.mu.Unlock()

				if subscribed {
					select {
					case client.send <- payload:
					default:
						// If send buffer is full, close and mark for cleanup
						h.mu.RUnlock()
						h.mu.Lock()
						delete(h.clients, client)
						close(client.send)
						h.mu.Unlock()
						h.mu.RLock()
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// ServeWS handles websocket requests from the peer
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Websocket upgrade error:", err)
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		channels: make(map[string]bool),
	}

	// Always auto-subscribe to public channel by default
	client.channels["public-channel"] = true

	// Check if a specific user channel or channels were requested via query params
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		client.channels["private-channel-"+userID] = true
	}
	if chanParam := r.URL.Query().Get("channel"); chanParam != "" {
		client.channels[chanParam] = true
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket client closed: %v", err)
			}
			break
		}

		var incoming struct {
			Event   string `json:"event"`
			Channel string `json:"channel"`
		}
		if err := json.Unmarshal(message, &incoming); err == nil {
			if incoming.Event == "subscribe" && incoming.Channel != "" {
				c.mu.Lock()
				c.channels[incoming.Channel] = true
				c.mu.Unlock()
			} else if incoming.Event == "unsubscribe" && incoming.Channel != "" {
				c.mu.Lock()
				delete(c.channels, incoming.Channel)
				c.mu.Unlock()
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Implement models.WSClient interface on Hub

// Trigger sends an event to a single channel
func (h *Hub) Trigger(channel string, eventName string, data interface{}) error {
	h.broadcast <- WSMessage{
		Channel: channel,
		Event:   eventName,
		Data:    data,
	}
	return nil
}

// TriggerMulti sends an event to multiple channels
func (h *Hub) TriggerMulti(channels []string, eventName string, data interface{}) error {
	for _, ch := range channels {
		_ = h.Trigger(ch, eventName, data)
	}
	return nil
}

// TriggerExclusive triggers an event excluding a socket
func (h *Hub) TriggerExclusive(channel string, eventName string, data interface{}, socketID string) error {
	return h.Trigger(channel, eventName, data)
}

// TriggerMultiExclusive triggers an event across multiple channels excluding a socket
func (h *Hub) TriggerMultiExclusive(channels []string, eventName string, data interface{}, socketID string) error {
	return h.TriggerMulti(channels, eventName, data)
}

// TriggerBatch triggers a batch of events
func (h *Hub) TriggerBatch(batch []pusher.Event) error {
	for _, ev := range batch {
		_ = h.Trigger(ev.Channel, ev.Name, ev.Data)
	}
	return nil
}

// Channels returns the list of channels
func (h *Hub) Channels(additionalQueries map[string]string) (*pusher.ChannelsList, error) {
	return &pusher.ChannelsList{}, nil
}

// Channel returns channel information
func (h *Hub) Channel(name string, additionalQueries map[string]string) (*pusher.Channel, error) {
	return &pusher.Channel{Name: name}, nil
}

// GetChannelUsers returns users in a channel
func (h *Hub) GetChannelUsers(name string) (*pusher.Users, error) {
	return &pusher.Users{}, nil
}

// AuthenticatePrivateChannel authenticates a private channel
func (h *Hub) AuthenticatePrivateChannel(params []byte) ([]byte, error) {
	return []byte(`{"auth":"native-auth-ok"}`), nil
}

// AuthenticatePresenceChannel authenticates a presence channel
func (h *Hub) AuthenticatePresenceChannel(params []byte, member pusher.MemberData) ([]byte, error) {
	return []byte(`{"auth":"native-auth-ok"}`), nil
}

// Webhook processes a pusher webhook
func (h *Hub) Webhook(header http.Header, body []byte) (*pusher.Webhook, error) {
	return &pusher.Webhook{}, nil
}
