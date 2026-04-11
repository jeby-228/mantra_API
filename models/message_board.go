package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageBoard 留言板
type MessageBoard struct {
	Message       string      `gorm:"size:2000;not null"                     json:"message"`
	QuoteRecordID uuid.UUID   `gorm:"type:uuid;not null;index"               json:"quote_record_id"`
	QuoteRecord   QuoteRecord `gorm:"foreignKey:QuoteRecordID;references:ID" json:"quote_record,omitempty"`
	IsEdited      bool        `gorm:"default:false"                          json:"is_edited"`
	Base
}

// BeforeCreate 若未指定 ID 則自動產生 UUID
func (m *MessageBoard) BeforeCreate(tx *gorm.DB) error {
	EnsureBaseID(&m.Base)
	return nil
}
