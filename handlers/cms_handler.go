package handlers

import (
	"compro/backend-golang/models"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CMSHandler struct {
	DB *gorm.DB
}

// Pastikan struct handler lu punya field DB seperti ini
type ConsultationHandler struct {
	DB *gorm.DB
}

// Struct untuk menampung request data dari Next.js
type ConsultationRequest struct {
	Name       string `json:"name"`
	WhatsApp   string `json:"phone"` // samakan dengan key di FE yaitu 'phone'
	Brand      string `json:"brand"` // Menerima string seperti "Lenovo"
	LaptopType string `json:"laptop_type"`
	Complaint  string `json:"complaint"`
	Symptom    string `json:"symptom"`
	Source     string `json:"source" binding:"required"`      // "cekharga"
}

func NewCMSHandler(db *gorm.DB) *CMSHandler {
	return &CMSHandler{DB: db}
}

// === MENU HANDLERS ===

// func (h *CMSHandler) CreateMenu(c *fiber.Ctx) error {
// 	menu := new(models.Menu)
// 	if err := c.BodyParser(menu); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	if err := h.DB.Create(&menu).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusCreated).JSON(menu)
// }

// === Yang *CMSHandler jalan ya ===
func (h *CMSHandler) CreateMenu(c *fiber.Ctx) error {
    menu := new(models.Menu)
    
    // 1. Parse text fields (Gunakan FormValue karena kita akan menerima Form-Data/Multipart)
    menu.Name = c.FormValue("name")
    menu.Slug = c.FormValue("slug")
    menu.Position = c.FormValue("position")
    
    // Handle manual data opsional & tipe non-string dari FormValue
    if parentIDStr := c.FormValue("parent_id"); parentIDStr != "" {
        var pID uint
        if _, err := fmt.Sscanf(parentIDStr, "%d", &pID); err == nil {
            menu.ParentID = &pID
        }
    }
    if orderStr := c.FormValue("order"); orderStr != "" {
        fmt.Sscanf(orderStr, "%d", &menu.Order)
    }
    menu.IsActive = c.FormValue("is_active") != "false" // default true jika tidak diset false

    // 2. Cek apakah ada file image yang di-upload
    // 2. Cek apakah ada file image yang di-upload
file, err := c.FormFile("image")
if err == nil && file != nil {
    // 💡 Sesuaikan foldernya murni ke ./uploads seperti setup static lu Bos
    if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat folder upload"})
    }

    // Generate nama file unik biar gak saling timpa
    fileExtension := filepath.Ext(file.Filename)
    uniqueFilename := fmt.Sprintf("menu-%d%s", time.Now().UnixNano(), fileExtension)
    
    // 💡 Simpan file langsung ke ./uploads
    filePath := filepath.Join("./uploads", uniqueFilename)

    // Simpan file ke storage server
    if err := c.SaveFile(file, filePath); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan file gambar"})
    }

    // Path di DB tetap "/uploads/..." karena app.Static lu nge-bind "/uploads" ke folder "./uploads"
    menu.ImageUrl = "/uploads/" + uniqueFilename
} else {
    menu.ImageUrl = c.FormValue("image_url")
}

    // 3. Eksekusi simpan ke DB via GORM
    if err := h.DB.Create(&menu).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(fiber.StatusCreated).JSON(menu)
}

// func (h *CMSHandler) GetMenus(c *fiber.Ctx) error {
// 	var allMenus []models.Menu

// 	// 1. Ambil SEMUA data menu navbar tanpa filter parent_id agar ID 5 & 6 ikut terbawa
// 	err := h.DB.Where("position = ?", "navbar").Order("`order` asc").Find(&allMenus).Error
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	var parentMenus []models.Menu

// 	// 2. Looping 1: Ambil yang BENAR-BENAR Parent Utama (ParentID == nil)
// 	for i := 0; i < len(allMenus); i++ {
// 		if allMenus[i].ParentID == nil {
// 			allMenus[i].Submenus = []models.Menu{} // Inisialisasi biar tidak null di JSON
// 			parentMenus = append(parentMenus, allMenus[i])
// 		}
// 	}

// 	// 3. Looping 2: Masukkan Submenu ke Parent dengan tepat memakai index reference
// 	for i := 0; i < len(parentMenus); i++ {
// 		for j := 0; j < len(allMenus); j++ {
// 			// Jika data memiliki parent dan ID parent cocok dengan ID utama saat ini
// 			if allMenus[j].ParentID != nil && *allMenus[j].ParentID == parentMenus[i].ID {
// 				// 💡 KUNCI FIX: Di-append langsung ke indeks parentMenus asli
// 				parentMenus[i].Submenus = append(parentMenus[i].Submenus, allMenus[j])
// 			}
// 		}
// 	}

// 	// 4. Kirim hasil bentukan hirarki pohonnya ke frontend
// 	return c.JSON(parentMenus)
// } IKO LAH JALAN

func (h *CMSHandler) GetMenus(c *fiber.Ctx) error {
	var allMenus []models.Menu

	// 🟢 FIX 1: Ubah "Page" jadi "Pages" sesuai nama field di struct Menu
	err := h.DB.Preload("Pages").Order("`order` asc").Find(&allMenus).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// 🟢 FIX 2: Cek len(m.Pages) > 0 karena m.Pages adalah Slice/Array ([]Page)
	syncMenuSlug := func(m *models.Menu) {
		if len(m.Pages) > 0 && m.Pages[0].Slug != "" {
			m.Slug = m.Pages[0].Slug // 🎯 Ambil slug dari page pertama (botbot)
		}
	}

	var finalResult []models.Menu

	// 1. Parent Utama (navbar & sidebar)
	for i := 0; i < len(allMenus); i++ {
		pos := strings.ToLower(strings.TrimSpace(allMenus[i].Position))
		if allMenus[i].ParentID == nil && (pos == "navbar" || pos == "sidebar") {
			syncMenuSlug(&allMenus[i])
			allMenus[i].Submenus = []models.Menu{}
			finalResult = append(finalResult, allMenus[i])
		}
	}

	// 2. Submenus
	for i := 0; i < len(finalResult); i++ {
		for j := 0; j < len(allMenus); j++ {
			if allMenus[j].ParentID != nil && *allMenus[j].ParentID == finalResult[i].ID {
				syncMenuSlug(&allMenus[j])
				finalResult[i].Submenus = append(finalResult[i].Submenus, allMenus[j])
			}
		}
	}

	// 3. Footer & Body
	for i := 0; i < len(allMenus); i++ {
		pos := strings.ToLower(strings.TrimSpace(allMenus[i].Position))
		if pos == "footer" || pos == "body" {
			syncMenuSlug(&allMenus[i])
			allMenus[i].Submenus = []models.Menu{}
			finalResult = append(finalResult, allMenus[i])
		}
	}

	return c.JSON(finalResult)
}

