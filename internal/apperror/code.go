package apperror

import (
	"net/http"
)

type ErrCode string

const (
	ErrBadRequest      ErrCode = "BAD_REQUEST"           // 不正なリクエスト（パースエラー・型不正など）
	ErrValidation      ErrCode = "VALIDATION_ERROR"      // バリデーション失敗（必須項目欠損・値範囲外など）
	ErrUnauthorized    ErrCode = "UNAUTHORIZED"          // 認証失敗（トークン不正・期限切れなど）
	ErrForbidden       ErrCode = "FORBIDDEN"             // 認可失敗（他ユーザーのリソースへのアクセスなど）
	ErrNotFound        ErrCode = "NOT_FOUND"             // リソース未存在
	ErrConflict        ErrCode = "CONFLICT"              // 重複登録（メールアドレス・UID重複など）
	ErrInternal        ErrCode = "INTERNAL_SERVER_ERROR" // 予期しないサーバーエラー
	ErrUnavailable     ErrCode = "SERVICE_UNAVAILABLE"   // 外部サービス障害（Gemini/Vertex AI/Meshy等）
	ErrTooManyRequests ErrCode = "TOO_MANY_REQUESTS"     // レートリミット超過
)

func (c ErrCode) HTTPStatus() int {
	switch c {
	case ErrBadRequest, ErrValidation:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	case ErrUnavailable:
		return http.StatusServiceUnavailable
	case ErrTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
