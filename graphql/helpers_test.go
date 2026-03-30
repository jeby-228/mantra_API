package graphql

import (
	"context"
	"testing"
	"time"

	"mantra_API/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRequireUserID_ReturnsErrorWhenContextMissingUser(t *testing.T) {
	got, err := requireUserID(context.Background())

	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Equal(t, uuid.Nil, got)
}

func TestRequireUserID_ReturnsUserIDWhenPresent(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	ctx := auth.ContextWithUserID(context.Background(), userID)

	got, err := requireUserID(ctx)

	assert.NoError(t, err)
	assert.Equal(t, userID, got)
}

func TestParseUUIDID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		id := "550e8400-e29b-41d4-a716-446655440002"
		got, err := parseUUIDID(id)

		assert.NoError(t, err)
		assert.Equal(t, id, got.String())
	})

	t.Run("invalid UUID", func(t *testing.T) {
		got, err := parseUUIDID("not-a-uuid")

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, got)
		assert.Contains(t, err.Error(), "需為 UUID 格式")
	})
}

func TestParseOptionalTime(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		got, err := parseOptionalTime(nil)

		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		empty := ""
		got, err := parseOptionalTime(&empty)

		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("accepts RFC3339", func(t *testing.T) {
		raw := "2026-03-30T16:08:09Z"
		got, err := parseOptionalTime(&raw)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "2026-03-30T16:08:09Z", got.UTC().Format(time.RFC3339))
	})

	t.Run("accepts yyyy-mm-dd hh:mm:ss", func(t *testing.T) {
		raw := "2026-03-30 16:08:09"
		got, err := parseOptionalTime(&raw)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, 2026, got.Year())
		assert.Equal(t, time.March, got.Month())
		assert.Equal(t, 30, got.Day())
		assert.Equal(t, 16, got.Hour())
		assert.Equal(t, 8, got.Minute())
		assert.Equal(t, 9, got.Second())
	})

	t.Run("accepts yyyy-mm-dd", func(t *testing.T) {
		raw := "2026-03-30"
		got, err := parseOptionalTime(&raw)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, 2026, got.Year())
		assert.Equal(t, time.March, got.Month())
		assert.Equal(t, 30, got.Day())
		assert.Equal(t, 0, got.Hour())
		assert.Equal(t, 0, got.Minute())
		assert.Equal(t, 0, got.Second())
	})

	t.Run("invalid format returns error", func(t *testing.T) {
		raw := "30/03/2026"
		got, err := parseOptionalTime(&raw)

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "時間格式錯誤")
	})
}

func TestNormalizeLimitOffset(t *testing.T) {
	t.Run("defaults when nil", func(t *testing.T) {
		lim, off := normalizeLimitOffset(nil, nil)

		assert.Equal(t, 50, lim)
		assert.Equal(t, 0, off)
	})

	t.Run("caps limit to 100", func(t *testing.T) {
		limit := 101
		offset := 3
		lim, off := normalizeLimitOffset(&limit, &offset)

		assert.Equal(t, 100, lim)
		assert.Equal(t, 3, off)
	})

	t.Run("ignores non-positive limit and negative offset", func(t *testing.T) {
		limit := 0
		offset := -1
		lim, off := normalizeLimitOffset(&limit, &offset)

		assert.Equal(t, 50, lim)
		assert.Equal(t, 0, off)
	})
}
