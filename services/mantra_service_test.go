package services

import (
	"testing"

	"mantra_API/internal/testhelper"

	"github.com/stretchr/testify/assert"
)

func TestMantraService_CreateMantra_EmptyContent(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMantraService(db)

	result, err := svc.CreateMantra("", "desc", 1)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMantraService_CreateAndGetMantra(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMantraService(db)

	created, err := svc.CreateMantra("好喔", "常用語", 7)
	assert.NoError(t, err)
	assert.Equal(t, uint(7), created.CreatorId)

	got, err := svc.GetMantraByID(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "好喔", got.Content)
	assert.Equal(t, "常用語", got.Description)
}
