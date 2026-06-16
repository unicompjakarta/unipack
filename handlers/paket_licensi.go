package handlers

import (
	"backend-golang/database"
	"backend-golang/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

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
	idStr := c.FormValue("id") // Menangkap input id hidden dari form edit lisensi
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("ID Lisensi tidak valid")
	}

	var license models.License
	if err := database.DB.First(&license, uint(id)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Data lisensi tidak ditemukan")
	}

	// Update data field sesuai input form edit lisensi Bos
	license.CustomerName = c.FormValue("name")
	license.CustomerEmail = c.FormValue("email")
	license.CustomerPhone = c.FormValue("phone")
	license.PlanType = c.FormValue("plan") // TRIAL / BULANAN / YEARLY
	license.UpdatedAt = time.Now()

	if err := database.DB.Save(&license).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Gagal memperbarui data lisensi")
	}

	return c.Redirect("/admin/dashboard")
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
