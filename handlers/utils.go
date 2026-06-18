package handlers

import (
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
	// Silakan ganti dengan Server Key Sandbox Midtrans Anda

	// Ganti baris lama Bos menjadi ini:
	s.New(os.Getenv("MIDTRANS_SERVER_KEY"), midtrans.Production)

	// LOGIKA PENENTUAN HARGA BERDASARKAN PLAN TYPE
	var harga int
	switch p.PlanType {
	case "PRO":
		harga = 150000 // Contoh: Rp 150.000 (Sesuaikan nilainya sendiri bos)
	case "PREMIUM":
		harga = 300000 // Contoh: Rp 300.000 (Sesuaikan nilainya sendiri bos)
	default:
		harga = 50000 // Harga standar jika tipe plan tidak sesuai
	}

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  p.InvoiceID,
			GrossAmt: int64(harga),
		},
		// ⚠️ PERBAIKAN DI SINI: Pasang langsung tanpa pointer ke snap, melainkan struct dari midtrans core
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
