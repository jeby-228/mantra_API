package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MantraDailyStat 口頭禪每日統計
type MantraDailyStat struct {
	MantraID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_mantra_date" json:"mantra_id"`
	Mantra   Mantra    `gorm:"foreignKey:MantraID;references:ID"              json:"mantra,omitempty"`
	StatDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_mantra_date" json:"stat_date"`
	Count    int       `gorm:"not null;default:0"                             json:"count"`
	Base
}

// BeforeCreate 若未指定 ID 則自動產生 UUID
func (s *MantraDailyStat) BeforeCreate(tx *gorm.DB) error {
	EnsureBaseID(&s.Base)
	return nil
}
