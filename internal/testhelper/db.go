package testhelper

import (
	"testing"
	"time"

	"mantra_API/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSQLiteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}

	err = db.AutoMigrate(
		&models.Mantra{},
		&models.MantraRecord{},
		&models.MantraDailyStat{},
		&models.QuoteRecord{},
		&models.MessageBoard{},
	)
	if err != nil {
		t.Fatalf("migrate test db failed: %v", err)
	}

	return db
}

func MustCreateMantra(t *testing.T, db *gorm.DB, content string) models.Mantra {
	t.Helper()
	now := time.Now()
	m := models.Mantra{
		Base: models.Base{
			CreationTime: now,
			CreatorId:    1,
			IsDeleted:    false,
		},
		Content:     content,
		Description: "test",
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("create mantra failed: %v", err)
	}
	return m
}

func MustCreateQuote(t *testing.T, db *gorm.DB, quote string) models.QuoteRecord {
	t.Helper()
	now := time.Now()
	q := models.QuoteRecord{
		Base: models.Base{
			CreationTime: now,
			CreatorId:    1,
			IsDeleted:    false,
		},
		JBName: "tester",
		Quote:  quote,
		SaidAt: now,
	}
	if err := db.Create(&q).Error; err != nil {
		t.Fatalf("create quote failed: %v", err)
	}
	return q
}
