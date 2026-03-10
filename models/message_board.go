package models

// MessageBoard 留言板
type MessageBoard struct {
	Message       string      `gorm:"size:2000;not null"                     json:"message"`
	QuoteRecordID uint        `gorm:"not null;index"                         json:"quote_record_id"`
	QuoteRecord   QuoteRecord `gorm:"foreignKey:QuoteRecordID;references:ID" json:"quote_record,omitempty"`
	IsEdited      bool        `gorm:"default:false"                          json:"is_edited"`
	Base
}
