package auth

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const userIDCtxKey ctxKey = iota

// ContextWithUserID 將已驗證的使用者 ID（GUID）寫入 context（供 GraphQL、gRPC 等使用）。
func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userID)
}

// UserIDFromContext 從 context 讀取使用者 ID；若未設定或為 Nil 則 ok 為 false。
func UserIDFromContext(ctx context.Context) (userID uuid.UUID, ok bool) {
	v := ctx.Value(userIDCtxKey)
	if v == nil {
		return uuid.Nil, false
	}
	id, typed := v.(uuid.UUID)
	if !typed || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}
