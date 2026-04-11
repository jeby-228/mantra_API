package models

import "gorm.io/gorm"

// Mantra 口頭禪主檔
type Mantra struct {
	Content     string `gorm:"size:255;not null" json:"content"`
	Description string `gorm:"size:500"          json:"description"`
	Base
}

// BeforeCreate 若未指定 ID 則自動產生 UUID
func (m *Mantra) BeforeCreate(tx *gorm.DB) error {
	EnsureBaseID(&m.Base)
	return nil
}
