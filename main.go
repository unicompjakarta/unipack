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

	// --- ROUTING API (Ditembak oleh Desktop App & Web Checkout Front-end) ---
	app.Post("/api/v1/check-license", handlers.CheckLicense)
	app.Post("/api/v1/license/generate", handlers.GenerateTokenAction) // Endpoint checkout-ready

	// --- ROUTING WEB DASHBOARD (Ditembak oleh Browser Owner) ---
	app.Get("/admin/login", handlers.GetLogin)
	app.Post("/admin/login", handlers.PostLogin)
	app.Get("/admin/dashboard", handlers.GetDashboard)
	app.Post("/admin/license/generate", handlers.GenerateTokenAction)
	app.Post("/admin/license/block/:id", handlers.BlockTokenAction)
	app.Post("/admin/license/reset/:id", handlers.ResetHwidAction)
	app.Get("/admin/logout", handlers.LogOut)

	// --- ROUTING FRONT-END CHECKOUT (Disajikan langsung oleh Golang) ---
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	log.Println("Server Golang berjalan otonom di port 3000, Bos!")
	log.Fatal(app.Listen(":3000"))
}
