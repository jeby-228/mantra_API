package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MantraRecord 口頭禪紀錄檔
type MantraRecord struct {
	MantraID uuid.UUID  `gorm:"type:uuid;not null;index"          json:"mantra_id"`
	Mantra   Mantra     `gorm:"foreignKey:MantraID;references:ID" json:"mantra,omitempty"`
	Location string     `gorm:"size:255"                          json:"location"`
	SaidAt   *time.Time `                                         json:"said_at"`
	Base
}

// BeforeCreate 若未指定 ID 則自動產生 UUID
func (r *MantraRecord) BeforeCreate(tx *gorm.DB) error {
	EnsureBaseID(&r.Base)
	return nil
}

// GetSaidTime 取得說出時間，若 SaidAt 為空則返回 CreationTime
func (r *MantraRecord) GetSaidTime() time.Time {
	if r.SaidAt != nil {
		return *r.SaidAt
	}
	return r.CreationTime
}
