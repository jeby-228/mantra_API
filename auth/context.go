package auth

import "context"

type ctxKey int

const userIDCtxKey ctxKey = iota

// ContextWithUserID 將已驗證的使用者 ID 寫入 context（供 GraphQL、gRPC 等使用）。
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userID)
}

// UserIDFromContext 從 context 讀取使用者 ID；若未設定或無效則 ok 為 false。
func UserIDFromContext(ctx context.Context) (userID int64, ok bool) {
	v := ctx.Value(userIDCtxKey)
	if v == nil {
		return 0, false
	}
	id, typed := v.(int64)
	if !typed || id <= 0 {
		return 0, false
	}
	return id, true
}
