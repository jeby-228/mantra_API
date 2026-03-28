package models

import (
	"github.com/google/uuid"
)

// EnsureBaseID 建立前若未指定主鍵則產生 UUID（由各實體 BeforeCreate 呼叫）
func EnsureBaseID(b *Base) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
}
