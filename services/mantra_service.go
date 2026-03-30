package services

import (
	"errors"

	"mantra_API/audit"
	"mantra_API/models"

	"github.com/google/uuid"
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
	creatorId uuid.UUID,
) (*models.Mantra, error) {
	if content == "" {
		return nil, errors.New("口頭禪內容不得為空")
	}

	mantra := &models.Mantra{
		Base:        audit.NewCreateBase(creatorId),
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
	id uuid.UUID,
	updates map[string]interface{},
	modifierId uuid.UUID,
) (*models.Mantra, error) {
	var mantra models.Mantra
	if err := s.DB.Where("is_deleted = ?", false).First(&mantra, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("口頭禪不存在")
		}
		return nil, err
	}

	audit.ApplyUpdateAudit(updates, modifierId)

	if err := s.DB.Model(&mantra).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.DB.First(&mantra, id).Error; err != nil {
		return nil, err
	}

	return &mantra, nil
}

// DeleteMantra 軟刪除口頭禪
func (s *MantraService) DeleteMantra(id, deleterId uuid.UUID) error {
	result := s.DB.Model(&models.Mantra{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(audit.SoftDeleteFields(deleterId))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("口頭禪不存在或已被刪除")
	}

	return nil
}

// GetMantraByID 取得單一口頭禪
func (s *MantraService) GetMantraByID(id uuid.UUID) (*models.Mantra, error) {
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
