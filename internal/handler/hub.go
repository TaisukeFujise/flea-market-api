package handler

import (
	"context"
	"log/slog"
	"time"

	"github.com/TaisukeFujise/flea-market-api/internal/service"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type EventType string

const (
	EventNewMessage              EventType = "new_message"
	EventDamageDetectionComplete EventType = "damage_detection_complete"
	EventDamageDetectionFailed   EventType = "damage_detection_failed"
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

type damageDetectionCompletePayload struct {
	Condition     string              `json:"condition"`
	ConditionNote string              `json:"condition_note"`
	Damages       []damageItemPayload `json:"damages"`
}

type damageItemPayload struct {
	ImageID     string  `json:"image_id"`
	ImageURL    string  `json:"image_url"`
	ImageAngle  string  `json:"image_angle"`
	DamageType  string  `json:"damage_type"`
	BboxX1      *int    `json:"bbox_x1"`
	BboxY1      *int    `json:"bbox_y1"`
	BboxX2      *int    `json:"bbox_x2"`
	BboxY2      *int    `json:"bbox_y2"`
	Description *string `json:"description"`
}

// NotifyDamageDetectionComplete implements service.DetectionNotifier.
func (h *Hub) NotifyDamageDetectionComplete(userID string, notif service.DamageDetectionNotification) {
	damages := make([]damageItemPayload, len(notif.Damages))
	for i, d := range notif.Damages {
		damages[i] = damageItemPayload{
			ImageID:     d.ImageID,
			ImageURL:    d.ImageURL,
			ImageAngle:  string(d.ImageAngle),
			DamageType:  string(d.DamageType),
			BboxX1:      d.BboxX1,
			BboxY1:      d.BboxY1,
			BboxX2:      d.BboxX2,
			BboxY2:      d.BboxY2,
			Description: d.Description,
		}
	}
	select {
	case h.send <- sendRequest{
		userID: userID,
		payload: wsEvent{
			Type: EventDamageDetectionComplete,
			Payload: damageDetectionCompletePayload{
				Condition:     string(notif.Condition),
				ConditionNote: notif.ConditionNote,
				Damages:       damages,
			},
		},
	}:
	default:
		slog.Warn("damage_detection_complete notification dropped: send channel full", "userID", userID)
	}
}

// NotifyDamageDetectionFailed implements service.DetectionNotifier.
func (h *Hub) NotifyDamageDetectionFailed(userID string) {
	select {
	case h.send <- sendRequest{
		userID: userID,
		payload: wsEvent{
			Type:    EventDamageDetectionFailed,
			Payload: nil,
		},
	}:
	default:
		slog.Warn("damage_detection_failed notification dropped: send channel full", "userID", userID)
	}
}
