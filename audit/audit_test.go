package audit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func asTimeValue(v interface{}) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, true
	case *time.Time:
		if t == nil {
			return time.Time{}, false
		}
		return *t, true
	default:
		return time.Time{}, false
	}
}

func TestNewCreateBaseAt_SetsAuditFieldsDeterministically(t *testing.T) {
	at := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	creatorID := uuid.MustParse("00000000-0000-0000-0000-000000000111")

	base := NewCreateBaseAt(at, creatorID)

	assert.Equal(t, at, base.CreationTime)
	assert.Equal(t, creatorID, base.CreatorId)
	assert.False(t, base.IsDeleted)
	assert.Nil(t, base.LastModificationTime)
	assert.Equal(t, uuid.Nil, base.LastModifierId)
	assert.Nil(t, base.DeletedAt)
}

func TestApplyUpdateAuditAt_AddsModifierAndTimestamp(t *testing.T) {
	at := time.Date(2026, 3, 30, 12, 30, 0, 0, time.UTC)
	modifierID := uuid.MustParse("00000000-0000-0000-0000-000000000222")
	updates := map[string]interface{}{
		"count": 5,
	}

	ApplyUpdateAuditAt(updates, modifierID, at)

	assert.Equal(t, 5, updates["count"])
	lastModifiedAt, ok := asTimeValue(updates["last_modification_time"])
	assert.True(t, ok, "last_modification_time should be time.Time or *time.Time")
	assert.Equal(t, at, lastModifiedAt)
	assert.Equal(t, modifierID, updates["last_modifier_id"])
}

func TestSoftDeleteFieldsAt_ReturnsConsistentDeleteAuditFields(t *testing.T) {
	at := time.Date(2026, 3, 30, 13, 0, 0, 0, time.UTC)
	deleterID := uuid.MustParse("00000000-0000-0000-0000-000000000333")

	fields := SoftDeleteFieldsAt(at, deleterID)

	assert.Equal(t, true, fields["is_deleted"])
	assert.Equal(t, deleterID, fields["last_modifier_id"])

	deletedAt, ok := fields["deleted_at"].(*time.Time)
	assert.True(t, ok)
	assert.NotNil(t, deletedAt)
	assert.Equal(t, at, *deletedAt)

	lastModifiedAt, ok := fields["last_modification_time"].(*time.Time)
	assert.True(t, ok)
	assert.NotNil(t, lastModifiedAt)
	assert.Equal(t, at, *lastModifiedAt)
}
