package models

import (
	"time"

	"gorm.io/gorm"
)

type License struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Token         string `gorm:"unique;not null" json:"token"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
	Status        string `gorm:"default:'inactive'" json:"status"`
	PlanType      string `json:"plan_type"`
	InvoiceID     string `json:"invoice_id"`

	// GANTI MENJADI POINTER AGAR MENERIMA NILAI NULL DI DATABASE
	ActivatedAt  *time.Time `json:"activated_at"`
	ExpiredAt    *time.Time `json:"expired_at"`
	LastSyncTime *time.Time `json:"last_sync_time"`
	HWID         string     `gorm:"column:hwid" json:"hwid"` // Ubah jadi HWID

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Packet struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100)" json:"name"` // Contoh: Paket Premium Bulanan
	Description string         `gorm:"type:text" json:"description"`  // Contoh: Batas 1 PC, fitur lengkap
	ActiveDays  int            `json:"active_days"`                   // Contoh: 7, 30, 365
	Price       int64          `json:"price"`
	Note        string         `gorm:"type:varchar(255)" json:"note"`             // Contoh: Masa aktif 30 Hari
	Type        string         `gorm:"type:varchar(50);default:null" json:"type"` // Mengisi tanda "BEST_SELLER" atau kosongi jika biasa
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
