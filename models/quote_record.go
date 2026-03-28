package models

import (
	"time"

	"gorm.io/gorm"
)

// QuoteRecord 名言紀錄檔案
type QuoteRecord struct {
	JBName string    `gorm:"size:255;not null"  json:"jb_name"`
	Quote  string    `gorm:"size:1000;not null" json:"quote"`
	SaidAt time.Time `                          json:"said_at"`
	Base
}

// BeforeCreate 若未指定 ID 則自動產生 UUID
func (q *QuoteRecord) BeforeCreate(tx *gorm.DB) error {
	EnsureBaseID(&q.Base)
	return nil
}
