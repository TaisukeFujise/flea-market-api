package handler

import (
	"context"
	"database/sql"
	"os"

	"firebase.google.com/go/v4/auth"
	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/coder/websocket"
	"github.com/labstack/echo/v5"
)

type WsHandler struct {
	hub *Hub
	fb  *auth.Client
	db  *sql.DB
}

func NewWsHandler(hub *Hub, fb *auth.Client, db *sql.DB) *WsHandler {
	return &WsHandler{hub: hub, fb: fb, db: db}
}

func (h *WsHandler) Handle(c *echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return apperror.ErrUnauthorized.New("missing token")
	}

	fbToken, err := h.fb.VerifyIDToken(c.Request().Context(), token)
	if err != nil {
		return apperror.ErrUnauthorized.New("invalid token")
	}

	var exists bool
	if err = h.db.QueryRowContext(c.Request().Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`,
		fbToken.UID,
	).Scan(&exists); err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to check user status")
	}
	if !exists {
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

	clientCtx, clientCancel := context.WithCancel(context.Background())
	client := &Client{
		userID: fbToken.UID,
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
