// package database

// import (
// 	"compro/backend-golang/models"
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/joho/godotenv"
// 	"gorm.io/driver/mysql"
// 	"gorm.io/gorm"
// )

// var DB *gorm.DB

// func ConnectDB() {
// 	// 1. Load file .env jika ada (di lokal maupun di VPS)
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Println("ℹ️ File .env tidak ditemukan, sistem akan membaca Environment OS.")
// 	}

// 	// 2. Ambil data dari .env secara dinamis, jika kosong gunakan default lokal
// 	username := getEnv("DB_USERNAME", "root")
// 	password := getEnv("DB_PASSWORD", "admin")
// 	host := getEnv("DB_HOST", "127.0.0.1")
// 	port := getEnv("DB_PORT", "3306")
// 	dbName := getEnv("DB_NAME", "db_unipack_lokal")

// 	// 3. Sambungkan ke MySQL Server
// 	dsnBase := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port)
// 	dbRaw, err := gorm.Open(mysql.Open(dsnBase), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("❌ Gagal terkoneksi ke MySQL Server: ", err)
// 	}

// 	// 4. Auto-create database jika belum ada
// 	createDbQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)
// 	err = dbRaw.Exec(createDbQuery).Error
// 	if err != nil {
// 		log.Fatal("❌ Gagal membuat database otomatis: ", err)
// 	}

// 	// 5. Masuk ke database utama
// 	dsnFinal := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
// 	db, err := gorm.Open(mysql.Open(dsnFinal), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("❌ Gagal masuk ke database "+dbName+": ", err)
// 	}

// 	log.Printf("🚀 Database '%s' Berhasil Terkoneksi.\n", dbName)

// 	// 6. GORM AutoMigrate
// 	err = db.AutoMigrate(&models.License{}, &models.Packet{}, &models.Menu{}, &models.Page{}, &models.PageComponent{},
// 		&models.Announcement{}, &models.WebProfile{}, &models.WebHeaderImage{}, &models.Customer{},
// 		&models.LaptopBrand{},
// 		&models.LaptopType{},
// 		&models.SerialLaptop{},
// 		&models.CompatibleSparepart{},
// 		&models.Consultation{}, &models.Category{}, &models.Symptom{},&models.Jasa{}, &models.LayananCategory{}, &models.Product{}, &models.ProductCategory{},&models.AppSetting{}  )
// 	if err != nil {
// 		log.Fatal("❌ Gagal melakukan AutoMigrate tabel: ", err)
// 	}

// 	DB = db
// }

// // Fungsi pembantu untuk membaca env dengan nilai default jika kosong
// func getEnv(key, fallback string) string {
// 	if value, exists := os.LookupEnv(key); exists {
// 		return value
// 	}
// 	return fallback
// }

package database

import (
	"compro/backend-golang/config"
	"compro/backend-golang/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Menyiapkan 2 Variable Global Pointer DB
var (
	DB     *gorm.DB // Konek ke DB Utama (db_unipack)
	DBNuxt *gorm.DB // Konek ke DB Nuxt (db_nuxt)
)

func ConnectDB() {
	// 1. Load Konfigurasi
	cfg := config.LoadConfig()

	// 2. Konek ke MySQL Server (Base DSN)
	dsnBase := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
	
	dbRaw, err := gorm.Open(mysql.Open(dsnBase), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal terkoneksi ke MySQL Server: ", err)
	}

	// 3. Auto-Create Kedua Database jika belum ada
	createDatabaseIfNotExist(dbRaw, cfg.DBName)
	createDatabaseIfNotExist(dbRaw, cfg.DBNuxtName)

	// 4. Konek ke Database Utama (db_unipack)
	dsnMain := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	
	dbMainInstance, err := gorm.Open(mysql.Open(dsnMain), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal masuk ke database "+cfg.DBName+": ", err)
	}
	log.Printf("🚀 Database Utama '%s' Berhasil Terkoneksi.\n", cfg.DBName)

	// 5. Konek ke Database Khusus Nuxt (db_nuxt)
	dsnNuxt := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBNuxtName)
	
	dbNuxtInstance, err := gorm.Open(mysql.Open(dsnNuxt), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal masuk ke database "+cfg.DBNuxtName+": ", err)
	}
	log.Printf("🚀 Database Nuxt '%s' Berhasil Terkoneksi.\n", cfg.DBNuxtName)

	// 🟢 DB Utama (db_unipack): Khusus License & Packet
	err = dbMainInstance.AutoMigrate(
		&models.License{}, 
		&models.Packet{},
	)
	if err != nil {
		log.Fatal("❌ Gagal AutoMigrate DB Utama: ", err)
	}

	// 🔵 DB Nuxt (db_nuxt)
	err = dbNuxtInstance.AutoMigrate(
		// 1. Tabel Induk / Independent terlebih dahulu
		&models.Announcement{},
		&models.AppSetting{},
		&models.Category{},
		&models.Customer{},
		&models.Jasa{},
		&models.LaptopBrand{}, // Induk dari LaptopType
		&models.LayananCategory{},
		&models.Menu{},
		&models.Page{},
		&models.PageComponent{},
		&models.ProductCategory{},
		&models.Symptom{},
		&models.WebHeaderImage{},
		&models.WebProfile{},

		// 2. Tabel Relasi Tingkat 1
		&models.LaptopType{},  // Butuh LaptopBrand (dibuat setelah LaptopBrand)
		&models.Product{},     // Butuh ProductCategory

		// 3. Tabel Relasi Tingkat 2 (Anak dari LaptopType)
		&models.CompatibleSparepart{}, // Butuh LaptopType
		&models.SerialLaptop{},        // Butuh LaptopType
		&models.Consultation{},
	)
	if err != nil {
		log.Fatal("❌ Gagal AutoMigrate DB Nuxt: ", err)
	}

	// 7. Assign ke variable global
	DB = dbMainInstance
	DBNuxt = dbNuxtInstance
}

// Helper untuk membuat DB jika belum ada
func createDatabaseIfNotExist(dbRaw *gorm.DB, dbName string) {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)
	if err := dbRaw.Exec(query).Error; err != nil {
		log.Fatalf("❌ Gagal membuat database %s otomatis: %v", dbName, err)
	}
}
