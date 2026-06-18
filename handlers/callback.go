package handlers

import (
	"backend-golang/database"
	"backend-golang/models"
	"fmt"
	"net/smtp"

	"github.com/gofiber/fiber/v2"
)

// Struct untuk menangkap data callback dari Midtrans
type MidtransCallbackReq struct {
	TransactionStatus string `json:"transaction_status"`
	OrderID           string `json:"order_id"`
	PaymentType       string `json:"payment_type"`
	StatusCode        string `json:"status_code"`
}

// Handler utama Callback Midtrans
func MidtransCallbackHandler(c *fiber.Ctx) error {
	var notification MidtransCallbackReq

	// 1. Parse data JSON dari Midtrans
	if err := c.BodyParser(&notification); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid payload"})
	}

	fmt.Printf("Menerima notifikasi Midtrans. OrderID: %s, Status: %s\n", notification.OrderID, notification.TransactionStatus)

	// 2. Cek apakah status transaksinya sukses (settlement = berhasil di-scan/bayar)
	if notification.TransactionStatus == "settlement" || notification.TransactionStatus == "capture" {

		var license models.License
		// Cari data lisensi berdasarkan InvoiceID / OrderID
		err := database.DB.Where("invoice_id = ?", notification.OrderID).First(&license).Error
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"message": "Data lisensi tidak ditemukan"})
		}

		// Jika status saat ini masih inactive, kita ganti jadi ACTIVE
		if license.Status != "ACTIVE" {
			license.Status = "ACTIVE"
			database.DB.Save(&license)

			// 3. 🚀 KIRIM TOKEN LANGSUNG KE EMAIL CUSTOMER
			go sendTokenEmail(license.CustomerEmail, license.CustomerName, license.Token, license.PlanType)
		}
	}

	// Midtrans hanya butuh response HTTP 200 OK sebagai tanda callback sukses diterima
	return c.Status(200).JSON(fiber.Map{"status": "ok"})
}

// Fungsi helper kirim email otomatis pake SMTP
func sendTokenEmail(toEmail string, name string, token string, plan string) {
	// ⚠️ ATUR KONFIGURASI EMAIL KELUAR ANDA DI SINI
	from := "basartech.meds@gmail.com" // ganti email bisnis bos
	password := "mrhm jzdz pgcx binq"  // ganti password smtp
	smtpHost := "smtp.gmail.com"       // jika pake gmail bisnis, atau host cpanel Anda
	smtpPort := "587"

	// Isi Konten Email (HTML Format agar rapi)
	subject := "Subject: [unicomputer] Token Lisensi Anda Telah Aktif! 🚀\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := fmt.Sprintf(`
		<h3>Halo %s, Terima Kasih!</h3>
		<p>Pembayaran QRIS Anda untuk paket <b>%s</b> telah kami terima.</p>
		<p>Berikut adalah kode token lisensi Anda yang sudah aktif dan siap digunakan:</p>
		<div style="background:#f4f4f4; padding:15px; font-family:monospace; font-size:18px; font-weight:bold; border:1px dashed #333; display:inline-block;">
			%s
		</div>
		<p>Silakan copy kode di atas dan masukkan ke dalam aplikasi gudang Bos Anda.</p>
		<hr>
		<p>Salam hangat,<br><b>Team unicomputer.id</b></p>
	`, name, plan, token)

	msg := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Jalankan pengiriman email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg)
	if err != nil {
		fmt.Println("Gagal mengirim email lisensi:", err.Error())
		return
	}
	fmt.Printf("Email lisensi berhasil terkirim ke %s!\n", toEmail)
}
