package config

import "time"

type ProductPlan struct {
	Name     string
	Price    int64
	Duration time.Duration
	Days     int
}

// Pusat Kendali Paket & Harga SaaS Anda
var Plans = map[string]ProductPlan{
	"TRIAL": {
		Name:  "🎁 Paket Trial Nyoba",
		Price: 0,
		Days:  7,
	},
	"MONTHLY": {
		Name:  "🗓️ Premium Bulanan",
		Price: 57000,
		Days:  30,
	},
	"YEARLY": {
		Name:  "🚀 Premium Tahunan",
		Price: 500000,
		Days:  365,
	},
}
