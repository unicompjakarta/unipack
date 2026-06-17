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
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Payload tidak valid"})
	}

	var license models.License
	// 1. Cek keberadaan token di database
	if err := database.DB.Where("token = ?", req.Token).First(&license).Error; err != nil {
		return c.Status(200).JSON(fiber.Map{"status": "invalid", "message": "Token tidak terdaftar"})
	}

	// 2. KUNCI HWID (Jika token sudah terikat ke PC lain)
	if license.HWID != "" && license.HWID != req.Hwid {
		return c.Status(200).JSON(fiber.Map{"status": "locked", "message": "Token sudah terkunci di perangkat lain!"})
	}

	// 3. AKTIVASI PERTAMA KALI
	if license.Status == "inactive" || license.HWID == "" {

		// --- CEK APAKAH HWID SUDAH PUNYA TOKEN AKTIF LAIN ---
		var existingLicense models.License
		err := database.DB.Where("hwid = ? AND status = ? AND token != ?", req.Hwid, "active", req.Token).First(&existingLicense).Error
		if err == nil {
			return c.Status(200).JSON(fiber.Map{"status": "locked", "message": "Perangkat ini sudah terdaftar dengan token lain!"})
		}
		// ----------------------------------------------------

		license.HWID = req.Hwid
		license.Status = "active"
		now := time.Now()
		license.ActivatedAt = &now

		// Tentukan Durasi
		duration := 30 * 24 * time.Hour
		if license.PlanType == "TRIAL" {
			duration = 7 * 24 * time.Hour
		}
		expiredAt := now.Add(duration)
		license.ExpiredAt = &expiredAt

		if err := database.DB.Save(&license).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Gagal menyimpan ke database"})
		}

		return c.Status(200).JSON(fiber.Map{"status": "active", "expired_at": license.ExpiredAt, "message": "Aktivasi berhasil!"})
	}

	// 4. VALIDASI STATUS AKTIF & KEDALUWARSA
	if license.Status != "active" {
		return c.Status(200).JSON(fiber.Map{"status": "blocked", "message": "Lisensi tidak aktif/diblokir"})
	}

	if license.ExpiredAt != nil && time.Now().After(*license.ExpiredAt) {
		license.Status = "expired"
		database.DB.Save(&license)
		return c.Status(200).JSON(fiber.Map{"status": "expired", "message": "Masa aktif telah habis"})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":     "active",
		"expired_at": license.ExpiredAt,
		"message":    "Lisensi valid",
	})
}
