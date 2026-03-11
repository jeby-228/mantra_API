package services

import (
	"errors"
	"time"

	"mantra_API/models"

	"gorm.io/gorm"
)

type QuoteRecordService struct {
	DB *gorm.DB
}

func NewQuoteRecordService(db *gorm.DB) *QuoteRecordService {
	return &QuoteRecordService{DB: db}
}

// CreateQuoteRecord 建立新名言紀錄
func (s *QuoteRecordService) CreateQuoteRecord(
	jbName, quote string,
	saidAt time.Time,
	creatorId uint,
) (*models.QuoteRecord, error) {
	if quote == "" {
		return nil, errors.New("名言內容不得為空")
	}

	now := time.Now()
	record := &models.QuoteRecord{
		Base: models.Base{
			CreationTime: now,
			CreatorId:    creatorId,
			IsDeleted:    false,
		},
		JBName: jbName,
		Quote:  quote,
		SaidAt: saidAt,
	}

	if err := s.DB.Create(record).Error; err != nil {
		return nil, err
	}

	return record, nil
}

// UpdateQuoteRecord 更新名言紀錄
func (s *QuoteRecordService) UpdateQuoteRecord(
	id uint,
	updates map[string]interface{},
	modifierId uint,
) (*models.QuoteRecord, error) {
	var record models.QuoteRecord
	if err := s.DB.Where("is_deleted = ?", false).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("名言紀錄不存在")
		}
		return nil, err
	}

	now := time.Now()
	updates["last_modification_time"] = &now
	updates["last_modifier_id"] = modifierId

	if err := s.DB.Model(&record).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.DB.First(&record, id).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

// DeleteQuoteRecord 軟刪除名言紀錄
func (s *QuoteRecordService) DeleteQuoteRecord(id, deleterId uint) error {
	now := time.Now()
	result := s.DB.Model(&models.QuoteRecord{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]interface{}{
			"is_deleted":             true,
			"deleted_at":             &now,
			"last_modifier_id":       deleterId,
			"last_modification_time": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("名言紀錄不存在或已被刪除")
	}

	return nil
}

// GetQuoteRecordByID 取得單一名言紀錄
func (s *QuoteRecordService) GetQuoteRecordByID(id uint) (*models.QuoteRecord, error) {
	var record models.QuoteRecord
	if err := s.DB.Where("is_deleted = ?", false).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("名言紀錄不存在")
		}
		return nil, err
	}
	return &record, nil
}

// GetQuoteRecords 取得名言紀錄列表（支援分頁）
func (s *QuoteRecordService) GetQuoteRecords(
	limit, offset int,
) ([]models.QuoteRecord, int64, error) {
	var records []models.QuoteRecord
	var total int64

	if err := s.DB.Model(&models.QuoteRecord{}).
		Where("is_deleted = ?", false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.DB.Where("is_deleted = ?", false).
		Order("sort ASC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
