package models

// Mantra 口頭禪主檔
type Mantra struct {
	Content     string `gorm:"size:255;not null" json:"content"`
	Description string `gorm:"size:500"          json:"description"`
	Base
}
