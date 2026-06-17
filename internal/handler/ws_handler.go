package handler

import (
	"context"
	"os"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/coder/websocket"
	"github.com/labstack/echo/v5"
)

type WsHandler struct {
	hub *Hub
}

func NewWsHandler(hub *Hub) *WsHandler {
	return &WsHandler{hub: hub}
}

func (h *WsHandler) Handle(c *echo.Context) error {
	uid, ok := c.Get("firebase_uid").(string)
	if !ok || uid == "" {
		return apperror.ErrUnauthorized.New("unauthorized")
	}

	opts := &websocket.AcceptOptions{}
	if origin := os.Getenv("FRONTEND_ORIGIN"); origin != "" {
		opts.OriginPatterns = []string{origin}
	} else {
		opts.InsecureSkipVerify = true
	}

	conn, err := websocket.Accept(c.Response(), c.Request(), opts)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to upgrade websocket")
	}

	clientCtx, clientCancel := context.WithCancel(c.Request().Context())
	client := &Client{
		userID: uid,
		conn:   conn,
		ctx:    clientCtx,
		cancel: clientCancel,
	}

	h.hub.register <- client
	defer func() {
		h.hub.unregister <- client
		conn.CloseNow()
	}()

	connClosed := conn.CloseRead(clientCtx)
	<-connClosed.Done()
	return nil
}
