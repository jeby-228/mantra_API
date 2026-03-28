package graphql

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"mantra_API/auth"
	"mantra_API/graphql/model"
	"mantra_API/models"
)

// dbToModel converts DB Member to GraphQL model
func dbToModel(m models.Member) *model.Member {
	var created, updated *string
	if !m.CreationTime.IsZero() {
		s := formatTime(m.CreationTime)
		created = &s
	}
	if m.LastModificationTime != nil && !m.LastModificationTime.IsZero() {
		s := formatTime(*m.LastModificationTime)
		updated = &s
	}
	return &model.Member{
		ID:        formatID(m.ID),
		Name:      m.Name,
		Email:     m.Email,
		CreatedAt: created,
		UpdatedAt: updated,
	}
}

// productDBToModel converts DB Product to GraphQL model
func productDBToModel(p models.Product) *model.Product {
	var created, updated *string
	if !p.CreationTime.IsZero() {
		s := formatTime(p.CreationTime)
		created = &s
	}
	if p.LastModificationTime != nil && !p.LastModificationTime.IsZero() {
		s := formatTime(*p.LastModificationTime)
		updated = &s
	}
	return &model.Product{
		ID:                 formatID(p.ID),
		ProductName:        p.ProductName,
		ProductPrice:       p.ProductPrice,
		ProductDescription: stringPtr(p.ProductDescription),
		ProductImage:       stringPtr(p.ProductImage),
		ProductStock:       p.ProductStock,
		CreatedAt:          created,
		UpdatedAt:          updated,
	}
}

func mantraDBToModel(m models.Mantra) *model.Mantra {
	var created, updated *string
	if !m.CreationTime.IsZero() {
		s := formatTime(m.CreationTime)
		created = &s
	}
	if m.LastModificationTime != nil && !m.LastModificationTime.IsZero() {
		s := formatTime(*m.LastModificationTime)
		updated = &s
	}

	return &model.Mantra{
		ID:          formatID(m.ID),
		Content:     m.Content,
		Description: stringPtr(m.Description),
		CreatedAt:   created,
		UpdatedAt:   updated,
	}
}

func mantraRecordDBToModel(r models.MantraRecord) *model.MantraRecord {
	var created, updated, saidAt *string
	if !r.CreationTime.IsZero() {
		s := formatTime(r.CreationTime)
		created = &s
	}
	if r.LastModificationTime != nil && !r.LastModificationTime.IsZero() {
		s := formatTime(*r.LastModificationTime)
		updated = &s
	}
	if r.SaidAt != nil && !r.SaidAt.IsZero() {
		s := formatTime(*r.SaidAt)
		saidAt = &s
	}

	return &model.MantraRecord{
		ID:        formatID(r.ID),
		MantraID:  formatID(r.MantraID),
		Location:  stringPtr(r.Location),
		SaidAt:    saidAt,
		CreatedAt: created,
		UpdatedAt: updated,
	}
}

func mantraDailyStatDBToModel(s models.MantraDailyStat) *model.MantraDailyStat {
	return &model.MantraDailyStat{
		MantraID: formatID(s.MantraID),
		StatDate: s.StatDate.Format("2006-01-02"),
		Count:    s.Count,
	}
}

func quoteRecordDBToModel(q models.QuoteRecord) *model.QuoteRecord {
	var created, updated, saidAt *string
	if !q.CreationTime.IsZero() {
		s := formatTime(q.CreationTime)
		created = &s
	}
	if q.LastModificationTime != nil && !q.LastModificationTime.IsZero() {
		s := formatTime(*q.LastModificationTime)
		updated = &s
	}
	if !q.SaidAt.IsZero() {
		s := formatTime(q.SaidAt)
		saidAt = &s
	}

	return &model.QuoteRecord{
		ID:        formatID(q.ID),
		JbName:    q.JBName,
		Quote:     q.Quote,
		SaidAt:    saidAt,
		CreatedAt: created,
		UpdatedAt: updated,
	}
}

func messageBoardDBToModel(m models.MessageBoard) *model.MessageBoard {
	var created, updated *string
	if !m.CreationTime.IsZero() {
		s := formatTime(m.CreationTime)
		created = &s
	}
	if m.LastModificationTime != nil && !m.LastModificationTime.IsZero() {
		s := formatTime(*m.LastModificationTime)
		updated = &s
	}

	return &model.MessageBoard{
		ID:            formatID(m.ID),
		Message:       m.Message,
		QuoteRecordID: formatID(m.QuoteRecordID),
		IsEdited:      m.IsEdited,
		CreatedAt:     created,
		UpdatedAt:     updated,
	}
}

// formatTime formats time to RFC3339 string
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// formatID converts uint ID to string
func formatID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

// getUserIDFromContext 從 JWT 注入的 context 取得使用者 ID（未登入則為 0）
func getUserIDFromContext(ctx context.Context) uint {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return 0
	}
	return uint(userID)
}

// stringPtr converts string to *string pointer
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ptrToString converts *string pointer to string
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseUintID(id string) (uint, error) {
	parsed, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("無效的 ID")
	}
	return uint(parsed), nil
}

func parseOptionalTime(input *string) (*time.Time, error) {
	if input == nil || *input == "" {
		return nil, nil
	}

	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, *input); err == nil {
			return &parsed, nil
		}
	}

	return nil, fmt.Errorf("時間格式錯誤，請使用 RFC3339 或 yyyy-mm-dd hh:mm:ss")
}

func normalizeLimitOffset(limit, offset *int) (int, int) {
	lim := 50
	if limit != nil && *limit > 0 {
		if *limit > 100 {
			lim = 100
		} else {
			lim = *limit
		}
	}

	off := 0
	if offset != nil && *offset >= 0 {
		off = *offset
	}

	return lim, off
}