func (h *CMSHandler) GetNavbarMenus(c *fiber.Ctx) error {
	var allMenus []models.Menu

	// 1. Ambil semua data menu yang aktif untuk navbar sekaligus
	// Gunakan backtick `order` jika Bos menggunakan MySQL agar tidak bentrok dengan SQL keyword ORDER
	err := h.DB.Where("position = ?", "navbar").Order("`order` asc").Find(&allMenus).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// 2. Pisahkan mana yang Menu Utama (Parent) dan mana yang Submenu (Child) via index slicing
	var parentMenus []models.Menu

	// Looping pertama: Ambil semua menu utama (yang parent_id-nya nil/kosong)
	for i := 0; i < len(allMenus); i++ {
		if allMenus[i].ParentID == nil {
			// Inisialisasi array submenus kosong agar di JSON tidak bernilai null (menghindari null di Nuxt)
			allMenus[i].Submenus = []models.Menu{}
			parentMenus = append(parentMenus, allMenus[i])
		}
	}

	// Looping kedua: Masukkan anak submenu ke parent yang cocok berdasarkan cocok ID
	for i := 0; i < len(parentMenus); i++ {
		for j := 0; j < len(allMenus); j++ {
			if allMenus[j].ParentID != nil && *allMenus[j].ParentID == parentMenus[i].ID {
				parentMenus[i].Submenus = append(parentMenus[i].Submenus, allMenus[j])
			}
		}
	}

	// 3. Return hasilnya ke Nuxt Frontend
	return c.JSON(parentMenus)
}



// === DYNAMIC PAGE BUILDER HANDLERS ===

// func (h *CMSHandler) CreatePage(c *fiber.Ctx) error {
// 	page := new(models.Page)
// 	if err := c.BodyParser(page); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	// GORM otomatis menyimpan data bertingkat (Page + Components sekaligus)
// 	if err := h.DB.Create(&page).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusCreated).JSON(page)
// }

func (h *CMSHandler) CreatePage(c *fiber.Ctx) error {
	page := new(models.Page)
	if err := c.BodyParser(page); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// 🟢 SANITASI MENU_ID: Ubah 0 menjadi nil (NULL di MySQL)
	if page.MenuID != nil && *page.MenuID == 0 {
		page.MenuID = nil
	}

	// 🟢 (Optional Safety) Jika MenuID diisi angka selain 0, pastikan ID menu tersebut benar-benar ada di DB
	if page.MenuID != nil {
		var menuExists bool
		h.DB.Model(&models.Menu{}).Select("count(*) > 0").Where("id = ?", *page.MenuID).Find(&menuExists)
		if !menuExists {
			page.MenuID = nil // Reset ke NULL kalau ID menunya gak ketemu di tabel menus
		}
	}

	// GORM otomatis menyimpan data bertingkat (Page + Components sekaligus)
	if err := h.DB.Create(&page).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(page)
}

// 💡 METHOD BARU: Khusus untuk menampilkan daftar di tabel Admin Bos
func (h *CMSHandler) GetAllPages(c *fiber.Ctx) error {
	var pages []models.Page

	// Mengambil semua data page tanpa filter WHERE slug
	err := h.DB.Preload("Components", func(db *gorm.DB) *gorm.DB {
		return db.Order("page_components.order asc")
	}).Find(&pages).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(pages)
}



// func (h *CMSHandler) GetPageBySlug(c *fiber.Ctx) error {
// 	// 💡 Ganti dari c.Params("*") menjadi c.Params("slug") agar sinkron dengan router baru
// 	slug := c.Params("slug")

// 	var page models.Page
	
// 	err := h.DB.Preload("Components", func(db *gorm.DB) *gorm.DB {
//         return db.Order("page_components.order asc")
//     }).
//     Joins("JOIN menus ON menus.id = pages.menu_id").
//     Where("menus.slug = ?", slug).
//     First(&page).Error

//     if err != nil {
//         if err == gorm.ErrRecordNotFound {
//             return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Halaman tidak ditemukan, Bos!"})
//         }
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
//     }


// // 🎯 AMBIL SEMUA MENU BERPOSISI SIDEBAR (CARA B)
//     var sidebarMenus []models.Menu
//     _ = h.DB.Where("position = ? AND is_active = ?", "sidebar", true).
//         Order("`order` asc").
//         Find(&sidebarMenus).Error

//     // 💡 REVISI FIX SAKTI: Gabungkan struct 'page' utuh dengan data sidebar
//     return c.JSON(fiber.Map{
//         "page":          page,          // Ini isinya tetap objek tunggal {...} dari struct page lu beserta Components-nya
//         "sidebar_menus": sidebarMenus,  // Ini array menu sidebarnya
//     })

// 	// Ini akan mengembalikan Objek Tunggal {...} bukan Array [...]
// 	//return c.JSON(page) ini jalan
// }
func (h *CMSHandler) GetPageBySlug(c *fiber.Ctx) error {
	// 💡 Ambil slug dari param router /api/pages/:slug
	slug := c.Params("slug")

	var page models.Page

	// 🟢 FIX UTAMA: Query langsung ke tabel `pages` berdasarkan `pages.slug`
	// Hilangkan JOIN menus agar page dengan menu_id = NULL tetap terbaca sempurna
	err := h.DB.Preload("Components", func(db *gorm.DB) *gorm.DB {
		return db.Order("page_components.order asc")
	}).
		Where("pages.slug = ?", slug).
		First(&page).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Halaman tidak ditemukan, Bos!",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 🎯 AMBIL SEMUA MENU BERPOSISI SIDEBAR
	var sidebarMenus []models.Menu
	_ = h.DB.Where("position = ? AND is_active = ?", "sidebar", true).
		Order("`order` asc").
		Find(&sidebarMenus).Error

	// 💡 Return gabungan data page & sidebar_menus
	return c.JSON(fiber.Map{
		"page":          page,          // Objek tunggal detail page & components
		"sidebar_menus": sidebarMenus,  // Array menu sidebar
	})
}

