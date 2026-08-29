package main

import (
	"compro/backend-golang/database" // Harus diawali dengan nama modul di go.mod
	"compro/backend-golang/handlers"
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

	app.Static("/uploads", "./uploads")

	//upload image page

	// Route ke LoginHandler
	app.Post("/admin/login", handlers.LoginHandler)
	
	

	// 4. REGISTER ENDPOINT CMS BARU UNTUK NUXT
	cmsHandler := handlers.NewCMSHandler(database.DBNuxt)

	api := app.Group("/api")

	// Endpoint navigasi menu FE
	api.Post("/menus", cmsHandler.CreateMenu)
	api.Get("/menus", cmsHandler.GetMenus)

	// Endpoint sub menu
	api.Get("/navbar", cmsHandler.GetNavbarMenus)

	// Endpoint footer
	//app.Get("/api/menus", cmsHandler.GetFooterMenus)
	//upload ima
	api.Post("/uploads", cmsHandler.UploadImage)

	// Endpoint builder page dinamis
	api.Post("/pages", cmsHandler.CreatePage)

	// 1. Ambil SEMUA halaman untuk tabel admin Bos (Menggunakan fungsi baru)
	api.Get("/pages", cmsHandler.GetAllPages)

	// 2. Ambil SATU halaman berdasarkan parameter slug (Fungsi catch-all milik Bos)
	// api.Get("/pages/*", cmsHandler.GetPageBySlug) yang lama
	api.Get("/pages/:slug", cmsHandler.GetPageBySlug)

	api.Put("/pages/:id", cmsHandler.UpdatePage)
	api.Delete("/pages/:id", cmsHandler.DeletePage)

	//END Blok APP Web

	// Route Single Row Konfigurasi Web Profile
	app.Get("/web-profile", cmsHandler.GetWebProfile)
	app.Post("/web-profile", cmsHandler.SaveWebProfile)

	// Endpoint Pengumuman

	


	//Endpoint Konsultasi
	app.Post("/consultation", cmsHandler.CreateConsultation)
    app.Get("/admin/consultations", cmsHandler.GetConsultations)
	app.Put("/admin/consultations/:id/status", cmsHandler.UpdateConsultationStatus)
    app.Delete("/admin/consultations/:id", cmsHandler.DeleteConsultation)

	// ==================== APP ROUTING REGISTRATION ====================
	
	// API Master Brands
	app.Get("/admin/brands", cmsHandler.GetBrands)
	app.Post("/admin/brands", cmsHandler.CreateBrand)
	app.Put("/admin/brands/:id", cmsHandler.UpdateBrand)
	app.Delete("/admin/brands/:id", cmsHandler.DeleteBrand)

	// API Master Categories
	app.Get("/admin/categories", cmsHandler.GetCategories)

	// API Master Symptoms (Kerusakan)
	app.Get("/admin/symptoms", cmsHandler.GetSymptoms)
	app.Post("/admin/symptoms", cmsHandler.CreateSymptom)
	app.Delete("/admin/symptoms/:id", cmsHandler.DeleteSymptom)

	// Blok Jasa - Daftarkan route baru di bawah instance cmsHandler yang sudah ada:
	app.Get("/admin/jasas", cmsHandler.GetJasas)
	app.Post("/admin/jasas", cmsHandler.CreateJasa)
	app.Put("/admin/jasas/:id", cmsHandler.UpdateJasa)
	app.Delete("/admin/jasas/:id", cmsHandler.DeleteJasa)

	app.Get("/admin/jasa-categories", cmsHandler.GetCategories)

	//Layanan Kategori
	app.Get("/layanancategory", cmsHandler.GetLayananCategories)

	//Cek estimasi harga
	app.Post("/public/estimasi", cmsHandler.CreatePublicConsultation)

	//Blok Upload Foto
	// app.Get("/api/announcements/active", cmsHandler.GetActiveAnnouncement)
	// app.Post("/api/announcements", cmsHandler.SaveAnnouncement)
	// app.Put("/api/announcements/:id/toggle", cmsHandler.ToggleActiveAnnouncement)
	// app.Delete("/api/announcements/:id", cmsHandler.DeleteAnnouncement)
	app.Get("/api/announcements", cmsHandler.GetAnnouncements)
	app.Get("/api/announcements/active", cmsHandler.GetActiveAnnouncement) // 🔥 Sudah beres Bos!
	app.Post("/api/announcements", cmsHandler.SaveAnnouncement)
	app.Put("/api/announcements/:id/toggle", cmsHandler.ToggleActiveAnnouncement)
	app.Delete("/api/announcements/:id", cmsHandler.DeleteAnnouncement)
	//End Upload


	// ==================== APP SECTION PRODUK ====================
	api.Get("/categories", cmsHandler.GetProductCategories) 
    api.Post("/categories", cmsHandler.CreateCategory)
    // Products Endpoints
	api.Get("/products", cmsHandler.GetProducts)
	api.Post("/products", cmsHandler.CreateProduct)
	api.Put("/products/:id", cmsHandler.UpdateProduct)
	api.Delete("/products/:id", cmsHandler.DeleteProduct)


	// ==================== SWITCH ON OFF SECTION ====================
	api.Get("/product-section", cmsHandler.GetProductSectionStatus)
    api.Post("/product-section", cmsHandler.UpdateProductSectionStatus)





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
