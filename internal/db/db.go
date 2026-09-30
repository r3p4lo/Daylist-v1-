// ============================================================================
// PACKAGE: db — Initialization SQLite via GORM + Migration + Seed Master Map
// ----------------------------------------------------------------------------
// CATATAN PENTING untuk Android (LANGKAH 7):
//
//	gorm.io/driver/sqlite memakai cgo (mattn/go-sqlite3) dan TIDAK bisa
//	dikompilasi dengan gomobile ke APK. Solusi standar proyek Fyne mobile:
//	pakai driver murni-Go "github.com/glebarez/sqlite" (GORM dialect, tanpa cgo).
//	API identik — skema & model di atas tidak berubah sedikit pun.
//
// ============================================================================
package db

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite" // driver GORM pure-Go (gomobile-safe)
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mastermap/internal/model"
)

// Open membuka (atau membuat) database di path yang diberikan.
// Di Android, path = <filesDir>/mastermap.db (Fyne: app.Storage().UserStorageRoot()).
func Open(path string) (*gorm.DB, error) {
	gdb, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // senyap di build rilis
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuka sqlite: %w", err)
	}

	// WAL untuk performa write. Catatan: PRAGMA foreign_keys sengaja TIDAK
	// diaktifkan per-koneksi karena database/sql membuka banyak koneksi dan
	// GORM membuat nested-create secara dinamis; integritas relasi berjenjang
	// dijaga lewat: (1) constraint tag OnDelete:CASCADE pada model,
	// (2) transaksi repository, (3) housekeeping CleanupOrphans() berkala.
	sqlDB, err := gdb.DB()
	if err == nil {
		sqlDB.Exec("PRAGMA journal_mode=WAL;")
	}

	// AutoMigrate: Go struct → tabel. Urutan = arah dependensi.
	tables := []any{
		model.Category{},
		model.SubModule{},
		model.Project{},
		model.SubTask{},
		model.DailyPlan{},
		model.TimeBlock{},
		model.FocusSession{},
		model.AppSetting{},
	}
	if err := gdb.AutoMigrate(tables...); err != nil {
		return nil, fmt.Errorf("migrasi gagal: %w", err)
	}

	if err := seed(gdb); err != nil {
		return nil, fmt.Errorf("seed gagal: %w", err)
	}
	return gdb, nil
}

// seed mengisi Master Map (6 kategori + sub-modul awal) HANYA jika kosong.
// Idempotent: dijalankan ulang setiap kali app dibuka tetap aman.
func seed(gdb *gorm.DB) error {
	var count int64
	if err := gdb.Model(&model.Category{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // sudah terisi → jangan sentuh data user
	}

	return gdb.Transaction(func(tx *gorm.DB) error {
		for i := range model.SeedCategories {
			cat := model.SeedCategories[i]
			if err := tx.Create(&cat).Error; err != nil {
				return err
			}
			for j, sm := range model.SeedSubModules[cat.Key] {
				sub := model.SubModule{
					CategoryKey:     cat.Key,
					Name:            sm.Name,
					CognitiveWeight: sm.Weight,
					IsActive:        true,
					SortOrder:       j + 1,
				}
				if err := tx.Create(&sub).Error; err != nil {
					return err
				}
			}
		}
		// Contoh project per BAGIAN 1 (JLPT N3, Automation Engine, dst.)
		seeds := []model.Project{
			{CategoryKey: model.CatLearning, Name: "JLPT N3", Status: model.ProjInProgress, Priority: 1, EstimatedTotalMin: 600},
			{CategoryKey: model.CatTechnology, Name: "Automation Engine", Status: model.ProjInProgress, Priority: 1, EstimatedTotalMin: 480},
			{CategoryKey: model.CatFinance, Name: "Finance Tracker", Status: model.ProjPlanned, Priority: 2, EstimatedTotalMin: 240},
			{CategoryKey: model.CatCareer, Name: "Japan Work Plan", Status: model.ProjInProgress, Priority: 1, EstimatedTotalMin: 360},
			{CategoryKey: model.CatLifestyle, Name: "Morning Routine", Status: model.ProjInProgress, Priority: 2, EstimatedTotalMin: 150},
			{CategoryKey: model.CatPersonalGrowth, Name: "Mindset Journal", Status: model.ProjPlanned, Priority: 3, EstimatedTotalMin: 120},
		}
		for _, p := range seeds {
			p.FolderTemplate = model.DefaultFolderTemplate
			// Rantai hierarki lengkap: project ditautkan ke sub-modul pertamanya.
			var firstSub model.SubModule
			if err := tx.Where("category_key = ? AND is_active = 1", p.CategoryKey).
				Order("sort_order").First(&firstSub).Error; err == nil {
				p.SubModuleID = firstSub.ID
			}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
			// Sub-task modular default: Riset→Desain→Implementasi→Testing→Dokumentasi
			for k, phase := range model.DefaultSubTaskPhases {
				task := model.SubTask{
					ProjectID:  p.ID,
					Name:       phase,
					PhaseOrder: k + 1,
					Status:     model.TaskTodo,
					Complexity: model.WeightMedium,
				}
				if err := tx.Create(&task).Error; err != nil {
					return err
				}
			}
		}

		// Nilai setting default
		settings := []model.AppSetting{
			{Key: model.SettDarkMode, Value: "true", Description: "Dark mode aktif"},
			{Key: model.SettNotifEnabled, Value: "true", Description: "Notifikasi akhir time-block"},
			{Key: model.SettDefaultEnergy, Value: "240", Description: "Budget energi harian (menit)"},
		}
		for _, s := range settings {
			if err := tx.Create(&s).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// TodayDailyPlan mengambil (atau membuat) DailyPlan untuk tanggal hari ini.
func TodayDailyPlan(gdb *gorm.DB) (*model.DailyPlan, error) {
	today := time.Now().Format("2006-01-02") // PlanDate disimpan sebagai string tanggal lokal

	var plan model.DailyPlan
	err := gdb.Where("plan_date = ?", today).First(&plan).Error
	if err == gorm.ErrRecordNotFound {
		plan = model.DailyPlan{PlanDate: today, EnergyBudgetMin: 240}
		if err := gdb.Create(&plan).Error; err != nil {
			return nil, err
		}
		return &plan, nil
	}
	return &plan, err
}