func (h *CMSHandler) UpdatePage(c *fiber.Ctx) error {
	id := c.Params("id")
	var existingPage models.Page

	// 1. Ambil data asli yang ada di database sekarang
	if err := h.DB.Where("id = ?", id).First(&existingPage).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Page tidak ada"})
	}

	newPageData := new(models.Page)
	if err := c.BodyParser(newPageData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// 🟢 2. SANITASI MENU_ID: Mencegah Foreign Key Constraint Error (1452)
	// Jika MenuID bernilai 0 dari frontend, ubah ke nil (NULL di MySQL)
	if newPageData.MenuID != nil && *newPageData.MenuID == 0 {
		newPageData.MenuID = nil
	}

	// Jika MenuID diisi selain 0, pastikan ID menu tersebut benar-benar ada di tabel menus
	if newPageData.MenuID != nil {
		var menuExists bool
		h.DB.Model(&models.Menu{}).Select("count(*) > 0").Where("id = ?", *newPageData.MenuID).Find(&menuExists)
		if !menuExists {
			newPageData.MenuID = nil // Fallback ke NULL jika menu ID tidak ditemukan
		}
	}

	// 3. Transaksi aman: Hapus komponen lama, ganti dengan susunan block komponen baru dari admin
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Hapus komponen lama bawaan page ini
		if err := tx.Where("page_id = ?", id).Delete(&models.PageComponent{}).Error; err != nil {
			return err
		}

		// 💡 KUNCI SINKRONISASI DATETIME NYA DI SINI, BOS:
		newPageData.ID = existingPage.ID
		newPageData.CreatedAt = existingPage.CreatedAt // Pasang kembali waktu buat aslinya agar tidak jadi 0000-00-00

		// Gunakan Omit("created_at") sebagai proteksi ganda agar GORM mengabaikan kolom ini saat query UPDATE dijalankan
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Omit("created_at").Save(newPageData).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	
	return c.JSON(fiber.Map{"status": "success", "message": "Page updated successfully"})
}

func (h *CMSHandler) DeletePage(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.DB.Delete(&models.Page{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Halaman berhasil dihapus Bos!"})
}

// A. Ambil Semua Data untuk Datatable Admin
// 1. GET ALL FOR DATATABLE
func (h *CMSHandler) GetAnnouncements(c *fiber.Ctx) error {
	var list []models.Announcement
	if err := h.DB.Order("id desc").Find(&list).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil list pengumuman"})
	}
	return c.JSON(list)
}

// 2. GET ACTIVE (Untuk Layout Tampilan Depan Nuxt)
// func (h *CMSHandler) GetActiveAnnouncement(c *fiber.Ctx) error {
// 	var announcement models.Announcement

// 	// Cari yang aktif dan urutan paling baru dimasukkan
// 	err := h.DB.Where("is_active = ?", true).Order("id desc").First(&announcement).Error
// 	if err != nil {
// 		// Jika kosong, return json aman agar front-end tidak crash membaca null
// 		return c.Status(200).JSON(fiber.Map{"is_active": false})
// 	}

//		return c.JSON(announcement)
//	}
func (h *CMSHandler) GetActiveAnnouncement(c *fiber.Ctx) error {
	var announcements []models.Announcement

	// Ambil SEMUA pengumuman yang di-set IsActive = true
	err := h.DB.Where("is_active = ?", true).Order("id desc").Find(&announcements).Error
	if err != nil {
		return c.Status(200).JSON([]interface{}{}) // kembalikan array kosong jika error/tidak ada
	}

	return c.JSON(announcements)
}

// 3. SAVE / CREATE NEW
func (h *CMSHandler) SaveAnnouncement(c *fiber.Ctx) error {
	annType := c.FormValue("type")
	content := c.FormValue("content")
	placement := c.FormValue("placement")
	slugTarget := c.FormValue("slug_target")
	isActive := c.FormValue("is_active") == "true"

	// var imageURL string
	// if annType == "image" {
	// 	file, err := c.FormFile("image")
	// 	if err == nil {
	// 		ext := filepath.Ext(file.Filename)
	// 		uniqueName := fmt.Sprintf("%d-%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	// 		targetPath := fmt.Sprintf("./uploads/%s", uniqueName)
	// 		if err := c.SaveFile(file, targetPath); err != nil {
	// 			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file banner"})
	// 		}
	// 		imageURL = fmt.Sprintf("%s/uploads/%s", c.BaseURL(), uniqueName)
	// 	}
	// }
	var imageURL string
	if annType == "image" {
		file, err := c.FormFile("image")
		if err == nil {
			ext := filepath.Ext(file.Filename)
			uniqueName := fmt.Sprintf("%d-%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)
			targetPath := fmt.Sprintf("./uploads/%s", uniqueName)
			if err := c.SaveFile(file, targetPath); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file banner"})
			}

			// 🔥 LOGIKA DINAMIS: Otomatis mendeteksi port/domain lokal maupun VPS production
			scheme := "http"
			if c.Protocol() == "https" {
				scheme = "https"
			}

			// c.Hostname() akan menghasilkan "localhost:5000" di lokal,
			// dan otomatis menjadi "api.unicomputer.id" saat sudah di-deploy
			imageURL = fmt.Sprintf("%s://%s/uploads/%s", scheme, c.Hostname(), uniqueName)
		}
	}

	// Logic anti dempet posisi
	if isActive {
		h.DB.Model(&models.Announcement{}).
			Where("placement = ? AND is_active = ?", placement, true).
			Update("is_active", false)
	}

	announcement := models.Announcement{
		Type:       annType,
		Content:    content,
		ImageURL:   imageURL,
		Placement:  placement,
		SlugTarget: slugTarget,
		IsActive:   isActive,
	}

	if err := h.DB.Create(&announcement).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan data ke database"})
	}

	return c.JSON(fiber.Map{"success": true, "data": announcement})
}

// 4. TOGGLE STATUS VIA DATATABLE SWITCH
func (h *CMSHandler) ToggleActiveAnnouncement(c *fiber.Ctx) error {
	id := c.Params("id")
	var announcement models.Announcement

	if err := h.DB.First(&announcement, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Data tidak ditemukan"})
	}

	newStatus := !announcement.IsActive

	// Jika diaktifkan, matikan yang lain dengan posisi/placement sejenis
	if newStatus {
		h.DB.Model(&models.Announcement{}).
			Where("placement = ? AND is_active = ?", announcement.Placement, true).
			Update("is_active", false)
	}

	if err := h.DB.Model(&announcement).Update("is_active", newStatus).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengubah status"})
	}

	return c.JSON(fiber.Map{"success": true, "is_active": newStatus})
}

// 5. DELETE
func (h *CMSHandler) DeleteAnnouncement(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.DB.Delete(&models.Announcement{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus data"})
	}
	return c.JSON(fiber.Map{"success": true})
}

// End banner

func (h *CMSHandler) GetWebProfile(c *fiber.Ctx) error {
	var profile models.WebProfile

	// Gunakan Preload untuk menarik data array multi-image slider
	if err := h.DB.Preload("HeaderImages").First(&profile, 1).Error; err != nil {
		profile.ID = 1
		profile.CompanyName = "Uni Computer"
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    profile,
	})
}

