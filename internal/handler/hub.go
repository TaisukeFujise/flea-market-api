package handler

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type EventType string

const (
	EventNewMessage              EventType = "new_message"
	EventDamageDetectionComplete EventType = "damage_detection_complete"
	EventModelGenerationComplete EventType = "model_generation_complete"
)

type wsEvent struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}

type newMessagePayload struct {
	RoomID string `json:"room_id"`
}

type Client struct {
	userID string
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
}

type sendRequest struct {
	userID  string
	payload wsEvent
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	send       chan sendRequest
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
		send:       make(chan sendRequest, 64),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			if old, ok := h.clients[c.userID]; ok {
				old.cancel()
				old.conn.CloseNow()
			}
			h.clients[c.userID] = c
		case c := <-h.unregister:
			if existing, ok := h.clients[c.userID]; ok && existing == c {
				delete(h.clients, c.userID)
				c.cancel()
			}
		case req := <-h.send:
			c, ok := h.clients[req.userID]
			if !ok {
				continue
			}
			writeCtx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
			if err := wsjson.Write(writeCtx, c.conn, req.payload); err != nil {
				delete(h.clients, c.userID)
				c.cancel()
			}
			cancel()
		}
	}
}

// NotifyNewMessage implements service.MessageNotifier.
func (h *Hub) NotifyNewMessage(userID string, roomID string) {
	select {
	case h.send <- sendRequest{
		userID: userID,
		payload: wsEvent{
			Type:    EventNewMessage,
			Payload: newMessagePayload{RoomID: roomID},
		},
	}:
	default:
	}
}
