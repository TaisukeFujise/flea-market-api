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
		var deletionAt *time.Time
		err = m.DB.QueryRowContext(c.Request().Context(),
			`SELECT deleted_at FROM users WHERE id = $1`,
			token.UID,
		).Scan(&deletionAt)
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrUnauthorized.New("user not registered")
		}
		if err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to check deletion status")
		}
		if deletionAt != nil {
			return apperror.ErrUnauthorized.New("account has been deleted")
		}
		c.Set("firebase_uid", token.UID)
		return next(c)
	}
}

// QueryTokenRequired はクエリパラメータ "token" からトークンを読み、AuthRequired と同じ検証を行う。
// WebSocket エンドポイントのように Authorization ヘッダーを使えない場合に使う。
func (m *AuthMiddleware) QueryTokenRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		idToken := c.QueryParam("token")
		if idToken == "" {
			return apperror.ErrUnauthorized.New("missing token")
		}
		token, err := m.Client.VerifyIDToken(c.Request().Context(), idToken)
		if err != nil {
			return apperror.ErrUnauthorized.New("invalid token")
		}
		var deletionAt *time.Time
		err = m.DB.QueryRowContext(c.Request().Context(),
			`SELECT deleted_at FROM users WHERE id = $1`,
			token.UID,
		).Scan(&deletionAt)
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrUnauthorized.New("user not registered")
		}
		if err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to check deletion status")
		}
		if deletionAt != nil {
			return apperror.ErrUnauthorized.New("account has been deleted")
		}
		c.Set("firebase_uid", token.UID)
		return next(c)
	}
}

// TokenOnly はトークン検証のみ行い DB チェックをしない。未登録ユーザーの Register に使う。
func (m *AuthMiddleware) TokenOnly(next echo.HandlerFunc) echo.HandlerFunc {
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
		c.Set("firebase_uid", token.UID)
		return next(c)
	}
}

// TokenOptional はトークンがあれば検証して firebase_uid をセットし、なければそのまま通す。
// 未認証でもアクセスでき、認証状態で挙動が変わるエンドポイントに使う。
func (m *AuthMiddleware) TokenOptional(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		authorization := c.Request().Header.Get("Authorization")
		idToken, ok := strings.CutPrefix(authorization, "Bearer ")
		if ok && idToken != "" {
			if token, err := m.Client.VerifyIDToken(c.Request().Context(), idToken); err == nil {
				c.Set("firebase_uid", token.UID)
			}
		}
		return next(c)
	}
}
