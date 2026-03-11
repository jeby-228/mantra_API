package services

import (
	"testing"

	"mantra_API/internal/testhelper"

	"github.com/stretchr/testify/assert"
)

func TestMessageBoardService_CreateMessage_QuoteNotFound(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMessageBoardService(db)

	msg, err := svc.CreateMessage("嗨", 999, 1)
	assert.Error(t, err)
	assert.Nil(t, msg)
}

func TestMessageBoardService_EditMessage_SetEditedFlag(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	quote := testhelper.MustCreateQuote(t, db, "測試名言")
	svc := NewMessageBoardService(db)

	created, err := svc.CreateMessage("原始留言", quote.ID, 1)
	assert.NoError(t, err)
	assert.False(t, created.IsEdited)

	updated, err := svc.EditMessage(created.ID, "更新留言", 2)
	assert.NoError(t, err)
	assert.Equal(t, "更新留言", updated.Message)
	assert.True(t, updated.IsEdited)
	assert.Equal(t, uint(2), updated.LastModifierId)
}
