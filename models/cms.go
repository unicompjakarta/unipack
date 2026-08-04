package models

import (
	"time"

	"gorm.io/gorm"
)

// Menu: Menentukan posisi penempatan di Front-End (Navbar, Body, Footer)
type Menu struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ParentID  *uint     `gorm:"default:null" json:"parent_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Slug      string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Position  string    `gorm:"type:varchar(50);not null" json:"position"` // 'navbar', 'body', 'footer'
	Order     int       `gorm:"default:0" json:"order"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	ImageUrl  string    `gorm:"type:varchar(255);default:null" json:"image_url"`
	Pages     []Page    `gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;" json:"pages,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 💡 RELASI GORM UNTUK REKURSIF SUBMENU

	Submenus []Menu `gorm:"-" json:"submenus"`
}

// Page: Konten utama halaman dinamis
type Page struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	//MenuID      uint            `gorm:"not null;index" json:"menu_id"`
	MenuID      *uint  `gorm:"default:null;index" json:"menu_id"`
	Title       string          `gorm:"type:varchar(255);not null" json:"title"`
	BreadcrumbTitle string     `gorm:"type:varchar(255)" json:"breadcrumb_title"`
	Slug        string          `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Category    string          `gorm:"type:varchar(100);not null" json:"category"` // 'landing_page', 'article', dll
	TemplateKey string          `gorm:"type:varchar(100);default:'default'" json:"template_key"`
	Components  []PageComponent `gorm:"foreignKey:PageID;constraint:OnDelete:CASCADE;" json:"components"`

	// 🎯 TAMBAHKAN TRINITY SEO INI SECARA LENGKAP, BOS!
	SeoTitle       string `gorm:"type:varchar(255)" json:"seo_title"`
	SeoDescription string `gorm:"type:text" json:"seo_description"`
	SeoKeywords    string `gorm:"type:varchar(255)" json:"seo_keywords"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PageComponent: Menyimpan sekat/section di dalam page secara dinamis (JSON text)
type PageComponent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PageID    uint      `gorm:"not null;index" json:"page_id"`
	Order     int       `gorm:"default:0" json:"order"`
	Type      string    `gorm:"type:varchar(100);not null" json:"type"` // 'hero_banner', 'features_grid', dll
	Content   string    `gorm:"type:text" json:"content"`               // JSON String dari Nuxt Admin Form
	Styles    string    `gorm:"type:text" json:"styles"`                // JSON String untuk Tailwind Object styling
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Announcement struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Type       string    `gorm:"type:varchar(10);default:'text'" json:"type"`     // "text" atau "image"
	Content    string    `gorm:"type:text" json:"content"`                        // Teks pengumuman
	ImageURL   string    `gorm:"type:varchar(255)" json:"image_url"`              // Path file gambar
	Placement  string    `gorm:"type:varchar(20);default:'all'" json:"placement"` // "all", "index", "slug", "specific"
	SlugTarget string    `gorm:"type:varchar(100)" json:"slug_target"`            // Target URL spesifik
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WebProfile (Single Row ID = 1)
type WebProfile struct {
	ID               uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyName      string           `gorm:"type:varchar(100);not null" json:"company_name"`
	LogoURL          string           `gorm:"type:varchar(255)" json:"logo_url"`
	Email            string           `gorm:"type:varchar(100)" json:"email"`
	AboutUs          string           `gorm:"type:text" json:"about_us"`
	OperationalHours string           `gorm:"type:text" json:"operational_hours"` // Diubah jadi text untuk multi-line
	DetailAddress    string           `gorm:"type:text" json:"detail_address"`
	WhatsAppNumber   string           `gorm:"type:varchar(20)" json:"whatsapp_number"`
	HeaderImages     []WebHeaderImage `gorm:"foreignKey:WebProfileID;constraint:OnDelete:CASCADE;" json:"header_images"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// WebHeaderImage menampung multi-image untuk slider header
type WebHeaderImage struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	WebProfileID uint   `json:"web_profile_id"`
	ImageURL     string `gorm:"type:varchar(255);not null" json:"image_url"`
	LinkTarget   string `gorm:"type:varchar(255)" json:"link_target"` // Link saat slider diklik
}

// LaptopBrand mewakili model LaptopBrand
type LaptopBrand struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Image     string         `gorm:"type:varchar(255);default:''" json:"image"`
	Types         []LaptopType   `gorm:"foreignKey:BrandID" json:"types,omitempty"`
	SerialLaptops []SerialLaptop `gorm:"foreignKey:BrandID" json:"serial_laptops,omitempty"`
}

// LaptopType mewakili model LaptopType dengan Composite Unique Index
type LaptopType struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	BrandID       uint           `gorm:"not null;uniqueIndex:idx_brand_type" json:"brand_id"` // Gabung jadi composite index
	Brand         LaptopBrand    `gorm:"foreignKey:BrandID" json:"brand,omitempty"`
	Type          string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_brand_type" json:"type"` // Gabung jadi composite index
	SerialLaptops []SerialLaptop `gorm:"foreignKey:TypeID" json:"serial_laptops,omitempty"`
	
	// Untuk Json di Prisma, GORM menggunakan type string (menyimpan JSON string) atau driver khusus database
	
	
	Images        string         `gorm:"type:json" json:"images"`

	CompatibleSpareparts []CompatibleSparepart `gorm:"foreignKey:TypeID" json:"compatible_spareparts,omitempty"`
}

