package services

import (
	"testing"
	"time"

	"mantra_API/internal/testhelper"
	"mantra_API/models"

	"github.com/stretchr/testify/assert"
)

func TestMantraRecordService_CreateMantraRecord_UpsertDailyStat(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	mantra := testhelper.MustCreateMantra(t, db, "真的假的")
	svc := NewMantraRecordService(db)

	saidAt := time.Date(2026, 3, 11, 10, 0, 0, 0, time.Local)
	first, err := svc.CreateMantraRecord(mantra.ID, "辦公室", &saidAt, 1)
	assert.NoError(t, err)
	assert.NotZero(t, first.ID)

	second, err := svc.CreateMantraRecord(mantra.ID, "會議室", &saidAt, 1)
	assert.NoError(t, err)
	assert.NotZero(t, second.ID)

	statDate := time.Date(2026, 3, 11, 0, 0, 0, 0, time.Local)
	var stat models.MantraDailyStat
	err = db.Where("mantra_id = ? AND stat_date = ?", mantra.ID, statDate).First(&stat).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, stat.Count)
}

func TestMantraRecordService_DeleteMantraRecord_DecreaseDailyStat(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	mantra := testhelper.MustCreateMantra(t, db, "先這樣")
	svc := NewMantraRecordService(db)

	saidAt := time.Date(2026, 3, 11, 8, 0, 0, 0, time.Local)
	record, err := svc.CreateMantraRecord(mantra.ID, "A", &saidAt, 1)
	assert.NoError(t, err)

	err = svc.DeleteMantraRecord(record.ID, 9)
	assert.NoError(t, err)

	var deleted models.MantraRecord
	err = db.First(&deleted, record.ID).Error
	assert.NoError(t, err)
	assert.True(t, deleted.IsDeleted)

	statDate := time.Date(2026, 3, 11, 0, 0, 0, 0, time.Local)
	var stat models.MantraDailyStat
	err = db.Where("mantra_id = ? AND stat_date = ?", mantra.ID, statDate).First(&stat).Error
	assert.NoError(t, err)
	assert.Equal(t, 0, stat.Count)
}

func TestMantraRecordService_GetDailyStats_InvalidDays(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMantraRecordService(db)

	stats, err := svc.GetDailyStats(1, 0)
	assert.Error(t, err)
	assert.Len(t, stats, 0)
}
