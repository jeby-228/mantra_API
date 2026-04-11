package services

import (
	"testing"
	"time"

	"mantra_API/internal/testhelper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	quoteTestCreator1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	quoteTestCreator9 = uuid.MustParse("00000000-0000-0000-0000-000000000009")
)

func TestQuoteRecordService_CreateQuoteRecord_EmptyQuote(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewQuoteRecordService(db)

	record, err := svc.CreateQuoteRecord("JB", "", time.Now(), quoteTestCreator1)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestQuoteRecordService_CreateAndDeleteQuoteRecord(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewQuoteRecordService(db)

	record, err := svc.CreateQuoteRecord("JB", "今天也要加油", time.Now(), quoteTestCreator1)
	assert.NoError(t, err)

	err = svc.DeleteQuoteRecord(record.ID, quoteTestCreator9)
	assert.NoError(t, err)

	_, err = svc.GetQuoteRecordByID(record.ID)
	assert.Error(t, err)
}
