package services

import (
	"errors"
	"time"

	"mantra_API/models"

	"gorm.io/gorm"
)

type MantraService struct {
	DB *gorm.DB
}

func NewMantraService(db *gorm.DB) *MantraService {
	return &MantraService{DB: db}
}

// CreateMantra 建立新口頭禪
func (s *MantraService) CreateMantra(
	content, description string,
	creatorId uint,
) (*models.Mantra, error) {
	if content == "" {
		return nil, errors.New("口頭禪內容不得為空")
	}

	now := time.Now()
	mantra := &models.Mantra{
		Base: models.Base{
			CreationTime: now,
			CreatorId:    creatorId,
			IsDeleted:    false,
		},
		Content:     content,
		Description: description,
	}

	if err := s.DB.Create(mantra).Error; err != nil {
		return nil, err
	}

	return mantra, nil
}

// UpdateMantra 更新口頭禪資訊
func (s *MantraService) UpdateMantra(
	id uint,
	updates map[string]interface{},
	modifierId uint,
) (*models.Mantra, error) {
	var mantra models.Mantra
	if err := s.DB.Where("is_deleted = ?", false).First(&mantra, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("口頭禪不存在")
		}
		return nil, err
	}

	now := time.Now()
	updates["last_modification_time"] = &now
	updates["last_modifier_id"] = modifierId

	if err := s.DB.Model(&mantra).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.DB.First(&mantra, id).Error; err != nil {
		return nil, err
	}

	return &mantra, nil
}

// DeleteMantra 軟刪除口頭禪
func (s *MantraService) DeleteMantra(id, deleterId uint) error {
	now := time.Now()
	result := s.DB.Model(&models.Mantra{}).
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
		return errors.New("口頭禪不存在或已被刪除")
	}

	return nil
}

// GetMantraByID 取得單一口頭禪
func (s *MantraService) GetMantraByID(id uint) (*models.Mantra, error) {
	var mantra models.Mantra
	if err := s.DB.Where("is_deleted = ?", false).First(&mantra, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("口頭禪不存在")
		}
		return nil, err
	}
	return &mantra, nil
}

// GetMantras 取得口頭禪列表（支援分頁）
func (s *MantraService) GetMantras(limit, offset int) ([]models.Mantra, int64, error) {
	var mantras []models.Mantra
	var total int64

	if err := s.DB.Model(&models.Mantra{}).
		Where("is_deleted = ?", false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.DB.Where("is_deleted = ?", false).
		Order("sort ASC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&mantras).Error; err != nil {
		return nil, 0, err
	}

	return mantras, total, nil
}