// 💡 Mockup Struct Pendukung (Agar relasi di atas tidak error saat di-compile)
type SerialLaptop struct {
	ID      uint        `gorm:"primaryKey" json:"id"`
	BrandID uint        `gorm:"not null" json:"brand_id"`
	TypeID  uint        `gorm:"not null" json:"type_id"`
	Serial  string      `gorm:"type:varchar(100);uniqueIndex;not null" json:"serial"`
}

type CompatibleSparepart struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	TypeID      uint   `gorm:"not null" json:"type_id"`
	SparepartName string `gorm:"type:varchar(150);not null" json:"sparepart_name"`
}


type Customer struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	WhatsApp  string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"whatsapp"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}


type Consultation struct {
    ID           uint           `gorm:"primaryKey" json:"id"`
    CustomerID   uint           `json:"customer_id"`
    Customer     Customer       `gorm:"foreignKey:CustomerID" json:"customer"`
    LaptopTypeID *uint          `json:"laptop_type_id"` // nullable jika tipe dimasukkan manual string
    LaptopType   *LaptopType    `gorm:"foreignKey:LaptopTypeID" json:"laptop_type"`

	SymptomID    *uint          `json:"symptom_id"`     // Nullable jika konsultasi umum tanpa lewat simulator harga
	Symptom      *Symptom       `gorm:"foreignKey:SymptomID" json:"symptom"`

	
    Complaint    string         `gorm:"type:text;not null" json:"complaint"`
    Source       string         `gorm:"type:varchar(20);default:'konsultasi';not null" json:"source"` // 'konsultasi' | 'cekharga'
    Status       string         `gorm:"type:varchar(20);default:'PENDING';not null" json:"status"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Category represents master data tingkatan kerusakan
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Type      string         `gorm:"type:varchar(50);not null;unique" json:"type"` // Ringan, Sedang, Berat
	Desc      string         `gorm:"type:text" json:"desc"`
	Symptoms  []Symptom      `gorm:"foreignKey:CategoryID" json:"items,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Symptom represents detail gejala/kendala kerusakan laptop
type Symptom struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CategoryID uint           `gorm:"not null" json:"category_id"`
	Category   Category       `gorm:"foreignKey:CategoryID" json:"category"`
	Name       string         `gorm:"type:varchar(255);not null" json:"name"`
	MinPrice   float64        `gorm:"type:decimal(15,2);not null" json:"min_price"`
	MaxPrice   float64        `gorm:"type:decimal(15,2);not null" json:"max_price"`
	Duration   string         `gorm:"type:varchar(100);default:'1 - 3 Hari'" json:"duration"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type Jasa struct {
	ID                 uint             `gorm:"primaryKey" json:"id"`
	Code               string           `gorm:"type:varchar(50);unique;not null" json:"code"`
	Kategori           string           `gorm:"type:varchar(100)" json:"kategori"` 
	Name               string           `gorm:"type:varchar(255);unique;not null" json:"name"`
	Price              float64          `gorm:"type:decimal(15,2);not null" json:"price"`
	Garansi            *int             `gorm:"type:int" json:"garansi"`             
	Duration           *int             `gorm:"type:int" json:"duration"`            
	Image              string           `gorm:"type:varchar(255)" json:"image"`      
	ItemNoteTemplate   string           `gorm:"type:text" json:"item_note_template"` 
	CreatedFrom        string           `gorm:"type:varchar(100);default:'SYSTEM'" json:"created_from"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	LayananCategoryID  *uint            `json:"layanan_category_id"`
	LayananCategory    *LayananCategory `gorm:"foreignKey:LayananCategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"category,omitempty"`
}

type LayananCategory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Jasas     []Jasa    `gorm:"foreignKey:LayananCategoryID" json:"jasas,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductCategory struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ProductCategories string    `gorm:"type:varchar(100);not null" json:"product_categories"`
	Products          []Product `gorm:"foreignKey:CategoryID" json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Product struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	NamaProduk  string          `gorm:"type:varchar(255);not null" json:"nama_produk"`
	Deskripsi   string          `gorm:"type:text" json:"deskripsi"`
	Images      string          `gorm:"type:varchar(255)" json:"images"` // Menyimpan path file lokal, misal: /uploads/produk-xxx.png
	Harga       int64           `gorm:"type:bigint;not null" json:"harga"`
	Kondisi     string          `gorm:"type:varchar(20);default:'new'" json:"kondisi"`     // new / second
	StatusStok  string          `gorm:"type:varchar(20);default:'ready'" json:"status_stok"` // ready / sold out
	CategoryID  uint            `json:"category_id"`
	Category    *ProductCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type AppSetting struct {
    Key   string `gorm:"primaryKey;type:varchar(100)" json:"key"`
    Value string `gorm:"type:text" json:"value"`
}