func (h *CMSHandler) SaveWebProfile(c *fiber.Ctx) error {
    // 🛠️ 1. Proteksi Folder Uploads
    if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
        errDir := os.MkdirAll("./uploads", 0755)
        if errDir != nil {
            return c.Status(500).JSON(fiber.Map{
                "success": false,
                "error":   "Gagal membuat folder uploads: " + errDir.Error(),
            })
        }
    }

    // 2. Ambil Form Value Teks
    companyName := c.FormValue("company_name")
    email := c.FormValue("email")
    aboutUs := c.FormValue("about_us")
    operationalHours := c.FormValue("operational_hours")
    detailAddress := c.FormValue("detail_address")
    whatsappNumber := c.FormValue("whatsapp_number")

    profile := models.WebProfile{
        ID:               1,
        CompanyName:      companyName,
        Email:            email,
        AboutUs:          aboutUs,
        OperationalHours: operationalHours,
        DetailAddress:    detailAddress,
        WhatsAppNumber:   whatsappNumber,
    }

    // 3. Proses Single File (Logo)
    logoFile, err := c.FormFile("logo")
    if err == nil {
        logoFilename := fmt.Sprintf("logo-%d-%s", time.Now().Unix(), logoFile.Filename)
        if errSave := c.SaveFile(logoFile, fmt.Sprintf("./uploads/%s", logoFilename)); errSave == nil {
            profile.LogoURL = fmt.Sprintf("/uploads/%s", logoFilename)
        }
    } else {
        var oldProfile models.WebProfile
        if errOld := h.DB.Select("logo_url").First(&oldProfile, 1).Error; errOld == nil {
            profile.LogoURL = oldProfile.LogoURL
        }
    }

    // 🔥 4. Proses Peta Multi-Image Slider & Sinkronisasi
    var newSliderImages []models.WebHeaderImage
    var retainedPaths []string

    form, err := c.MultipartForm()
    if err == nil && form != nil {
        // A. Ambil daftar path gambar lama yang dipertahankan oleh Frontend
        // Fiber membaca multiple values dari key yang sama via form.Value
        retainedPaths = form.Value["existing_header_images"]

        // B. Proses file biner baru jika ada yang di-upload
        files := form.File["header_images"]
        for i, file := range files {
            uniqueID := time.Now().UnixMicro()
            sliderFilename := fmt.Sprintf("slider-%d-%d-%s", uniqueID, i, file.Filename)

            if errSave := c.SaveFile(file, fmt.Sprintf("./uploads/%s", sliderFilename)); errSave == nil {
                // Sesuai dengan pengiriman frontend, index link_target sejajar dengan urutan file baru
                linkTarget := c.FormValue(fmt.Sprintf("link_target_%d", i))

                newSliderImages = append(newSliderImages, models.WebHeaderImage{
                    WebProfileID: 1,
                    ImageURL:     fmt.Sprintf("/uploads/%s", sliderFilename),
                    LinkTarget:   linkTarget,
                })
            }
        }
    }

    // 5. Eksekusi Database Transaction
    tx := h.DB.Begin()

    var existingProfile models.WebProfile
    errFind := tx.First(&existingProfile, 1).Error

    if errFind != nil {
        // KONDISI DATA BARU PERTAMA KALI
        profile.ID = 1
        if errSave := tx.Create(&profile).Error; errSave != nil {
            tx.Rollback()
            return c.Status(500).JSON(fiber.Map{
                "success": false,
                "error":   "Gagal membuat data profil baru: " + errSave.Error(),
            })
        }
        
        // Simpan slider baru jika ada
        if len(newSliderImages) > 0 {
            if errSlider := tx.Create(&newSliderImages).Error; errSlider != nil {
                tx.Rollback()
                return c.Status(500).JSON(fiber.Map{
                    "success": false,
                    "error":   "Gagal menyimpan slider baru: " + errSlider.Error(),
                })
            }
        }
    } else {
        // KONDISI EDIT / UPDATE DATA
        existingProfile.CompanyName = companyName
        existingProfile.Email = email
        existingProfile.AboutUs = aboutUs
        existingProfile.OperationalHours = operationalHours
        existingProfile.DetailAddress = detailAddress
        existingProfile.WhatsAppNumber = whatsappNumber

        if logoFile != nil && profile.LogoURL != "" {
            existingProfile.LogoURL = profile.LogoURL
        }

        // 💡 PERBAIKAN UTAMA: Hapus HANYA slider yang tidak ada di daftar retainedPaths
        delQuery := tx.Where("web_profile_id = ?", 1)
        if len(retainedPaths) > 0 {
            // Jika ada yang dipertahankan, hapus yang TIDAK ADA di list tersebut
            delQuery = delQuery.Where("image_url NOT IN ?", retainedPaths)
        }
        
        if errDel := delQuery.Delete(&models.WebHeaderImage{}).Error; errDel != nil {
            tx.Rollback()
            return c.Status(500).JSON(fiber.Map{
                "success": false,
                "error":   "Gagal mensinkronisasi data slider lama: " + errDel.Error(),
            })
        }

        // Simpan profil teks utama
        if errSave := tx.Save(&existingProfile).Error; errSave != nil {
            tx.Rollback()
            return c.Status(500).JSON(fiber.Map{
                "success": false,
                "error":   "Gagal memperbarui data profil: " + errSave.Error(),
            })
        }

        // Tambahkan data slider baru (jika ada) ke dalam database
        if len(newSliderImages) > 0 {
            if errSlider := tx.Create(&newSliderImages).Error; errSlider != nil {
                tx.Rollback()
                return c.Status(500).JSON(fiber.Map{
                    "success": false,
                    "error":   "Gagal menambahkan slider baru: " + errSlider.Error(),
                })
            }
        }

        profile = existingProfile
    }

    tx.Commit()

    // Ambil data paling fresh beserta relasi slidernya untuk dikembalikan ke frontend
    h.DB.Preload("HeaderImages").First(&profile, 1)

    return c.Status(200).JSON(fiber.Map{
        "success": true,
        "message": "Konfigurasi website berhasil diperbarui!",
        "data":     profile,
    })
}

// 🚀 HANDLER POST: Kirim Konsultasi
func (h *CMSHandler) CreateConsultation(c *fiber.Ctx) error {
	req := new(ConsultationRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format request salah"})
	}

	// Cari / Buat Customer berdasarkan nomor telepon
	var customer models.Customer
	err := h.DB.Where("whats_app = ?", req.WhatsApp).First(&customer).Error
	if err == gorm.ErrRecordNotFound {
		customer = models.Customer{Name: req.Name, WhatsApp: req.WhatsApp}
		if err := h.DB.Create(&customer).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal simpan customer"})
		}
	}

	// 🌟 REVISI DISINI BOS: Cari / Buat Brand berdasarkan teks string dari frontend
	var brand models.LaptopBrand
	err = h.DB.Where("name = ?", req.Brand).First(&brand).Error
	if err == gorm.ErrRecordNotFound {
		// Jika brand belum terdaftar di master database, otomatis buat baru
		brand = models.LaptopBrand{Name: req.Brand}
		if err := h.DB.Create(&brand).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat master brand"})
		}
	}
