package handlers

import (
	"backend-golang/database"
	"backend-golang/models"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Helper generate string token acak
func generateRandomToken(plan string) string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("UNI-%s-%X", plan, b)
}

// Render Halaman Login Admin
func GetLogin(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{})
}

// Proses Login Sederhana
func PostLogin(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Hardcoded credentials untuk Owner UNICOMPUTER
	if username == "admin" && password == "admin123" {
		c.Cookie(&fiber.Cookie{
			Name:    "admin_session",
			Value:   "logged_in_secret_key",
			Expires: time.Now().Add(24 * time.Hour),
		})
		return c.Redirect("/admin/dashboard")
	}
	return c.Redirect("/admin/login")
}

// Render Dashboard Utama (CRUD Token & Stats)
func GetDashboard(c *fiber.Ctx) error {
	// Proteksi Session Cookie
	cookie := c.Cookies("admin_session")
	if cookie != "logged_in_secret_key" {
		return c.Redirect("/admin/login")
	}

	// 1. Tarik Data Lisensi
	var licenses []models.License
	database.DB.Order("id desc").Find(&licenses)

	// 2. Tarik Data Paket (TAMBAHKAN INI BOS)
	var packets []models.Packet
	if err := database.DB.Find(&packets).Error; err != nil {
		// Jika ada error saat tarik data paket, log atau handle di sini
		packets = []models.Packet{}
	}

	// Hitung Statistik Mini
	var totalActive int64
	database.DB.Model(&models.License{}).Where("status = ?", "active").Count(&totalActive)

	// 3. Masukkan "Packets" ke dalam Map Render (WAJIB Huruf P Kapital)
	return c.Render("dashboard", fiber.Map{
		"Licenses":    licenses,
		"Packets":     packets, // SEKARANG TABLE PAKET DATA DIJAMIN MUNCUL!
		"TotalActive": totalActive,
	})
}

// Aksi Manual Generate Token dari Dashboard (Dan Checkout Ready API)
// func GenerateTokenAction(c *fiber.Ctx) error {
// 	// 1. Ambil data dari form/request
// 	customerName := c.FormValue("customer_name")
// 	customerEmail := c.FormValue("customer_email")
// 	customerPhone := c.FormValue("customer_phone") // Menambahkan field ini
// 	planType := c.FormValue("plan_type")
// 	invoiceID := c.FormValue("invoice_id")

// 	// Fallback jika dipanggil via JSON API
// 	if customerName == "" {
// 		type ApiReq struct {
// 			CustomerName  string `json:"customer_name"`
// 			CustomerEmail string `json:"customer_email"`
// 			CustomerPhone string `json:"customer_phone"` // Tambahkan di struct
// 			PlanType      string `json:"plan_type"`
// 			InvoiceID     string `json:"invoice_id"`
// 		}
// 		var apiData ApiReq
// 		if err := c.BodyParser(&apiData); err == nil && apiData.CustomerName != "" {
// 			customerName = apiData.CustomerName
// 			customerEmail = apiData.CustomerEmail
// 			customerPhone = apiData.CustomerPhone
// 			planType = apiData.PlanType
// 			invoiceID = apiData.InvoiceID
// 		}
// 	}

// 	// 2. Simpan ke database
// 	newToken := models.License{
// 		Token:         generateRandomToken(planType),
// 		CustomerName:  customerName,
// 		CustomerEmail: customerEmail,
// 		CustomerPhone: customerPhone, // Pastikan model.License punya field ini
// 		PlanType:      planType,
// 		Status:        "inactive",
// 		InvoiceID:     invoiceID,
// 	}

// 	if err := database.DB.Create(&newToken).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Gagal menyimpan ke database"})
// 	}

// 	// 3. Respon sesuai tipe request
// 	if c.Accepts("json") == "json" || c.Get("Content-Type") == "application/json" {
// 		return c.JSON(fiber.Map{
// 			"status": "success",
// 			"token":  newToken.Token,
// 		})
// 	}

// 	// Redirect PRG (Post-Redirect-Get) untuk mencegah double-post
// 	return c.Redirect("/admin/dashboard", 302)
// }

// New
func GenerateTokenAction(c *fiber.Ctx) error {
	// 1. Definisikan struct penampung request JSON (Sesuai payload frontend Anda)
	var p Payload

	// Parsing data JSON yang dikirim oleh frontend fetch
	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid: " + err.Error(),
		})
	}

	// 2. Validasi Input Dasar
	if p.CustomerName == "" || p.PlanType == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Nama dan Paket wajib diisi!",
		})
	}

	// 3. LOGIKA PEMISAHAN PAKET
	if p.PlanType == "TRIAL" {
		// --- JIKA TRIAL: Langsung Generate & Aktifkan ---
		// Kita gunakan fungsi generateUniqueToken() yang sudah ada di paket_licensi.go Anda
		token := generateUniqueToken()

		newToken := models.License{
			Token:         token,
			CustomerName:  p.CustomerName,
			CustomerEmail: p.CustomerEmail,
			CustomerPhone: p.CustomerPhone,
			PlanType:      p.PlanType,
			Status:        "ACTIVE", // Langsung aktif karena trial
			InvoiceID:     p.InvoiceID,
		}

		if err := database.DB.Create(&newToken).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Gagal menyimpan token trial: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"token":   token,
			"message": "Token Trial berhasil diaktifkan!",
		})
	}

	// --- JIKA BUKAN TRIAL: Mintakan Token Snap ke Midtrans ---
	// Menggunakan fungsi getSnapTokenFromMidtrans yang ada di utils.go kita kemarin
	snapToken := getSnapTokenFromMidtrans(p)

	if snapToken == "" {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mendapatkan token pembayaran dari Midtrans",
		})
	}

	// Mengembalikan snap_token agar frontend memunculkan pop-up pembayaran Midtrans
	return c.JSON(fiber.Map{
		"status":     "success",
		"snap_token": snapToken,
		"message":    "Silakan selesaikan pembayaran Anda",
	})
}

//End

// Aksi Blokir Token Kustomer Nakal
func BlockTokenAction(c *fiber.Ctx) error {
	id := c.Params("id")
	var license models.License
	if err := database.DB.First(&license, id).Error; err == nil {
		license.Status = "blocked"
		database.DB.Save(&license)
	}
	return c.Redirect("/admin/dashboard")
}

// Aksi Reset HWID (Jika Kustomer ganti komputer/meja packing baru)
func ResetHwidAction(c *fiber.Ctx) error {
	id := c.Params("id")
	var license models.License
	if err := database.DB.First(&license, id).Error; err == nil {
		license.HWID = ""           // SEKARANG SUDAH FIX KAPITAL (HWID)
		license.Status = "inactive" // Kembalikan ke inactive agar bisa mengunci HWID baru lagi
		database.DB.Save(&license)
	}
	return c.Redirect("/admin/dashboard")
}

// Log Out Admin
func LogOut(c *fiber.Ctx) error {
	c.ClearCookie("admin_session")
	return c.Redirect("/admin/login")
}
