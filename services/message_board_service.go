package services

import (
	"errors"

	"mantra_API/audit"
	"mantra_API/models"

	"gorm.io/gorm"
)

type MessageBoardService struct {
	DB *gorm.DB
}

func NewMessageBoardService(db *gorm.DB) *MessageBoardService {
	return &MessageBoardService{DB: db}
}

// CreateMessage 對名言紀錄新增留言
func (s *MessageBoardService) CreateMessage(
	message string,
	quoteRecordID uint,
	creatorId uint,
) (*models.MessageBoard, error) {
	if message == "" {
		return nil, errors.New("留言內容不得為空")
	}

	// 確認名言紀錄存在
	var quote models.QuoteRecord
	if err := s.DB.Where("id = ? AND is_deleted = ?", quoteRecordID, false).
		First(&quote).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("名言紀錄不存在")
		}
		return nil, err
	}

	msg := &models.MessageBoard{
		Base:          audit.NewCreateBase(creatorId),
		Message:       message,
		QuoteRecordID: quoteRecordID,
		IsEdited:      false,
	}

	if err := s.DB.Create(msg).Error; err != nil {
		return nil, err
	}

	return msg, nil
}

// EditMessage 編輯留言內容，並標記 IsEdited = true
func (s *MessageBoardService) EditMessage(
	id uint,
	message string,
	modifierId uint,
) (*models.MessageBoard, error) {
	if message == "" {
		return nil, errors.New("留言內容不得為空")
	}

	var msg models.MessageBoard
	if err := s.DB.Where("is_deleted = ?", false).First(&msg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("留言不存在")
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"message":   message,
		"is_edited": true,
	}
	audit.ApplyUpdateAudit(updates, modifierId)
	if err := s.DB.Model(&msg).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.DB.First(&msg, id).Error; err != nil {
		return nil, err
	}

	return &msg, nil
}

// DeleteMessage 軟刪除留言
func (s *MessageBoardService) DeleteMessage(id, deleterId uint) error {
	result := s.DB.Model(&models.MessageBoard{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(audit.SoftDeleteFields(deleterId))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("留言不存在或已被刪除")
	}

	return nil
}

// GetMessageByID 取得單一留言
func (s *MessageBoardService) GetMessageByID(id uint) (*models.MessageBoard, error) {
	var msg models.MessageBoard
	if err := s.DB.Preload("QuoteRecord").
		Where("is_deleted = ?", false).
		First(&msg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("留言不存在")
		}
		return nil, err
	}
	return &msg, nil
}

// GetMessagesByQuoteRecord 取得特定名言的所有留言（支援分頁）
func (s *MessageBoardService) GetMessagesByQuoteRecord(
	quoteRecordID uint,
	limit, offset int,
) ([]models.MessageBoard, int64, error) {
	var messages []models.MessageBoard
	var total int64

	query := s.DB.Model(&models.MessageBoard{}).
		Where("quote_record_id = ? AND is_deleted = ?", quoteRecordID, false)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("QuoteRecord").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}