// Cari / Buat tipe laptop di bawah brand tersebut
var laptopType models.LaptopType
err = h.DB.Where("brand_id = ? AND type = ?", brand.ID, req.LaptopType).First(&laptopType).Error
if err == gorm.ErrRecordNotFound {
    
    // 🌟 LANGSUNG ISI DI SINI, BOS! Di dalam fungsi aman dari eror expected declaration
    laptopType = models.LaptopType{
        BrandID: brand.ID,
        Type:    req.LaptopType,
        Images:  "[]", // Ini string literal valid untuk kolom JSON MySQL lu
    }
    
    if err := h.DB.Create(&laptopType).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false, 
            "message": "Gagal simpan master tipe laptop",
        })
    }
}

    laptopTypeIDPtr := &laptopType.ID
	// Simpan transaksi
	consultation := models.Consultation{
		CustomerID:   customer.ID,
		LaptopTypeID: laptopTypeIDPtr,
		Complaint:    req.Complaint,
	}
	if err := h.DB.Create(&consultation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal simpan konsultasi"})
	}

	// Buat link WA redirect
	waText := fmt.Sprintf("Halo Admin Uni Computer...\nNama: %s\nLaptop: %s %s\nKendala: %s", customer.Name, brand.Name, laptopType.Type, consultation.Complaint)
	encodedText := url.QueryEscape(waText)
	adminPhone := "628112626146" 
	waRedirectURL := fmt.Sprintf("https://wa.me/%s?text=%s", adminPhone, encodedText)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":      true,
		"redirect_url": waRedirectURL,
	})
}

// 🚀 HANDLER GET: Tarik Semua Data untuk Halaman Management Admin
func (h *CMSHandler) GetConsultations(c *fiber.Ctx) error {
    var consultations []models.Consultation
    var total int64 = 0

    search := c.Query("search")
    brand := c.Query("brand")
    status := c.Query("status")

    // 1. Buat Base Query yang MURNI dari tabel consultations
    query := h.DB.Model(&models.Consultation{})

    // 2. Filter Search (jika ada)
    if search != "" {
        searchTerm := "%" + search + "%"
        query = query.Where(
            "customer_name LIKE ? OR customer_phone LIKE ? OR laptop_type_name LIKE ?", 
            searchTerm, searchTerm, searchTerm,
        )
    }

    // 3. Filter Status (jika ada)
    if status != "" {
        query = query.Where("status = ?", status)
    }

    // 4. Filter Brand (jika ada)
    if brand != "" {
        query = query.Where("laptop_type_id IN (SELECT id FROM laptop_types WHERE brand_id = ?)", brand)
    }

    // 🚀 HITUNG TOTAL ASLI DULU (Sebelum Preload/Joins agar Count-nya akurat & tidak melipatgandakan data)
    if err := query.Count(&total).Error; err != nil {
        total = 0
    }

    // Jika data memang 0, LANGSUNG RETURN response kosong! Jangan lanjut Preload.
    if total == 0 {
        return c.Status(fiber.StatusOK).JSON(fiber.Map{
            "success": true,
            "data":    []models.Consultation{},
            "total":   0,
        })
    }

    // 5. Eksekusi Ambil Data + Preload Relasi HANYA JIKA ADA DATA
    err := query.
        Preload("Customer").
        Preload("LaptopType").
        Preload("LaptopType.Brand").
        Preload("Symptom").
        Order("created_at desc").
        Find(&consultations).Error

    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Gagal mengambil data antrean",
            "data":    []models.Consultation{},
            "total":   0,
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "success": true,
        "data":    consultations,
        "total":   total,
    })
}

// 1. Handler untuk update status konsultasi
func (h *CMSHandler) UpdateConsultationStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	
	type RequestBody struct {
		Status string `json:"status"`
	}
	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Bad request"})
	}

	// Sesuaikan nama tabel/model database lu (contoh: Consultation)
	if err := h.DB.Table("consultations").Where("id = ?", id).Update("status", body.Status).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui status"})
	}

	return c.JSON(fiber.Map{"message": "Status berhasil diperbarui"})
}

// 2. Handler untuk menghapus data konsultasi
func (h *CMSHandler) DeleteConsultation(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.DB.Table("consultations").Where("id = ?", id).Delete(nil).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus data"})
	}

	return c.JSON(fiber.Map{"message": "Data konsultasi berhasil dihapus"})
}




//cek harga
func (h *CMSHandler) CreatePublicConsultation(c *fiber.Ctx) error {
	var req ConsultationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	// 👥 3. LOGIC FIND OR CREATE CUSTOMER
	var customer struct {
		ID       uint   `gorm:"column:id"`
		Name     string `gorm:"column:name"`
		WhatsApp string `gorm:"column:whats_app"`
	}
	
	err := h.DB.Table("customers").Where("whats_app = ?", req.WhatsApp).Limit(1).Find(&customer).Error
	if err != nil || customer.ID == 0 {
		newCustomer := map[string]interface{}{
			"name":       req.Name,
			"whats_app":  req.WhatsApp,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
		if err := h.DB.Table("customers").Create(&newCustomer).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan data customer baru: " + err.Error()})
		}
		h.DB.Table("customers").Where("whats_app = ?", req.WhatsApp).Limit(1).Find(&customer)
	}

	// 🏷️ 4a. RESOLVE BRAND ID
	var brand struct {
		ID uint `gorm:"column:id"`
	}
	if req.Brand != "" {
		h.DB.Table("brands").Where("name LIKE ?", "%"+req.Brand+"%").Limit(1).Find(&brand)
	}
	if brand.ID == 0 {
		brand.ID = 1 
	}

	// 💻 4b. LOGIC FIND OR CREATE LAPTOP TYPE
	var lType struct {
		ID   uint   `gorm:"column:id"`
		Type string `gorm:"column:type"`
	}
	
	if req.LaptopType == "" {
		req.LaptopType = "Universal / Custom"
	}

	err = h.DB.Table("laptop_types").Where("type = ?", req.LaptopType).Limit(1).Find(&lType).Error
	if err != nil || lType.ID == 0 {
		newType := map[string]interface{}{
			"brand_id": brand.ID,
			"type":     req.LaptopType,
			"images":   "[]",
		}
		if err := h.DB.Table("laptop_types").Create(&newType).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan master tipe laptop baru: " + err.Error()})
		}
		h.DB.Table("laptop_types").Where("type = ?", req.LaptopType).Limit(1).Find(&lType)
	}

	// 🔍 5. SEARCH SYMPTOM ID (MENCARI GEJALA)
	var symptom struct {
		ID uint `gorm:"column:id"`
	}
	if req.Symptom != "" {
		// Mencari ID gejala berdasarkan string yang dikirim (misal: "matot")
		h.DB.Table("symptoms").Where("name LIKE ?", "%"+req.Symptom+"%").Limit(1).Find(&symptom)
	}

	// Proteksi akhir memastikan ID relasi valid sebelum ke consultations
	if customer.ID == 0 || lType.ID == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Relasi data Customer atau Tipe Laptop gagal dibuat dengan benar"})
	}

	// 📝 6. INSERT DATA KONSULTASI / CEK HARGA
	consultation := map[string]interface{}{
		"customer_id":    customer.ID,
		"laptop_type_id": lType.ID,
		"complaint":      req.Complaint,
		"source":         req.Source,
		"status":         "PENDING",
		"created_at":     time.Now(),
		"updated_at":     time.Now(),
	}

	// 🌟 ISI SYMPTOM ID JIKA DITEMUKAN
	if symptom.ID != 0 {
		consultation["symptom_id"] = symptom.ID
	} else {
		consultation["symptom_id"] = nil // Tetap aman jika gejala tidak match di master data
	}

	if err := h.DB.Table("consultations").Create(&consultation).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan data estimasi biaya: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data cek harga berhasil masuk database dengan Symptom ID, Bos!",
	})
}

