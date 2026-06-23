package controllers

const (
	errDBNotConfigured = "數據庫連接未配置"
	//nolint:gosec // G101: user-facing error message, not a credential
	errTokenGenerateFailed = "Token 生成失敗"
	errNotAuthenticated    = "未認證"
)
