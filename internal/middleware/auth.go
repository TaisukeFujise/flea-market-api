package middleware

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/labstack/echo/v5"
)

type AuthMiddleware struct {
	Client *auth.Client
	DB     *sql.DB
}

func (m *AuthMiddleware) AuthRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		authorization := c.Request().Header.Get("Authorization")

		idToken, ok := strings.CutPrefix(authorization, "Bearer ")
		if !ok || idToken == "" {
			return apperror.ErrUnauthorized.New("invalid authorization header")
		}
		token, err := m.Client.VerifyIDToken(c.Request().Context(), idToken)
		if err != nil {
			return apperror.ErrUnauthorized.New("invalid authorization header")
		}
		var deletionRequestedAt *time.Time
		err = m.DB.QueryRowContext(c.Request().Context(),
			`SELECT deleted_at FROM users WHERE id = $1`,
			token.UID,
		).Scan(&deletionRequestedAt)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrInternal.Wrap(err, "failed to check deletion status")
		}
		if deletionRequestedAt != nil {
			return apperror.ErrUnauthorized.New("account has been deleted")
		}
		c.Set("firebase_uid", token.UID)
		return next(c)
	}
}
