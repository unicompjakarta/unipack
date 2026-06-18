package main

import (
	"backend-golang/database"
	"backend-golang/handlers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// 1. Inisialisasi Database GORM
	database.ConnectDB()

	// 2. Setup Template Engine untuk UI Dashboard HTML
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// 3. Tambahkan CORS agar Front-End Web Checkout (React/Next.js) Anda bebas menembak API ini tanpa terblokir browser
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Tambahkan ini agar folder fisik di VPS bisa diakses via URL /downloads
	app.Static("/downloads", "/var/www/unipack/downloads")

	// --- ROUTING API (Ditembak oleh Desktop App & Web Checkout Front-end) ---
	app.Post("/api/v1/check-license", handlers.CheckLicense)
	app.Post("/api/v1/license/generate", handlers.GenerateTokenAction) // Endpoint checkout-ready

	app.Get("/admin/dashboard", handlers.GetDashboard)
	app.Post("/admin/license/update", handlers.UpdateLicenseHandler)
	app.Post("/admin/license/delete/:id", handlers.DeleteLicenseHandler)

	// --- ROUTING WEB DASHBOARD (Ditembak oleh Browser Owner) ---
	app.Get("/admin/login", handlers.GetLogin)
	app.Post("/admin/login", handlers.PostLogin)
	app.Get("/admin/dashboard", handlers.GetDashboard)
	app.Post("/admin/license/generate", handlers.GenerateTokenAction)
	app.Post("/admin/license/block/:id", handlers.BlockTokenAction)
	app.Post("/admin/license/reset/:id", handlers.ResetHwidAction)
	app.Get("/admin/logout", handlers.LogOut)

	// --- ROUTING PACKET MANAGEMENT (DATABASE OPERATIONAL) ---
	app.Post("/admin/packet/create", handlers.CreatePacketHandler)
	app.Post("/admin/packet/update", handlers.UpdatePacketHandler)
	app.Post("/admin/packet/delete/:id", handlers.DeletePacketHandler)
	app.Get("/api/v1/packets", handlers.GetAllPacketsAPI) // Endpoint publik opsional

	// --- ROUTING FRONT-END CHECKOUT (Disajikan langsung oleh Golang) ---
	// KODE BARU (Arahkan langsung ke fungsi yang sudah Anda buat di admin_panel.go)
	app.Get("/", handlers.GetIndexPage)

	//Checkout

	app.Post("/api/get-snap-token", handlers.GenerateTokenAction)

	// 🚀 RUTE BARU: Callback/Webhook Penangkap QRIS Sukses dari Midtrans
	app.Post("/api/midtrans-callback", handlers.MidtransCallbackHandler)

	log.Println("Server Golang berjalan otonom di port 3000, Bos!")
	log.Fatal(app.Listen(":3000"))
}
