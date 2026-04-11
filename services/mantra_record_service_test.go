package services

import (
	"testing"
	"time"

	"mantra_API/internal/testhelper"
	"mantra_API/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	mrCreator1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	mrCreator9 = uuid.MustParse("00000000-0000-0000-0000-000000000009")
)

func TestMantraRecordService_CreateMantraRecord_UpsertDailyStat(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	mantra := testhelper.MustCreateMantra(t, db, "真的假的")
	svc := NewMantraRecordService(db)

	saidAt := time.Date(2026, 3, 11, 10, 0, 0, 0, time.Local)
	first, err := svc.CreateMantraRecord(mantra.ID, "辦公室", &saidAt, mrCreator1)
	assert.NoError(t, err)
	assert.NotZero(t, first.ID)

	second, err := svc.CreateMantraRecord(mantra.ID, "會議室", &saidAt, mrCreator1)
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
	record, err := svc.CreateMantraRecord(mantra.ID, "A", &saidAt, mrCreator1)
	assert.NoError(t, err)

	err = svc.DeleteMantraRecord(record.ID, mrCreator9)
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

	stats, err := svc.GetDailyStats(uuid.MustParse("00000000-0000-0000-0000-000000000001"), 0)
	assert.Error(t, err)
	assert.Len(t, stats, 0)
}

func TestMantraRecordService_CreateMantraRecord_UseCreationTimeWhenSaidAtNil(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	mantra := testhelper.MustCreateMantra(t, db, "下次一定")
	svc := NewMantraRecordService(db)

	record, err := svc.CreateMantraRecord(mantra.ID, "客廳", nil, mrCreator1)
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Nil(t, record.SaidAt)

	expectedDate := time.Date(
		record.CreationTime.Year(),
		record.CreationTime.Month(),
		record.CreationTime.Day(),
		0, 0, 0, 0,
		record.CreationTime.Location(),
	)

	var stat models.MantraDailyStat
	err = db.Where("mantra_id = ? AND stat_date = ?", mantra.ID, expectedDate).First(&stat).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, stat.Count)
}

func TestMantraRecordService_CreateMantraRecord_MantraNotFound(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMantraRecordService(db)

	record, err := svc.CreateMantraRecord(uuid.New(), "客廳", nil, mrCreator1)
	assert.Nil(t, record)
	assert.EqualError(t, err, "口頭禪不存在")

	var recordCount int64
	err = db.Model(&models.MantraRecord{}).Count(&recordCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), recordCount)

	var statCount int64
	err = db.Model(&models.MantraDailyStat{}).Count(&statCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), statCount)
}

func TestMantraRecordService_DeleteMantraRecord_NotFound(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMantraRecordService(db)

	err := svc.DeleteMantraRecord(uuid.New(), mrCreator9)
	assert.EqualError(t, err, "口頭禪紀錄不存在")
}

func TestMantraRecordService_GetDailyStats_FilterByDaysAndSortAscending(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	mantra := testhelper.MustCreateMantra(t, db, "不是吧")
	svc := NewMantraRecordService(db)

	now := time.Now()
	outsideRange := now.AddDate(0, 0, -5)
	inRangeOlder := now.AddDate(0, 0, -2)
	inRangeNewer := now.AddDate(0, 0, -1)

	_, err := svc.CreateMantraRecord(mantra.ID, "A", &outsideRange, mrCreator1)
	assert.NoError(t, err)
	_, err = svc.CreateMantraRecord(mantra.ID, "B", &inRangeOlder, mrCreator1)
	assert.NoError(t, err)
	_, err = svc.CreateMantraRecord(mantra.ID, "C", &inRangeNewer, mrCreator1)
	assert.NoError(t, err)

	stats, err := svc.GetDailyStats(mantra.ID, 3)
	assert.NoError(t, err)
	assert.Len(t, stats, 2)

	expectedOlderDate := inRangeOlder.Format("2006-01-02")
	expectedNewerDate := inRangeNewer.Format("2006-01-02")

	assert.Equal(t, expectedOlderDate, stats[0].StatDate.Format("2006-01-02"))
	assert.Equal(t, expectedNewerDate, stats[1].StatDate.Format("2006-01-02"))
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, 1, stats[1].Count)
	assert.Equal(t, mantra.ID, stats[0].MantraID)
	assert.Equal(t, mantra.ID, stats[1].MantraID)
}
