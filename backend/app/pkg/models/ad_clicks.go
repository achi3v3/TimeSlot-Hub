package models

import "time"

// AdClickStats stores cumulative counters for ad clicks (two slots)
type AdClickStats struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Clicks1   int64     `json:"clicks1"`
	Clicks2   int64     `json:"clicks2"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
