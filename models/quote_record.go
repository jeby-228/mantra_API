package models

import "time"

// QuoteRecord 名言紀錄檔案
type QuoteRecord struct {
	JBName string    `gorm:"size:255;not null"  json:"jb_name"`
	Quote  string    `gorm:"size:1000;not null" json:"quote"`
	SaidAt time.Time `                          json:"said_at"`
	Base
}
