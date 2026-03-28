package services

import (
	"testing"

	"mantra_API/internal/testhelper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	mbCreator1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	mbCreator2 = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func TestMessageBoardService_CreateMessage_QuoteNotFound(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMessageBoardService(db)

	msg, err := svc.CreateMessage("嗨", 999, mbCreator1)
	assert.Error(t, err)
	assert.Nil(t, msg)
}

func TestMessageBoardService_EditMessage_SetEditedFlag(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	quote := testhelper.MustCreateQuote(t, db, "測試名言")
	svc := NewMessageBoardService(db)

	created, err := svc.CreateMessage("原始留言", quote.ID, mbCreator1)
	assert.NoError(t, err)
	assert.False(t, created.IsEdited)

	updated, err := svc.EditMessage(created.ID, "更新留言", mbCreator2)
	assert.NoError(t, err)
	assert.Equal(t, "更新留言", updated.Message)
	assert.True(t, updated.IsEdited)
	assert.Equal(t, mbCreator2, updated.LastModifierId)
}