// ==================== MASTER BRAND HANDLERS ====================

func (h *CMSHandler) GetBrands(c *fiber.Ctx) error {
	var brands []models.LaptopBrand
	if err := h.DB.Order("id desc").Find(&brands).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data brand"})
	}
	return c.JSON(brands)
}

func (h *CMSHandler) CreateBrand(c *fiber.Ctx) error {
	// 1. Ambil input nama brand dari Form Value (karena kita kirim pakai FormData)
	brandName := strings.TrimSpace(c.FormValue("name"))
	if brandName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nama brand tidak boleh kosong",
		})
	}

	var imagePath string

	// 2. Tangkap file image yang dikirim oleh Nuxt/Vue tadi (Key: "image")
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		// Buat nama file unik agar tidak bentrok (contoh: brand-asus-171829102.png)
		cleanBrandName := strings.ToLower(strings.ReplaceAll(brandName, " ", "-"))
		extension := filepath.Ext(file.Filename)
		newFilename := fmt.Sprintf("brand-%s-%d%s", cleanBrandName, time.Now().Unix(), extension)

		// Simpan file fisik ke folder ./uploads lokal backend Anda
		errSave := c.SaveFile(file, fmt.Sprintf("./uploads/%s", newFilename))
		if errSave == nil {
			// Simpan path relatif ke database agar bisa langsung dipanggil
			imagePath = fmt.Sprintf("/uploads/%s", newFilename)
		} else {
			// Jika gagal simpan file, log error-nya
			fmt.Println("Gagal menyimpan file gambar brand:", errSave)
		}
	}

	// 3. Masukkan datanya ke Database GORM
	newBrand := models.LaptopBrand{ // Sesuaikan nama struct Model Brand Anda
		Name:  brandName,
		Image: imagePath, // Menyimpan path ke kolom `image` (misal: /uploads/brand-lenovo-12345.png)
	}

	if err := h.DB.Create(&newBrand).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan brand ke database: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Brand berhasil ditambahkan!",
		"data":    newBrand,
	})
}

func (h *CMSHandler) UpdateBrand(c *fiber.Ctx) error {
	id := c.Params("id")
	var brand models.LaptopBrand
	if err := c.BodyParser(&brand); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Payload tidak valid"})
	}
	if err := h.DB.Table("brands").Where("id = ?", id).Updates(&brand).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui brand"})
	}
	return c.JSON(fiber.Map{"message": "Brand berhasil diperbarui"})
}

func (h *CMSHandler) DeleteBrand(c *fiber.Ctx) error {
	id := c.Params("id")
	
	if err := h.DB.Where("id = ?", id).Delete(&models.LaptopBrand{}).Error; err != nil {
		// Jika error disebabkan oleh relasi foreign key constraint di MySQL
		if strings.Contains(err.Error(), "1451") {
			return c.Status(400).JSON(fiber.Map{
				"message": "Gagal menghapus! Brand ini masih digunakan oleh beberapa tipe laptop. Hapus dulu tipe laptop terkait.",
			})
		}
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus brand"})
	}
	
	return c.JSON(fiber.Map{"message": "Brand berhasil dihapus"})
}

// ==================== MASTER CATEGORY HANDLERS ====================

func (h *CMSHandler) GetCategories(c *fiber.Ctx) error {
	var categories []models.Category
	if err := h.DB.Order("id asc").Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data kategori"})
	}
	return c.JSON(categories)
}

// ==================== MASTER SYMPTOM HANDLERS ====================

// func (h *CMSHandler) GetSymptoms(c *fiber.Ctx) error {
// 	var symptoms []models.Symptom
// 	if err := h.DB.Preload("Category").Order("id desc").Find(&symptoms).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data gejala"})
// 	}
// 	return c.JSON(symptoms)
// }
func (h *CMSHandler) GetSymptoms(c *fiber.Ctx) error {
    var symptoms []models.Symptom

    // 🚀 KUNCINYA DI SINI: Tambahkan .Preload("Category")
    err := h.DB.Preload("Category").
        Order("id asc").
        Find(&symptoms).Error

    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": "Gagal mengambil data gejala",
        })
    }

    return c.Status(fiber.StatusOK).JSON(symptoms)
}

func (h *CMSHandler) CreateSymptom(c *fiber.Ctx) error {
	var symptom models.Symptom
	if err := c.BodyParser(&symptom); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Payload tidak valid"})
	}
	if err := h.DB.Create(&symptom).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan gejala"})
	}
	return c.Status(201).JSON(symptom)
}

func (h *CMSHandler) DeleteSymptom(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.DB.Table("symptoms").Where("id = ?", id).Delete(nil).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus gejala"})
	}
	return c.JSON(fiber.Map{"message": "Gejala berhasil dihapus"})
}


// Blok Jasa
// GET /admin/jasas


func (h *CMSHandler) GetJasas(c *fiber.Ctx) error {
    var jasas []models.Jasa
    
    // WAJIB Preload("LayananCategory") agar GORM melakukan JOIN otomatis ke tabel kategori
    if err := h.DB.Preload("LayananCategory").Find(&jasas).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data jasa"})
    }
    
    return c.JSON(jasas)
}

