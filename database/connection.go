// package database

// import (
// 	"backend-golang/models"
// 	"fmt"
// 	"log"

// 	"gorm.io/driver/mysql"
// 	"gorm.io/gorm"
// )

// // Wadah koneksi database global
// var DB *gorm.DB

// func ConnectDB() {
// 	var err error

// 	// 1. Kredensial MySQL Lokal (Default XAMPP: user 'root', password kosong '')
// 	username := "root"
// 	password := "admin"
// 	host := "127.0.0.1"
// 	port := "3306"
// 	dbName := "db_unipack_lokal"

// 	// 2. Konek ke MySQL tanpa sebut nama database dulu (untuk check/create otomatis)
// 	dsnBase := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port)
// 	dbRaw, err := gorm.Open(mysql.Open(dsnBase), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("❌ Gagal terkoneksi ke MySQL Server lokal: ", err)
// 	}

// 	// 3. Eksekusi perintah SQL untuk membuat database otomatis jika belum ada
// 	createDbQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)
// 	err = dbRaw.Exec(createDbQuery).Error
// 	if err != nil {
// 		log.Fatal("❌ Gagal membuat database otomatis: ", err)
// 	}

// 	// 4. Setelah dipastikan database ada, koneksikan ulang langsung ke database tersebut
// 	dsnFinal := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
// 	db, err := gorm.Open(mysql.Open(dsnFinal), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("❌ Gagal masuk ke database "+dbName+": ", err)
// 	}

// 	log.Printf("🚀 Mantap, Bos! Database '%s' Berhasil Terkoneksi (MySQL Lokal).\n", dbName)

// 	// 5. GORM otomatis membuat tabel 'licenses' berdasarkan struct models
// 	err = db.AutoMigrate(&models.License{})
// 	if err != nil {
// 		log.Fatal("❌ Gagal melakukan AutoMigrate tabel: ", err)
// 	}
// 	log.Println("⚡ Tabel berhasil di-migrate otomatis oleh GORM.")

//		// Masukkan ke variabel global agar bisa dipakai di handler/routes
//		DB = db
//	}
package database

import (
	"backend-golang/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// 1. Load file .env jika ada (di lokal maupun di VPS)
	err := godotenv.Load()
	if err != nil {
		log.Println("ℹ️ File .env tidak ditemukan, sistem akan membaca Environment OS.")
	}

	// 2. Ambil data dari .env secara dinamis, jika kosong gunakan default lokal
	username := getEnv("DB_USERNAME", "root")
	password := getEnv("DB_PASSWORD", "admin")
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "db_unipack_lokal")

	// 3. Sambungkan ke MySQL Server
	dsnBase := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port)
	dbRaw, err := gorm.Open(mysql.Open(dsnBase), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal terkoneksi ke MySQL Server: ", err)
	}

	// 4. Auto-create database jika belum ada
	createDbQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)
	err = dbRaw.Exec(createDbQuery).Error
	if err != nil {
		log.Fatal("❌ Gagal membuat database otomatis: ", err)
	}

	// 5. Masuk ke database utama
	dsnFinal := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
	db, err := gorm.Open(mysql.Open(dsnFinal), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal masuk ke database "+dbName+": ", err)
	}

	log.Printf("🚀 Database '%s' Berhasil Terkoneksi.\n", dbName)

	// 6. GORM AutoMigrate
	err = db.AutoMigrate(&models.License{}, &models.Packet{})
	if err != nil {
		log.Fatal("❌ Gagal melakukan AutoMigrate tabel: ", err)
	}

	DB = db
}

// Fungsi pembantu untuk membaca env dengan nilai default jika kosong
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
