package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextWithUserID(t *testing.T) {
	ctx := context.Background()
	_, ok := UserIDFromContext(ctx)
	assert.False(t, ok)

	ctx = ContextWithUserID(ctx, 42)
	id, ok := UserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, int64(42), id)
}