// GET /admin/jasa-categories
func (h *CMSHandler) GetLayananCategories(c *fiber.Ctx) error {
    var categories []models.LayananCategory
    
    // Gunakan .Preload("Jasas") jika di model LayananCategory Bos sudah mendefinisikan relasi HasMany ke Jasa
    if err := h.DB.Order("id desc").Preload("Jasas").Find(&categories).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil kategori layanan"})
    }
    
    return c.JSON(categories)
}

// POST /admin/jasas
func (h *CMSHandler) CreateJasa(c *fiber.Ctx) error {
	// 1. Tangkap semua input teks dari FormValue
	code := c.FormValue("code")
	kategori := c.FormValue("kategori")
	name := c.FormValue("name")
	priceStr := c.FormValue("price")
	garansiStr := c.FormValue("garansi")
	durationStr := c.FormValue("duration")
	itemNoteTemplate := c.FormValue("item_note_template")
	layananCategoryIDStr := c.FormValue("layanan_category_id")

	// Validasi dasar
	if code == "" || name == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Code dan Name tidak boleh kosong"})
	}

	// 2. Parsing tipe data khusus (Price float64)
	var price float64
	if priceStr != "" {
		if p, err := strconv.ParseFloat(priceStr, 64); err == nil {
			price = p
		}
	}

	// Parsing pointer integer (Garansi)
	var garansi *int
	if garansiStr != "" {
		if g, err := strconv.Atoi(garansiStr); err == nil {
			garansi = &g
		}
	}

	// Parsing pointer integer (Duration)
	var duration *int
	if durationStr != "" {
		if d, err := strconv.Atoi(durationStr); err == nil {
			duration = &d
		}
	}

	// Parsing pointer uint (LayananCategoryID)
	var layananCategoryID *uint
	if layananCategoryIDStr != "" {
		if id, err := strconv.ParseUint(layananCategoryIDStr, 10, 32); err == nil {
			uid := uint(id)
			layananCategoryID = &uid
		}
	}

	var imagePath string

	// 3. Proses upload file gambar
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		cleanName := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		extension := filepath.Ext(file.Filename)
		newFilename := fmt.Sprintf("jasa-%s-%d%s", cleanName, time.Now().Unix(), extension)

		// Simpan file ke folder ./uploads lokal
		errSave := c.SaveFile(file, fmt.Sprintf("./uploads/%s", newFilename))
		if errSave == nil {
			imagePath = fmt.Sprintf("/uploads/%s", newFilename)
		} else {
			fmt.Println("Gagal menyimpan gambar jasa:", errSave)
		}
	}

	// 4. Mapping data ke struct Jasa
	jasa := models.Jasa{
		Code:              code,
		Kategori:          kategori,
		Name:              name,
		Price:             price,
		Garansi:           garansi,
		Duration:          duration,
		ItemNoteTemplate:  itemNoteTemplate,
		CreatedFrom:       "SYSTEM",
		LayananCategoryID: layananCategoryID,
		Image:             imagePath, // Jika kosong, di database akan bernilai string kosong ""
	}

	// 5. Simpan ke Database
	if err := h.DB.Create(&jasa).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan data jasa: " + err.Error()})
	}

	return c.Status(201).JSON(jasa)
}

// PUT /admin/jasas/:id
// PUT /admin/jasas/:id
func (h *CMSHandler) UpdateJasa(c *fiber.Ctx) error {
	id := c.Params("id")
	var jasa models.Jasa

	// 1. Cari data jasa lama di database
	if err := h.DB.First(&jasa, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Data jasa tidak ditemukan"})
	}

	// 2. Tangkap input teks dari FormValue (karena dikirim via FormData)
	code := c.FormValue("code")
	kategori := c.FormValue("kategori")
	name := c.FormValue("name")
	priceStr := c.FormValue("price")
	garansiStr := c.FormValue("garansi")
	durationStr := c.FormValue("duration")
	itemNoteTemplate := c.FormValue("item_note_template")
	layananCategoryIDStr := c.FormValue("layanan_category_id")

	// Update field teks jika dikirimkan
	if code != "" {
		jasa.Code = code
	}
	if name != "" {
		jasa.Name = name
	}
	jasa.Kategori = kategori
	jasa.ItemNoteTemplate = itemNoteTemplate

	// Parsing Price (float64)
	if priceStr != "" {
		if p, err := strconv.ParseFloat(priceStr, 64); err == nil {
			jasa.Price = p
		}
	}

	// Parsing Garansi (Pointer int)
	if garansiStr != "" {
		if g, err := strconv.Atoi(garansiStr); err == nil {
			jasa.Garansi = &g
		}
	} else if c.FormValue("garansi") == "" && garansiStr == "" {
		jasa.Garansi = nil
	}

	// Parsing Duration (Pointer int)
	if durationStr != "" {
		if d, err := strconv.Atoi(durationStr); err == nil {
			jasa.Duration = &d
		}
	} else if c.FormValue("duration") == "" && durationStr == "" {
		jasa.Duration = nil
	}

	// Parsing LayananCategoryID (Pointer uint)
	if layananCategoryIDStr != "" {
		if lid, err := strconv.ParseUint(layananCategoryIDStr, 10, 32); err == nil {
			uid := uint(lid)
			jasa.LayananCategoryID = &uid
		}
	} else if c.FormValue("layanan_category_id") == "" && layananCategoryIDStr == "" {
		jasa.LayananCategoryID = nil
	}

	// 3. Proses upload file gambar BARU (jika ada)
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		cleanName := strings.ToLower(strings.ReplaceAll(jasa.Name, " ", "-"))
		extension := filepath.Ext(file.Filename)
		newFilename := fmt.Sprintf("jasa-%s-%d%s", cleanName, time.Now().Unix(), extension)

		// Simpan file fisik baru ke folder ./uploads
		errSave := c.SaveFile(file, fmt.Sprintf("./uploads/%s", newFilename))
		if errSave == nil {
			// Opsional: Hapus file gambar lama dari server jika Bos ingin menghemat storage VPS
			// if jasa.Image != "" {
			//     os.Remove(fmt.Sprintf(".%s", jasa.Image))
			// }

			// Set path gambar baru
			jasa.Image = fmt.Sprintf("/uploads/%s", newFilename)
		} else {
			fmt.Println("Gagal menyimpan gambar baru jasa:", errSave)
		}
	}
	// Catatan: Jika 'file' nil, kita biarkan jasa.Image tetap menggunakan nilai lamanya.

	// 4. Simpan perubahan ke database menggunakan Save
	if err := h.DB.Save(&jasa).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui data jasa"})
	}

	return c.JSON(jasa)
}


