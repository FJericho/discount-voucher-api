package entity

import "time"

type Voucher struct {
	ID              string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Code            string    `gorm:"uniqueIndex;size:100" json:"code"`
	DiscountPercent int       `gorm:"type:int;check:discount_percent >= 0 AND discount_percent <= 100" json:"discount_percent"`
	ExpiryDate      time.Time 
	CreatedAt       time.Time `gorm:"default:current_timestamp"`
	UpdatedAt       time.Time 
}
