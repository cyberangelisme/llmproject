package model

import "time"

type ForwardIndex struct {
	ID        uint   `gorm:"primaryKey"`
	DocId     string `gorm:"size:64;uniqueIndex"` //采用string 为了方便以后的扩展
	Title     string `gorm:"size:255"`
	Body      string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
