package services

import (
	"testing"

	"mantra_API/internal/testhelper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	mantraTestCreator1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	mantraTestCreator7 = uuid.MustParse("00000000-0000-0000-0000-000000000007")
)

func TestMantraService_CreateMantra_EmptyContent(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMantraService(db)

	result, err := svc.CreateMantra("", "desc", mantraTestCreator1)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMantraService_CreateAndGetMantra(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMantraService(db)

	created, err := svc.CreateMantra("好喔", "常用語", mantraTestCreator7)
	assert.NoError(t, err)
	assert.Equal(t, mantraTestCreator7, created.CreatorId)

	got, err := svc.GetMantraByID(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "好喔", got.Content)
	assert.Equal(t, "常用語", got.Description)
}
