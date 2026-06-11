package handlers

import (
	"backend-golang/database"
	"backend-golang/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Endpoint yang ditembak oleh Aplikasi Desktop (.exe) Python
func CheckLicense(c *fiber.Ctx) error {
	type RequestBody struct {
		Token string `json:"token"`
		Hwid  string `json:"hwid"`
	}

	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Bad Request"})
	}

	var license models.License
	// Cari token di database
	if err := database.DB.Where("token = ?", body.Token).First(&license).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"status": "invalid", "message": "Token tidak ditemukan!"})
	}

	// 1. Cek Status Blocked
	if license.Status == "blocked" {
		return c.Status(403).JSON(fiber.Map{"status": "blocked", "message": "Token diblokir oleh Owner!"})
	}

	// 2. Cek Kedaluwarsa
	if license.Status == "active" && time.Now().After(license.ExpiredAt) {
		license.Status = "expired"
		database.DB.Save(&license)
		return c.Status(401).JSON(fiber.Map{"status": "expired", "message": "Masa aktif token telah habis!"})
	}

	// 3. FASE AKTIVASI PERTAMA KALI (Jika status masih inactive)
	if license.Status == "inactive" {
		now := time.Now()
		license.Hwid = body.Hwid
		license.Status = "active"
		license.ActivatedAt = &now

		// Jika saat checkout/generate durasi expired belum diset, set otomatis di sini
		if license.ExpiredAt.IsZero() {
			if license.PlanType == "TRIAL" {
				license.ExpiredAt = now.AddDate(0, 0, 7)
			} else if license.PlanType == "MONTHLY" {
				license.ExpiredAt = now.AddDate(0, 1, 0)
			} else if license.PlanType == "YEARLY" {
				license.ExpiredAt = now.AddDate(1, 0, 0)
			}
		}

		database.DB.Save(&license)
		return c.JSON(fiber.Map{"status": "active", "message": "Aktivasi sukses di PC pertama!"})
	}

	// 4. FASE PENGECEKAN RUTIN (Harus mencocokkan HWID)
	if license.Hwid != body.Hwid {
		return c.Status(403).JSON(fiber.Map{"status": "failed", "message": "Token sudah terikat di perangkat lain!"})
	}

	return c.JSON(fiber.Map{"status": "active", "message": "Lisensi terverifikasi aktif."})
}
