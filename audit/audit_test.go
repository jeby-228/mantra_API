package audit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyUpdateAuditAt_SetsPointerTimeAndModifier(t *testing.T) {
	modifierID := uuid.MustParse("00000000-0000-0000-0000-000000000123")
	at := time.Date(2026, 3, 30, 16, 0, 0, 123456789, time.UTC)

	updates := map[string]interface{}{
		"count": 3,
	}

	ApplyUpdateAuditAt(updates, modifierID, at)

	gotModifier, ok := updates["last_modifier_id"].(uuid.UUID)
	require.True(t, ok, "last_modifier_id 應為 uuid.UUID")
	assert.Equal(t, modifierID, gotModifier)

	gotTimePtr, ok := updates["last_modification_time"].(*time.Time)
	require.True(t, ok, "last_modification_time 應為 *time.Time")
	require.NotNil(t, gotTimePtr)
	assert.True(t, gotTimePtr.Equal(at))

	// 呼叫端即使改動自己的 at，也不應影響 map 內的時間快照。
	at = at.Add(2 * time.Hour)
	assert.NotEqual(t, at, *gotTimePtr)
}

func TestApplyUpdateAudit_SetsNowAsPointerTime(t *testing.T) {
	modifierID := uuid.MustParse("00000000-0000-0000-0000-000000000456")
	updates := map[string]interface{}{}

	before := time.Now()
	ApplyUpdateAudit(updates, modifierID)
	after := time.Now()

	gotModifier, ok := updates["last_modifier_id"].(uuid.UUID)
	require.True(t, ok, "last_modifier_id 應為 uuid.UUID")
	assert.Equal(t, modifierID, gotModifier)

	gotTimePtr, ok := updates["last_modification_time"].(*time.Time)
	require.True(t, ok, "last_modification_time 應為 *time.Time")
	require.NotNil(t, gotTimePtr)
	assert.False(t, gotTimePtr.Before(before))
	assert.False(t, gotTimePtr.After(after))
}
