package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestContextWithUserID(t *testing.T) {
	ctx := context.Background()
	_, ok := UserIDFromContext(ctx)
	assert.False(t, ok)

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440042")
	ctx = ContextWithUserID(ctx, id)
	got, ok := UserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, id, got)
}
