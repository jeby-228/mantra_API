package audit

import (
	"time"

	"mantra_API/models"

	"github.com/google/uuid"
)

// NewCreateBase 建立寫入時的審計欄位（建立者、建立時間）。
func NewCreateBase(creatorID uuid.UUID) models.Base {
	return NewCreateBaseAt(time.Now(), creatorID)
}

// NewCreateBaseAt 以指定時間建立寫入審計欄位（同一交易內可共用 at）。
func NewCreateBaseAt(at time.Time, creatorID uuid.UUID) models.Base {
	return models.Base{
		CreationTime: at,
		CreatorId:    creatorID,
		IsDeleted:    false,
	}
}

// ApplyUpdateAudit 在 Updates map 上附加最後修改者與修改時間。
func ApplyUpdateAudit(updates map[string]interface{}, modifierID uuid.UUID) {
	now := time.Now()
	updates["last_modification_time"] = &now
	updates["last_modifier_id"] = modifierID
}

// SoftDeleteFieldsAt 以指定時間寫入軟刪除審計欄位（同一交易內可共用 at）。
func SoftDeleteFieldsAt(at time.Time, deleterID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		"is_deleted":             true,
		"deleted_at":             &at,
		"last_modifier_id":       deleterID,
		"last_modification_time": &at,
	}
}

// SoftDeleteFields 回傳軟刪除時一併寫入的審計欄位。
func SoftDeleteFields(deleterID uuid.UUID) map[string]interface{} {
	return SoftDeleteFieldsAt(time.Now(), deleterID)
}

// NewUUIDCreateBase 建立主鍵為 GUID 的實體之審計欄位（例如 Member）。
func NewUUIDCreateBase(creatorID uuid.UUID) models.UUIDBase {
	return NewUUIDCreateBaseAt(time.Now(), creatorID)
}

// NewUUIDCreateBaseAt 以指定時間建立 UUIDBase 審計欄位。
func NewUUIDCreateBaseAt(at time.Time, creatorID uuid.UUID) models.UUIDBase {
	return models.UUIDBase{
		CreationTime: at,
		CreatorId:    creatorID,
		IsDeleted:    false,
	}
}
