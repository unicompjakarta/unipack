package handlers

import (
	"compro/backend-golang/database"
	"compro/backend-golang/models"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

// 1. Fungsi untuk menyuplai kebutuhan generateRandomString di paket_licensi.go
func generateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	rand.Seed(time.Now().UnixNano())
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// 2. Fungsi request token Snap Midtrans menggunakan data Payload asli Anda
func getSnapTokenFromMidtrans(p Payload) string {
	var s snap.Client
	s.New(os.Getenv("MIDTRANS_SERVER_KEY"), midtrans.Production)

	// 🚀 1. Ambil harga asli paket langsung dari database VPS bos (Aman dari manipulasi frontend)
	var masterPlan models.Packet // sesuaikan nama struct table plan Anda
	err := database.DB.Where("name = ?", p.PlanType).First(&masterPlan).Error

	var harga int64
	if err != nil {
		// Fallback jika nama plan tidak terdaftar di DB
		harga = 50000
	} else {
		harga = masterPlan.Price // Ambil nominal harga asli dari data server
	}

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  p.InvoiceID,
			GrossAmt: harga, // 👈 Nominal otomatis dinamis dan aman!
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: p.CustomerName,
			Email: p.CustomerEmail,
		},
	}

	snapResp, err := s.CreateTransaction(req)
	if err != nil {
		return ""
	}

	return snapResp.Token
}

// 3. Fungsi dummy untuk simpan ke database (Sesuaikan dengan DB/GORM Anda nanti)
func saveToDatabase(p Payload, token string, status string) {
	fmt.Printf("Menyimpan lisensi %s untuk %s ke DB dengan status %s\n", p.PlanType, p.CustomerName, status)
}
