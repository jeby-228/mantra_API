package models

import "time"

// MantraDailyStat 口頭禪每日統計
type MantraDailyStat struct {
	MantraID uint      `gorm:"not null;uniqueIndex:idx_mantra_date"           json:"mantra_id"`
	Mantra   Mantra    `gorm:"foreignKey:MantraID;references:ID"              json:"mantra,omitempty"`
	StatDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_mantra_date" json:"stat_date"`
	Count    int       `gorm:"not null;default:0"                             json:"count"`
	Base
}
