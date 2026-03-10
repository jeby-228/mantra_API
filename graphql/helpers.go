package graphql

import (
	"context"
	"strconv"
	"time"

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

// formatTime formats time to RFC3339 string
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// formatID converts uint ID to string
func formatID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

// getUserIDFromContext extracts user ID from context
func getUserIDFromContext(ctx context.Context) uint {
	userID, ok := ctx.Value("user_id").(int64)
	if !ok || userID <= 0 {
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
