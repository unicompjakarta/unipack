package database

import (
	"backend-golang/models"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TAMBAHKAN BARIS INI (Untuk wadah koneksi database global)
var DB *gorm.DB

func ConnectDB() {
	db, err := gorm.Open(sqlite.Open("licenses.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database: ", err)
	}

	log.Println("Database berhasil tersambung (Pure Go SQLite).")

	db.AutoMigrate(&models.License{})

	// Sekarang baris ini tidak akan error lagi karena variabel DB sudah ada di atas
	DB = db
}
