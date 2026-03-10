package models

import "time"

// MantraRecord 口頭禪紀錄檔
type MantraRecord struct {
	MantraID uint       `gorm:"not null;index"                    json:"mantra_id"`
	Mantra   Mantra     `gorm:"foreignKey:MantraID;references:ID" json:"mantra,omitempty"`
	Location string     `gorm:"size:255"                          json:"location"`
	SaidAt   *time.Time `                                         json:"said_at"`
	Base
}

// GetSaidTime 取得說出時間，若 SaidAt 為空則返回 CreationTime
func (r *MantraRecord) GetSaidTime() time.Time {
	if r.SaidAt != nil {
		return *r.SaidAt
	}
	return r.CreationTime
}
