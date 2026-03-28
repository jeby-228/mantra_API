package services

import (
	"errors"
	"time"

	"mantra_API/audit"
	"mantra_API/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MantraRecordService struct {
	DB *gorm.DB
}

func NewMantraRecordService(db *gorm.DB) *MantraRecordService {
	return &MantraRecordService{DB: db}
}

// CreateMantraRecord 建立口頭禪紀錄，並同步更新每日統計（交易式）
func (s *MantraRecordService) CreateMantraRecord(
	mantraID uint,
	location string,
	saidAt *time.Time,
	creatorId uuid.UUID,
) (*models.MantraRecord, error) {
	// 確認口頭禪存在
	var mantra models.Mantra
	if err := s.DB.Where("id = ? AND is_deleted = ?", mantraID, false).
		First(&mantra).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("口頭禪不存在")
		}
		return nil, err
	}

	now := time.Now()
	record := &models.MantraRecord{
		Base:     audit.NewCreateBaseAt(now, creatorId),
		MantraID: mantraID,
		Location: location,
		SaidAt:   saidAt,
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		// 取得統計日期：優先使用 SaidAt，否則使用 CreationTime
		statTime := record.GetSaidTime()
		dateOnly := time.Date(
			statTime.Year(),
			statTime.Month(),
			statTime.Day(),
			0, 0, 0, 0,
			statTime.Location(),
		)

		dailyStat := &models.MantraDailyStat{
			Base:     audit.NewCreateBaseAt(now, creatorId),
			MantraID: mantraID,
			StatDate: dateOnly,
			Count:    1,
		}

		// Upsert MantraDailyStat：存在則 +1，不存在則新建
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "mantra_id"},
				{Name: "stat_date"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"count":                  gorm.Expr("count + ?", 1),
				"last_modification_time": now,
				"last_modifier_id":       creatorId,
			}),
		}).Create(dailyStat).Error
	})
	if err != nil {
		return nil, err
	}

	return record, nil
}

// DeleteMantraRecord 軟刪除口頭禪紀錄，並同步遞減每日統計（交易式）
func (s *MantraRecordService) DeleteMantraRecord(id uint, deleterId uuid.UUID) error {
	var record models.MantraRecord
	if err := s.DB.Where("is_deleted = ?", false).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("口頭禪紀錄不存在")
		}
		return err
	}

	statTime := record.GetSaidTime()
	dateOnly := time.Date(
		statTime.Year(),
		statTime.Month(),
		statTime.Day(),
		0, 0, 0, 0,
		statTime.Location(),
	)

	now := time.Now()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// 軟刪除紀錄
		if err := tx.Model(&record).
			Updates(audit.SoftDeleteFieldsAt(now, deleterId)).
			Error; err != nil {
			return err
		}

		// 遞減每日統計 count（最低為 0）
		auditUpdates := map[string]interface{}{
			"count": gorm.Expr(
				"CASE WHEN count > 0 THEN count - 1 ELSE 0 END",
			),
		}
		audit.ApplyUpdateAudit(auditUpdates, deleterId)
		return tx.Model(&models.MantraDailyStat{}).
			Where("mantra_id = ? AND stat_date = ? AND is_deleted = ?", record.MantraID, dateOnly, false).
			Updates(auditUpdates).Error
	})
}

// GetMantraRecordByID 取得單一口頭禪紀錄
func (s *MantraRecordService) GetMantraRecordByID(id uint) (*models.MantraRecord, error) {
	var record models.MantraRecord
	if err := s.DB.Preload("Mantra").
		Where("is_deleted = ?", false).
		First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("口頭禪紀錄不存在")
		}
		return nil, err
	}
	return &record, nil
}

// GetMantraRecords 取得特定口頭禪的紀錄列表（支援分頁）
func (s *MantraRecordService) GetMantraRecords(
	mantraID uint,
	limit, offset int,
) ([]models.MantraRecord, int64, error) {
	var records []models.MantraRecord
	var total int64

	query := s.DB.Model(&models.MantraRecord{}).Where("is_deleted = ?", false)
	if mantraID > 0 {
		query = query.Where("mantra_id = ?", mantraID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Mantra").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetDailyStats 取得口頭禪最近 N 天的每日統計
func (s *MantraRecordService) GetDailyStats(
	mantraID uint,
	days int,
) ([]models.MantraDailyStat, error) {
	var stats []models.MantraDailyStat
	if days <= 0 {
		return stats, errors.New("days 必須大於 0")
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	cutoffDate := time.Date(
		cutoff.Year(),
		cutoff.Month(),
		cutoff.Day(),
		0, 0, 0, 0,
		cutoff.Location(),
	)

	if err := s.DB.Where(
		"mantra_id = ? AND stat_date >= ? AND is_deleted = ?",
		mantraID, cutoffDate, false,
	).Order("stat_date ASC").Find(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}
