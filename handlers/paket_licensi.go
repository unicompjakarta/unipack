package handlers

import (
	"backend-golang/database"
	"backend-golang/models"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"net/http"
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"

	"github.com/gofiber/fiber/v2"
)

type Payload struct {
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
	PlanType      string `json:"plan_type"`
	InvoiceID     string `json:"invoice_id"`
}

// ==========================================
// 1. LANDING PAGE HANDLER
// ==========================================

// GetIndexPage merender halaman utama landing page Unipacking
func GetIndexPage(c *fiber.Ctx) error {
	var packets []models.Packet
	database.DB.Find(&packets)

	// Render file HTML index sambil melempar data packets dari database
	return c.Render("index", fiber.Map{
		"Packets": packets,
	})
}

// ==========================================
// 2. MANAGEMENT PACKET HANDLERS
// ==========================================

// CreatePacketHandler menangani pembuatan paket lisensi baru
func CreatePacketHandler(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.FormValue("active_days"))
	price, _ := strconv.ParseInt(c.FormValue("price"), 10, 64)

	newPacket := models.Packet{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		ActiveDays:  days,
		Price:       price,
		Note:        c.FormValue("note"),
		Type:        c.FormValue("type"), // "BEST_SELLER" atau kosong
	}

	if err := database.DB.Create(&newPacket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menyimpan paket")
	}

	return c.Redirect("/admin/dashboard")
}

// UpdatePacketHandler menangani perubahan data paket berdasarkan form dashboard admin
func UpdatePacketHandler(c *fiber.Ctx) error {
	idStr := c.FormValue("id") // Menangkap input id hidden
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("ID Paket tidak valid")
	}

	var packet models.Packet
	if err := database.DB.First(&packet, uint(id)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Paket tidak ditemukan di database")
	}

	days, errAtoi := strconv.Atoi(c.FormValue("active_days"))
	if errAtoi != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Masa aktif harus berupa angka")
	}

	price, errPrice := strconv.ParseInt(c.FormValue("price"), 10, 64)
	if errPrice != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Harga paket harus berupa angka")
	}

	// Sinkronisasi data baru
	packet.Name = c.FormValue("name")
	packet.Description = c.FormValue("description")
	packet.ActiveDays = days
	packet.Price = price
	packet.Note = c.FormValue("note")
	packet.Type = c.FormValue("type")
	packet.UpdatedAt = time.Now()

	if err := database.DB.Save(&packet).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal memperbarui data paket")
	}

	return c.Redirect("/admin/dashboard")
}

// DeletePacketHandler menangani penghapusan paket berdasarkan ID
func DeletePacketHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("ID Paket tidak valid")
	}

	if err := database.DB.Delete(&models.Packet{}, uint(id)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menghapus data paket")
	}

	return c.Redirect("/admin/dashboard")
}

// GetAllPacketsAPI mengembalikan seluruh paket dalam bentuk JSON jika dibutuhkan frontend
func GetAllPacketsAPI(c *fiber.Ctx) error {
	var packets []models.Packet
	database.DB.Find(&packets)
	return c.JSON(packets)
}

// ==========================================
// 3. MANAGEMENT LICENSE HANDLERS
// ==========================================

// UpdateLicenseHandler menangani perubahan data lisensi kustomer dari form dashboard admin
func UpdateLicenseHandler(c *fiber.Ctx) error {
	idStr := c.FormValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "ID tidak valid"})
	}

	var license models.License
	if err := database.DB.First(&license, uint(id)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Data tidak ditemukan"})
	}

	// SESUAIKAN NAMA DENGAN FORM HTML BOS
	license.CustomerName = c.FormValue("customer_name")
	license.CustomerEmail = c.FormValue("customer_email")
	license.CustomerPhone = c.FormValue("customer_phone")
	license.PlanType = c.FormValue("plan_type")

	if err := database.DB.Save(&license).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Gagal update"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Berhasil diupdate!"})
}

// DeleteLicenseHandler menangani penghapusan lisensi berdasarkan ID dari URL parameter
func DeleteLicenseHandler(c *fiber.Ctx) error {
	idStr := c.Params("id") // Menangkap parameter :id dari routing app.Post
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("ID Lisensi tidak valid")
	}

	var license models.License
	if err := database.DB.First(&license, uint(id)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Data lisensi tidak ditemukan")
	}

	if err := database.DB.Delete(&license).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal menghapus data lisensi")
	}

	return c.Redirect("/admin/dashboard")
}

// Blok Midtrans
func GetSnapTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Inisialisasi Midtrans
	var s = snap.Client{}
	s.New(os.Getenv("SERVER_KEY"), midtrans.Sandbox) // Ganti ke Production nanti

	// 2. Buat Request Snap
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  "ORDER-" + generateRandomString(8), // Wajib unik
			GrossAmt: 50000,                              // Ambil dari harga paket yang dipilih
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: "Nama Customer",
			Email: "email@customer.com",
		},
	}

	// 3. Panggil API Midtrans
	snapResp, err := s.CreateTransaction(req)
	if err != nil {
		http.Error(w, "Gagal membuat transaksi", 500)
		return
	}

	// 4. Kirim Token kembali ke Frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"snap_token": snapResp.Token})
}

func generateUniqueToken() string {
	// Membuat random bytes
	bytes := make([]byte, 8)
	rand.Read(bytes)

	// Format: UNI-[JENIS]-[RANDOM_HEX]
	// Kita bisa ambil waktu saat ini agar lebih unik
	timestamp := time.Now().Format("060102")
	return fmt.Sprintf("UNI-%s-%s", timestamp, hex.EncodeToString(bytes))
}

func HandleCheckout(w http.ResponseWriter, r *http.Request) {
	// 1. Decode payload dari frontend
	var p Payload
	json.NewDecoder(r.Body).Decode(&p)

	// 2. LOGIKA PEMISAHAN PAKET
	if p.PlanType == "TRIAL" {
		// Langsung generate lisensi tanpa Midtrans
		token := generateUniqueToken()
		saveToDatabase(p, token, "ACTIVE") // Status langsung aktif

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"token":   token,
			"message": "Token Trial berhasil diaktifkan!",
		})
		return
	}

	// 3. JIKA BUKAN TRIAL: Gunakan Midtrans
	// Logic untuk mendapatkan Snap Token
	snapToken := getSnapTokenFromMidtrans(p)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"snap_token": snapToken,
		"message":    "Silakan selesaikan pembayaran",
	})
}
