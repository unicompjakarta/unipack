package models

import (
	"time"

	"gorm.io/gorm"
)

type License struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Token         string         `gorm:"uniqueIndex;type:varchar(100)" json:"token"`
	CustomerName  string         `gorm:"type:varchar(100)" json:"customer_name"`
	CustomerEmail string         `gorm:"type:varchar(100)" json:"customer_email"`
	Hwid          string         `gorm:"type:varchar(255);default:null" json:"hwid"`
	Status        string         `gorm:"type:varchar(20);default:'inactive'" json:"status"` // inactive, active, expired, blocked
	PlanType      string         `gorm:"type:varchar(20)" json:"plan_type"`                 // TRIAL, MONTHLY, YEARLY
	ActivatedAt   *time.Time     `json:"activated_at"`
	ExpiredAt     time.Time      `json:"expired_at"`
	InvoiceID     string         `gorm:"type:varchar(50);default:null" json:"invoice_id"` // Support Front-end Checkout / PG
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
