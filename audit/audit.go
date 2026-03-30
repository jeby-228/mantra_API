package audit

import (
	"time"

	"mantra_API/models"

	"github.com/google/uuid"
)

// SelfRegistrationCreatorID 公開註冊 API 建立會員時寫入的 CreatorId（註冊當下尚無既有會員可指為建立者，與 uuid.Nil 區隔以利稽核）。
var SelfRegistrationCreatorID = uuid.MustParse("00000000-0000-4000-8000-000000000001")

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

// ApplyUpdateAudit 在 Updates map 上附加最後修改者與修改時間（時間為呼叫當下）。
func ApplyUpdateAudit(updates map[string]interface{}, modifierID uuid.UUID) {
	ApplyUpdateAuditAt(updates, modifierID, time.Now())
}

// ApplyUpdateAuditAt 在 Updates map 上附加最後修改者與修改時間（同一交易內請傳入共用的 at，與 SoftDeleteFieldsAt 等一致）。
func ApplyUpdateAuditAt(updates map[string]interface{}, modifierID uuid.UUID, at time.Time) {
	updates["last_modification_time"] = at
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
