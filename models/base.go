package models

import (
	"time"

	"github.com/google/uuid"
)

// Base 共用模型：主鍵為 GUID，審計欄位為會員 UUID
type Base struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Sort                 int        `                            json:"sort"`
	CreationTime         time.Time  `gorm:"autoCreateTime"       json:"created_at"`
	CreatorId            uuid.UUID  `gorm:"type:uuid"            json:"creator_id"`
	LastModificationTime *time.Time `gorm:"autoUpdateTime"       json:"last_modification_time"`
	LastModifierId       uuid.UUID  `gorm:"type:uuid"            json:"last_modifier_id"`
	IsDeleted            bool       `gorm:"default:false"        json:"-"`
	DeletedAt            *time.Time `gorm:"index"                json:"-"`
}
