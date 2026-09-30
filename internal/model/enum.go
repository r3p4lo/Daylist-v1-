// ============================================================================
// PACKAGE: model — Definisi Enum & Konstanta Domain "Master Map Life Planner"
// ----------------------------------------------------------------------------
// Prinsip Modular + Satu Peta Induk: setiap entitas punya "rumah" (category_key)
// sehingga modul yang satu TIDAK mengganggu modul yang lain.
// ============================================================================
package model

// -----------------------------------------------------------------------------
// CATEGORY KEY — 6 kategori tetap pada Master Map (Peta Induk).
// Key bersifat stabil (dipakai sebagai foreign key logis di seluruh aplikasi,
// termasuk untuk color-coding UI dan bobot kognitif engine time-boxing).
// -----------------------------------------------------------------------------
type CategoryKey string

const (
	CatLearning       CategoryKey = "LEARNING"        // 🎓 Hijau   #2E7D32
	CatTechnology     CategoryKey = "TECHNOLOGY"      // ⚙️ Biru    #1565C0
	CatFinance        CategoryKey = "FINANCE"         // 💰 Oranye  #E65100
	CatCareer         CategoryKey = "CAREER"          // 💼 Ungu    #6A1B9A
	CatLifestyle      CategoryKey = "LIFESTYLE"       // ❤️ Crimson #C62828
	CatPersonalGrowth CategoryKey = "PERSONAL_GROWTH" // 🌱 Teal    #00695C
)

// AllCategories — urutan tampilan pada Master Map View (LANGKAH 3).
var AllCategories = []CategoryKey{
	CatLearning, CatTechnology, CatFinance,
	CatCareer, CatLifestyle, CatPersonalGrowth,
}

// -----------------------------------------------------------------------------
// COGNITIVE WEIGHT — "Bobot Kognitif" kategori/sub-modul.
// Ini adalah JANTUNG dari sistem Dynamic Time-Boxing (LANGKAH 5):
// engine memakai nilai ini untuk merekomendasikan Blok A atau Blok B.
// -----------------------------------------------------------------------------
type CognitiveWeight int

const (
	WeightLight  CognitiveWeight = 1 // Administratif, budgeting, email, olahraga, meditasi
	WeightMedium CognitiveWeight = 2 // Belajar bahasa, kursus, refleksi, desain standar
	WeightHeavy  CognitiveWeight = 3 // Riset berat, AI & Automation, Web3, arsitektur
)

// -----------------------------------------------------------------------------
// TIME BLOCK TYPE — Preset dasar time-boxing.
//
//	Blok A (Tugas Ringan/Administratif) : 30 menit (rentang dinamis 15-30)
//	Blok B (Tugas Berat/Deep Work)      : 50 menit (rentang dinamis 45-90)
//
// -----------------------------------------------------------------------------
type BlockType string

const (
	BlockA BlockType = "A" // ringan
	BlockB BlockType = "B" // deep work
)

// Durasi preset (menit) — nilai AWAL, bukan angka tetap selamanya.
// Engine boleh memilih durasi di dalam RentangMin..RentangMax sesuai kompleksitas.
const (
	PresetBlockADuration = 30
	PresetBlockBDuration = 50
	BreakShortMinutes    = 5  // jeda antar blok ringan
	BreakLongMinutes     = 10 // jeda antar blok berat
)

// -----------------------------------------------------------------------------
// STATUS & STATE enums — dipakai untuk progress bar (Master Map) & checklist.
// -----------------------------------------------------------------------------
type ProjectStatus string

const (
	ProjPlanned    ProjectStatus = "PLANNED"
	ProjInProgress ProjectStatus = "IN_PROGRESS"
	ProjOnHold     ProjectStatus = "ON_HOLD"
	ProjDone       ProjectStatus = "DONE"
	ProjArchived   ProjectStatus = "ARCHIVED"
)

type SubTaskStatus string

const (
	TaskTodo    SubTaskStatus = "TODO"
	TaskDoing   SubTaskStatus = "DOING"
	TaskBlocked SubTaskStatus = "BLOCKED"
	TaskReview  SubTaskStatus = "REVIEW"
	TaskDone    SubTaskStatus = "DONE"
)

type BlockStatus string

const (
	BlkPlanned BlockStatus = "PLANNED"
	BlkActive  BlockStatus = "ACTIVE" // sedang dijalankan di Focus Mode
	BlkDone    BlockStatus = "DONE"
	BlkSkipped BlockStatus = "SKIPPED"
	BlkOverrun BlockStatus = "OVERRUN" // selesai melewati durasi → sinyal kalibrasi
)

// -----------------------------------------------------------------------------
// TIMER MODE — state machine Active Timer (LANGKAH 6).
// -----------------------------------------------------------------------------
type TimerMode string

const (
	TimerIdle     TimerMode = "IDLE"
	TimerRunning  TimerMode = "RUNNING"
	TimerPaused   TimerMode = "PAUSED"
	TimerBreak    TimerMode = "BREAK"
	TimerFinished TimerMode = "FINISHED"
)

// -----------------------------------------------------------------------------
// FILE/FOLDER TEMPLATE — struktur folder standar project modular, contoh:
// Technology → Automation Engine: docs/, scripts/, tools/, tests/, README.md
// Disimpan sebagai JSON array di kolom Project.FolderTemplate.
// -----------------------------------------------------------------------------
const DefaultFolderTemplate = `["docs/","scripts/","tools/","tests/","README.md"]`

// SubTask template modular untuk project baru (LANGKAH 5 - contoh alur kerja):
var DefaultSubTaskPhases = []string{
	"Riset & Analisis",
	"Desain Arsitektur",
	"Implementasi",
	"Testing & Debugging",
	"Dokumentasi",
}
