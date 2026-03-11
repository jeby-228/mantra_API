package services

import (
	"testing"
	"time"

	"mantra_API/internal/testhelper"

	"github.com/stretchr/testify/assert"
)

func TestQuoteRecordService_CreateQuoteRecord_EmptyQuote(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewQuoteRecordService(db)

	record, err := svc.CreateQuoteRecord("JB", "", time.Now(), 1)
	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestQuoteRecordService_CreateAndDeleteQuoteRecord(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewQuoteRecordService(db)

	record, err := svc.CreateQuoteRecord("JB", "今天也要加油", time.Now(), 1)
	assert.NoError(t, err)

	err = svc.DeleteQuoteRecord(record.ID, 9)
	assert.NoError(t, err)

	_, err = svc.GetQuoteRecordByID(record.ID)
	assert.Error(t, err)
}
