package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Add proper origin checking
	},
}

type WebSocketHandler struct {
	natsConn *nats.Conn
	clients  map[string]*WebSocketClient
	mu       sync.RWMutex
}

type WebSocketClient struct {
	conn          *websocket.Conn
	userID        string
	subscriptions map[string]*nats.Subscription
	send          chan []byte
	handler       *WebSocketHandler
}

func NewWebSocketHandler(nc *nats.Conn) *WebSocketHandler {
	return &WebSocketHandler{
		natsConn: nc,
		clients:  make(map[string]*WebSocketClient),
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (wsh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 1. Get user ID from query parameter (JWT validation disabled for now)
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anonymous" // Default user ID for testing
	}

	// TODO: Enable JWT validation later
	// token := c.Query("token")
	// if token == "" {
	//     token = c.GetHeader("Authorization")
	// }
	// claims, err := utils.ValidateJWT(token)
	// if err != nil {
	//     c.JSON(401, gin.H{"error": "Invalid token"})
	//     return
	// }
	// userID := claims["user_id"].(string)

	// 2. Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// 3. Create client
	client := &WebSocketClient{
		conn:          conn,
		userID:        userID,
		subscriptions: make(map[string]*nats.Subscription),
		send:          make(chan []byte, 256),
		handler:       wsh,
	}

	wsh.mu.Lock()
	wsh.clients[userID] = client
	wsh.mu.Unlock()

	log.Printf("WebSocket connected: user=%s", userID)

	// 4. Start read/write pumps
	go client.writePump()
	go client.readPump()
}

// WebSocket message types
type WSMessage struct {
	Type      string          `json:"type"` // "subscribe", "unsubscribe", "request"
	Subject   string          `json:"subject,omitempty"`
	Topics    []string        `json:"topics,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
}

type WSResponse struct {
	Type      string          `json:"type"` // "response", "event", "error"
	Subject   string          `json:"subject,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
	Error     string          `json:"error,omitempty"`
}

func (client *WebSocketClient) readPump() {
	defer func() {
		client.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			client.SendError("Invalid message format")
			continue
		}

		switch wsMsg.Type {
		case "subscribe":
			client.Subscribe(wsMsg.Topics)
		case "unsubscribe":
			client.Unsubscribe(wsMsg.Topics)
		case "request":
			client.ForwardRequest(wsMsg)
		}
	}
}

func (client *WebSocketClient) writePump() {
	for message := range client.send {
		if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// Subscribe client to NATS topics and forward messages to WebSocket
func (client *WebSocketClient) Subscribe(topics []string) {
	for _, topic := range topics {
		// Replace {user_id} placeholder with actual userID
		topic = strings.ReplaceAll(topic, "{user_id}", client.userID)

		sub, err := client.handler.natsConn.Subscribe(topic, func(m *nats.Msg) {
			// Forward NATS message to WebSocket
			response := WSResponse{
				Type:    "event",
				Subject: m.Subject,
				Data:    json.RawMessage(m.Data),
			}
			responseJSON, _ := json.Marshal(response)
			client.send <- responseJSON
		})

		if err != nil {
			log.Printf("Subscribe error for %s: %v", topic, err)
			continue
		}

		client.subscriptions[topic] = sub
		log.Printf("User %s subscribed to %s", client.userID, topic)
	}
}

func (client *WebSocketClient) Unsubscribe(topics []string) {
	for _, topic := range topics {
		topic = strings.ReplaceAll(topic, "{user_id}", client.userID)

		if sub, exists := client.subscriptions[topic]; exists {
			sub.Unsubscribe()
			delete(client.subscriptions, topic)
			log.Printf("User %s unsubscribed from %s", client.userID, topic)
		}
	}
}

// ForwardRequest forwards WebSocket request to NATS and returns response
func (client *WebSocketClient) ForwardRequest(wsMsg WSMessage) {
	// Make NATS request
	resp, err := client.handler.natsConn.Request(wsMsg.Subject, wsMsg.Data, 5*time.Second)

	response := WSResponse{
		Type:      "response",
		RequestID: wsMsg.RequestID,
	}

	if err != nil {
		response.Error = "Request timeout or error"
	} else {
		response.Data = json.RawMessage(resp.Data)
	}

	responseJSON, _ := json.Marshal(response)
	client.send <- responseJSON
}

func (client *WebSocketClient) SendError(message string) {
	response := WSResponse{
		Type:  "error",
		Error: message,
	}
	responseJSON, _ := json.Marshal(response)
	client.send <- responseJSON
}

func (client *WebSocketClient) Close() {
	// Unsubscribe from all NATS topics
	for _, sub := range client.subscriptions {
		sub.Unsubscribe()
	}

	// Remove from clients map
	client.handler.mu.Lock()
	delete(client.handler.clients, client.userID)
	client.handler.mu.Unlock()

	close(client.send)
	client.conn.Close()

	log.Printf("WebSocket disconnected: user=%s", client.userID)
}
