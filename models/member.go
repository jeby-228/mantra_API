package models

import (
	"gorm.io/gorm"
)

// Member represents a user stored in PostgreSQL and managed by GORM.
type Member struct {
	Name         string `gorm:"size:255;not null"             json:"name"`
	Email        string `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"size:255"                      json:"-"`
	LineID       string `gorm:"size:255;uniqueIndex"          json:"line_id"`
	Base
}

// BeforeCreate 若未指定 ID 則自動產生 UUID
func (m *Member) BeforeCreate(tx *gorm.DB) error {
	EnsureBaseID(&m.Base)
	return nil
}