// DELETE /admin/jasas/:id
func (h *CMSHandler) DeleteJasa(c *fiber.Ctx) error {
	id := c.Params("id")
	var jasa models.Jasa

	// 1. Cari dulu datanya untuk memastikan data itu ada
	if err := h.DB.First(&jasa, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Data jasa tidak ditemukan"})
	}

	// Opsional: Hapus file gambar fisik di server agar storage VPS unicomputer.id tidak bengkak
	// if jasa.Image != "" {
	//     os.Remove(fmt.Sprintf(".%s", jasa.Image))
	// }

	// 2. Hapus datanya (GORM otomatis mendeteksi Soft Delete jika memakai gorm.Model / deleted_at)
	if err := h.DB.Delete(&jasa).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus data jasa"})
	}

	return c.JSON(fiber.Map{"message": "Jasa berhasil dihapus"})
}

// === HANDLERS UNTUK CATEGORIES ===

func (h *CMSHandler) GetProductCategories(c *fiber.Ctx) error {
	var categories []models.ProductCategory
	if err := h.DB.Order("id desc").Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": categories})
}

func (h *CMSHandler) CreateCategory(c *fiber.Ctx) error {
	var cat models.ProductCategory
	if err := c.BodyParser(&cat); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid input"})
	}
	if err := h.DB.Create(&cat).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": cat})
}

// === HANDLERS UNTUK PRODUCTS ===

func (h *CMSHandler) GetProducts(c *fiber.Ctx) error {
	var products []models.Product
	
	// Preload relasi kategori biar namanya kebaca di frontend
	dbQuery := h.DB.Preload("Category")
	
	// Filter berdasarkan kategori jika ada param query ?category_id=X
	if catID := c.Query("category_id"); catID != "" {
		dbQuery = dbQuery.Where("category_id = ?", catID)
	}

	if err := dbQuery.Order("id desc").Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": products})
}

func (h *CMSHandler) CreateProduct(c *fiber.Ctx) error {
	// Parse multipart form data untuk file upload
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Failed to parse form data"})
	}

	namaProduk := form.Value["nama_produk"][0]
	deskripsi := form.Value["deskripsi"][0]
	kondisi := form.Value["kondisi"][0]
	statusStok := form.Value["status_stok"][0]
	harga, _ := strconv.ParseInt(form.Value["harga"][0], 10, 64)
	catID, _ := strconv.ParseUint(form.Value["category_id"][0], 10, 64)

	var imagePath string

	// Handle File Image Upload jika ada file yang di-choose
	files := form.File["image"]
	if len(files) > 0 {
		file := files[0]
		// Bikin folder uploads jika belum ada
		if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to create upload dir"})
		}

		// Generate nama unik agar file tidak bertabrakan
		ext := filepath.Ext(file.Filename)
		uniqueName := fmt.Sprintf("produk-%d%s", time.Now().UnixNano(), ext)
		targetPath := filepath.Join("./uploads", uniqueName)

		// Simpan file secara lokal
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to open temporary file"})
		}
		defer src.Close()

		dst, err := os.Create(targetPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to create local file"})
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to save file"})
		}

		imagePath = "/uploads/" + uniqueName
	}

	product := models.Product{
		NamaProduk: namaProduk,
		Deskripsi:  deskripsi,
		Harga:      harga,
		Kondisi:    kondisi,
		StatusStok: statusStok,
		CategoryID: uint(catID),
		Images:     imagePath,
	}

	if err := h.DB.Create(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": product})
}

func (h *CMSHandler) UpdateProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Product not found"})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Failed to parse form data"})
	}

	if len(form.Value["nama_produk"]) > 0 { product.NamaProduk = form.Value["nama_produk"][0] }
	if len(form.Value["deskripsi"]) > 0 { product.Deskripsi = form.Value["deskripsi"][0] }
	if len(form.Value["kondisi"]) > 0 { product.Kondisi = form.Value["kondisi"][0] }
	if len(form.Value["status_stok"]) > 0 { product.StatusStok = form.Value["status_stok"][0] }
	if len(form.Value["harga"]) > 0 { product.Harga, _ = strconv.ParseInt(form.Value["harga"][0], 10, 64) }
	if len(form.Value["category_id"]) > 0 {
		catID, _ := strconv.ParseUint(form.Value["category_id"][0], 10, 64)
		product.CategoryID = uint(catID)
	}

	// Cek jika ada upload image baru pengganti image lama
	files := form.File["image"]
	if len(files) > 0 {
		file := files[0]
		ext := filepath.Ext(file.Filename)
		uniqueName := fmt.Sprintf("produk-%d%s", time.Now().UnixNano(), ext)
		targetPath := filepath.Join("./uploads", uniqueName)

		src, err := file.Open()
		if err == nil {
			defer src.Close()
			dst, err := os.Create(targetPath)
			if err == nil {
				defer dst.Close()
				if _, err = io.Copy(dst, src); err == nil {
					// Hapus image lama di local disk jika ada agar tidak menumpuk sampah
					if product.Images != "" {
						oldFile := "." + product.Images
						_ = os.Remove(oldFile)
					}
					product.Images = "/uploads/" + uniqueName
				}
			}
		}
	}

	if err := h.DB.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": product})
}

func (h *CMSHandler) DeleteProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Product not found"})
	}

	// Hapus file fisiknya dari folder uploads
	if product.Images != "" {
		_ = os.Remove("." + product.Images)
	}

	if err := h.DB.Delete(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Product successfully deleted"})
}


//Switch ON / OFF SECTION
func (h *CMSHandler) GetProductSectionStatus(c *fiber.Ctx) error {
    var setting models.AppSetting
    
    // Tambahkan backtick pada `key` di dalam query Where
    err := h.DB.Table("app_settings").Where("`key` = ?", "show_product_section").First(&setting).Error
    
    if err != nil {
        // Jika belum ada datanya di DB, default ke true
        return c.JSON(fiber.Map{"success": true, "data": true})
    }
    
    return c.JSON(fiber.Map{
        "success": true, 
        "data":    setting.Value == "true",
    })
}

func (h *CMSHandler) UpdateProductSectionStatus(c *fiber.Ctx) error {
    type Request struct {
        Status bool `json:"status"`
    }
    
    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid request"})
    }

    valStr := "false"
    if req.Status {
        valStr = "true"
    }

    // Gunakan models.AppSetting di sini juga saat proses Upsert/Save
    err := h.DB.Table("app_settings").Save(&models.AppSetting{Key: "show_product_section", Value: valStr}).Error
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"success": false, "message": "Gagal update database"})
    }

    return c.JSON(fiber.Map{"success": true, "message": "Status berhasil diperbarui"})
}




