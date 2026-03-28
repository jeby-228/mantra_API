package models

import (
	"time"

	"github.com/google/uuid"
)

// Base 基礎模型結構，包含通用的審計欄位（建立者／修改者為會員 GUID）
type Base struct {
	ID                   uint       `gorm:"primaryKey"     json:"id"`
	Sort                 int        `                      json:"sort"`
	CreationTime         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	CreatorId            uuid.UUID  `gorm:"type:uuid"      json:"creator_id"`
	LastModificationTime *time.Time `gorm:"autoUpdateTime" json:"last_modification_time"`
	LastModifierId       uuid.UUID  `gorm:"type:uuid"      json:"last_modifier_id"`
	IsDeleted            bool       `gorm:"default:false"  json:"-"`
	DeletedAt            *time.Time `gorm:"index"          json:"-"`
}

// UUIDBase 用於主鍵為 GUID 的實體（例如 Member）
type UUIDBase struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Sort                 int        `                            json:"sort"`
	CreationTime         time.Time  `gorm:"autoCreateTime"       json:"created_at"`
	CreatorId            uuid.UUID  `gorm:"type:uuid"            json:"creator_id"`
	LastModificationTime *time.Time `gorm:"autoUpdateTime"       json:"last_modification_time"`
	LastModifierId       uuid.UUID  `gorm:"type:uuid"            json:"last_modifier_id"`
	IsDeleted            bool       `gorm:"default:false"        json:"-"`
	DeletedAt            *time.Time `gorm:"index"                json:"-"`
}
