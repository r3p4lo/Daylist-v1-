// ============================================================================
// PACKAGE: model — Hierarki Dinamis: SubModule → Project → SubTask
// ----------------------------------------------------------------------------
// Aturan modularitas: TIDAK ADA relasi silang antar kategori.
// Semua anak terhubung lewat CategoryKey sehingga query per-modul murah,
// dan menghapus/memodifikasi satu modul tidak menyentuh modul lain.
// ============================================================================
package model

import "time"

// SubModule — divisi kerja di dalam satu kategori (mis. "AI & Automation").
// CognitiveWeight di sini meng-override bobot default kategori untuk
// rekomendasi time-boxing yang lebih presisi (lihat engine, LANGKAH 5).
type SubModule struct {
	ID              uint            `gorm:"primaryKey"`
	CategoryKey     CategoryKey     `gorm:"type:text;index:idx_submod_cat"`
	Name            string          `gorm:"not null"`
	Description     string          // catatan tujuan sub-modul
	CognitiveWeight CognitiveWeight `gorm:"default:2"` // 1=ringan 2=menengah 3=berat
	IsActive        bool            `gorm:"default:true"`
	SortOrder       int             `gorm:"default:0"`

	Projects []Project `gorm:"foreignKey:SubModuleID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SubModule) TableName() string { return "sub_modules" }

// Project — pekerjaan konkret (mis. "Automation Engine", "JLPT N3").
// FolderTemplate menyimpan struktur folder standar project sebagai JSON:
// ["docs/","scripts/","tools/","tests/","README.md"]
type Project struct {
	ID                uint          `gorm:"primaryKey"`
	SubModuleID       uint          `gorm:"index:idx_proj_submod"`
	CategoryKey       CategoryKey   `gorm:"type:text;index:idx_proj_cat"` // denormalisasi: percepat agregasi Master Map
	Name              string        `gorm:"not null"`
	Description       string        // ringkasan & "definition of done"
	Status            ProjectStatus `gorm:"default:'PLANNED';index:idx_proj_status"`
	Priority          int           `gorm:"default:2"` // 1=P1 kritis .. 4=P4 santai
	FolderTemplate    string        `gorm:"default:'[\"docs/\",\"scripts/\",\"tools/\",\"tests/\",\"README.md\"]'"`
	TargetDate        *time.Time    // deadline opsional (nullable)
	EstimatedTotalMin int           `gorm:"default:0"` // estimasi user (menit) → input engine
	ActualTotalSec    int64         `gorm:"default:0"` // akumulasi waktu fokus nyata (kalibrasi)

	SubTasks []SubTask `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`

	// Catatan arsitektur: TIDAK ada relasi GORM Category di sini.
	// CategoryKey adalah FK LOGIS bertipe text; lookup warna/ikon kategori
	// dilakukan lewat repository (JOIN eksplisit), sehingga nested-create
	// GORM tidak pernah mencoba menulisi PK integer tabel categories.

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Project) TableName() string { return "projects" }

// SubTask — unit kerja terkecil pada checklist project.
// Template modular default (contoh Automation Engine):
//  1. Riset & Analisis  2. Desain Arsitektur  3. Implementasi
//  4. Testing & Debugging  5. Dokumentasi
type SubTask struct {
	ID          uint            `gorm:"primaryKey"`
	ProjectID   uint            `gorm:"index:idx_task_proj"`
	Name        string          `gorm:"not null"`
	Description string          // acceptance criteria
	Status      SubTaskStatus   `gorm:"default:'TODO';index:idx_task_status"`
	PhaseOrder  int             `gorm:"default:0"` // urutan fase 1..5
	IsChecked   bool            `gorm:"default:false"`
	Complexity  CognitiveWeight `gorm:"default:2"` // bobot tugas ini sendiri
	EstimateMin int             `gorm:"default:0"` // estimasi menit; 0 = minta rekomendasi engine

	Blocks []TimeBlock `gorm:"->;foreignKey:SubTaskID"` // read-only: blok dibuat engine via repository

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SubTask) TableName() string { return "sub_tasks" }
