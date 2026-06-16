package handlers

import (
	"backend-golang/database"
	"backend-golang/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Endpoint yang ditembak oleh Aplikasi Desktop (.exe) Python
func CheckLicense(c *fiber.Ctx) error {
	type LicenseRequest struct {
		Token string `json:"token"`
		Hwid  string `json:"hwid"`
	}

	var req LicenseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Invalid request payload"})
	}

	var license models.License
	// Cari token di database
	if err := database.DB.Where("token = ?", req.Token).First(&license).Error; err != nil {
		return c.Status(200).JSON(fiber.Map{"status": "invalid", "message": "Token tidak ditemukan"})
	}

	// =================================================================
	// 1. VALIDASI HWID DULUAN (Paling Aman)
	// =================================================================
	// Jika token sudah aktif tapi HWID yang menembak tidak sama dengan yang terdaftar
	if license.Status == "active" && license.HWID != "" && license.HWID != req.Hwid {
		return c.Status(200).JSON(fiber.Map{"status": "locked", "message": "Token sudah terkunci di PC lain!"})
	}

	// =================================================================
	// 2. PROSES AKTIVASI JIKA MASIH BARU (INACTIVE)
	// =================================================================
	if license.Status == "inactive" || license.HWID == "" {
		license.HWID = req.Hwid
		license.Status = "active"

		// Tentukan masa aktif berdasarkan PlanType saat checkout web
		now := time.Now()
		license.ActivatedAt = &now

		var expiredAt time.Time
		if license.PlanType == "TRIAL" {
			expiredAt = now.AddDate(0, 0, 7)
		} else { // BULANAN / MONTHLY / YEARLY
			expiredAt = now.AddDate(0, 1, 0)
		}
		license.ExpiredAt = &expiredAt

		// Simpan perubahan aktivasi awal
		if err := database.DB.Save(&license).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Gagal menyimpan aktivasi"})
		}
	}

	// =================================================================
	// 3. JIKA LOLOS SEMUA, UPDATE DATA SINKRONISASI
	// =================================================================
	now := time.Now()
	// Gunakan struct update agar lebih aman dan efisien
	database.DB.Model(&license).Update("last_sync_time", now)

	// Kirim response sukses beserta tanggal expired_at asli dari server VPS
	return c.Status(200).JSON(fiber.Map{
		"status":     "active",
		"expired_at": license.ExpiredAt,
	})
}
