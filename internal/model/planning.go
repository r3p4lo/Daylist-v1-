// ============================================================================
// PACKAGE: model — Perencanaan Harian: DailyPlan, TimeBlock, FocusSession
// ----------------------------------------------------------------------------
// TimeBlock adalah eksekusi nyata dari hasil Dynamic Time-Boxing Engine.
// Fokus sesi memakai TIMESTAMP (bukan hitungan tick) agar akurat walau
// aplikasi di-minimize / proses dihentikan Android sementara.
// ============================================================================
package model

import "time"

// DailyPlan — kerangka jadwal satu hari. Ringan: hanya menampung meta.
type DailyPlan struct {
	ID              uint   `gorm:"primaryKey"`
	PlanDate        string `gorm:"type:text;uniqueIndex:uq_plan_date"` // format "2006-01-02" (SQLite tidak punya tipe DATE murni)
	Theme           string // niat utama hari ini, mis. "Fokus Automation Engine"
	EnergyBudgetMin int    `gorm:"default:240"` // pagi energi harian (menit) — anti "satu kategori menyedot semua"
	Notes           string

	Blocks []TimeBlock `gorm:"foreignKey:DailyPlanID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DailyPlan) TableName() string { return "daily_plans" }

// Tanggal — parse kembali string PlanDate menjadi time.Time (mulai tengah malam).
func (d DailyPlan) Tanggal() time.Time {
	t, _ := time.ParseInLocation("2006-01-02", d.PlanDate, time.Local)
	return t
}

// TimeBlock — satu blok waktu (Blok A/B atau istirahat) di timeline harian.
// Fields Min/Max/Durasi dalam MENIT; kolom *_at untuk eksekusi nyata.
type TimeBlock struct {
	ID          uint        `gorm:"primaryKey"`
	DailyPlanID uint        `gorm:"index:idx_blk_plan"`
	SubTaskID   uint        `gorm:"index:idx_blk_task;default:0"` // 0 = blok bebas/break
	CategoryKey CategoryKey `gorm:"type:text;index:idx_blk_cat"`  // color-coding timeline
	BlockType   BlockType   `gorm:"default:'A'"`                  // A / B / "BREAK"
	Title       string      // label tampil di timeline
	DurationMin int         `gorm:"not null"`         // durasi yang direncanakan
	SequenceNo  int         `gorm:"default:0"`        // urutan dalam rantai blok satu tugas
	IsBreak     bool        `gorm:"default:false"`    // true = jeda istirahat
	Source      string      `gorm:"default:'ENGINE'"` // ENGINE | USER_OVERRIDE
	Status      BlockStatus `gorm:"default:'PLANNED'"`

	PlannedStart *time.Time // anchor posisi di timeline hari itu
	StartedAt    *time.Time // timestamp nyata saat Start (Focus Mode)
	EndedAt      *time.Time // timestamp nyata saat selesai/pause akhir

	// Sinyal kalibrasi engine (LANGKAH 5):
	OverrunMin int `gorm:"default:0"` // +melebihi / -selesai lebih cepat

	// Asosiasi read-only (preloading UI saja) — penulisan selalu via repository
	// dengan ID eksplisit, mencegah GORM nested-create menabrak constraint FK.
	FocusSessions []FocusSession `gorm:"->;foreignKey:TimeBlockID"`
	SubTask       *SubTask       `gorm:"->"`
	DailyPlan     *DailyPlan     `gorm:"->"`
	Category      *Category      `gorm:"->"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TimeBlock) TableName() string { return "time_blocks" }

// FocusSession — rekaman segment timer (start→pause/resume→stop).
// Total fokus sebuah blok = SUM(EndedAt - StartedAt) seluruh sesinya.
// Ini kunci akurasi timer lintas minimize: sisa waktu dihitung dari
// now - StartedAt - PausedAccumSec, bukan dari akumulasi tick UI.
type FocusSession struct {
	ID             uint       `gorm:"primaryKey"`
	TimeBlockID    uint       `gorm:"index:idx_sess_blk"`
	StartedAt      time.Time  `gorm:"not null"`
	EndedAt        *time.Time // nil = masih berjalan
	PausedAccumSec int        `gorm:"default:0"`     // total detik saat status paused
	Interrupted    bool       `gorm:"default:false"` // user berhenti sebelum tuntas
	Note           string     // jurnal singkat pasca-sesi
}

func (FocusSession) TableName() string { return "focus_sessions" }

// -----------------------------------------------------------------------------
// AppSetting — preferensi UI/global (Dark Mode, notifikasi, dsb — LANGKAH 6/7).
// -----------------------------------------------------------------------------
type AppSetting struct {
	Key         string `gorm:"primaryKey;type:text"`
	Value       string `gorm:"not null"`
	Description string

	UpdatedAt time.Time
}

func (AppSetting) TableName() string { return "app_settings" }

// Kunci setting yang dikenal aplikasi:
const (
	SettDarkMode      = "ui.dark_mode"    // "true"/"false"
	SettNotifEnabled  = "notif.enabled"   // pengingat akhir blok
	SettDefaultEnergy = "plan.energy_min" // budget energi harian default
	SettLastBackupAt  = "backup.last_at"  // timestamp RFC3339
)
