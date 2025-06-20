package dto

import "time"

type VoucherResponse struct {
	ID              string     `json:"id"`
	Code            string     `json:"code"`
	DiscountPercent int        `json:"discount_percent"`
	ExpiryDate      time.Time  `json:"expiry_date"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type VoucherRequest struct {
	Code            string    `json:"code" validate:"required"`
	DiscountPercent int       `json:"discount_percent" validate:"required"`
	ExpiryDate      time.Time `json:"expiry_date" validate:"required"`
}

type CSVImportReport struct {
	SuccessCount int            `json:"success_count"`
	FailedCount  int            `json:"failed_count"`
	FailedRows   []CSVFailedRow `json:"failed_rows,omitempty"`
}

type CSVFailedRow struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}
