// ============================================================================
// PACKAGE: state — State Management Terpusat (ringan & modular)
// ----------------------------------------------------------------------------
// Prinsip arsitektur (lihat LANGKAH 1 §1.5):
//  1. SATU sumber kebenaran: struct App dipegang main.go, di-inject ke semua
//     layar. Tidak ada singleton tersembunyi → tiap modul UI terisolasi.
//  2. DarkMode adalah *bool — theme engine membaca pointer yang sama, sehingga
//     toggle mode terang/gelap berlaku instan di seluruh aplikasi tanpa restart.
//  3. Timer TIDAK menyimpan "sisa detik" di memori: cukup StartedAt &
//     PausedAccumSec di DB; nilai live dihitung dari selisih timestamp, jadi
//     tetap akurat walau app di-minimize / proses Android dihentikan.
//  4. Navigasi lewat stack layar (Navigator) — lihat ui/navigator.go.
//
// ============================================================================
package state

import (
	"sync"
	"time"

	"gorm.io/gorm"

	"mastermap/internal/model"
)

// App — agregat state global aplikasi.
type App struct {
	DB       *gorm.DB // koneksi SQLite (GORM) — dibuka sekali di main.go
	DarkMode *bool    // pointer bersama dengan AppTheme (live switch)

	NavMu sync.Mutex // guard navigasi agar transisi tidak bertumpuk

	tickMu sync.Mutex // guard callback ticker 1 Hz
	onTick func()     // dipasang oleh layar Focus Mode (lihat SetTick)

	// ActiveTimer — status Focus Mode saat ini (diisi dari DB saat resume).
	ActiveTimer TimerState
}

// TimerState — proyeksi satu TimeBlock yang sedang berjalan.
// Fields waktu memakai basis TIMESTAMP (akurat lintas minimize).
type TimerState struct {
	BlockID        uint
	CategoryKey    model.CategoryKey
	Title          string
	Mode           model.TimerMode // IDLE/RUNNING/PAUSED/BREAK/FINISHED
	StartedAt      time.Time       // anchor sesi (dari FocusSession.StartedAt)
	PausedAccumSec int             // total detik selama pause
	DurationSec    int             // durasi rencana blok (menit × 60)
	BreakAfterSec  int             // durasi jeda berikutnya (0 = tanpa break)
}

// RemainingSec — sisa detik timer dihitung dari SELISIH TIMESTAMP, bukan tick.
// Jika RUNNING : elapsed = now - StartedAt - PausedAccum
// Jika PAUSED : elapsed dibekukan pada saat pause (caller menyimpan snapshot).
func (t *TimerState) RemainingSec(now time.Time) int {
	if t.Mode == model.TimerIdle || t.DurationSec <= 0 {
		return 0
	}
	elapsed := int(now.Sub(t.StartedAt).Seconds()) - t.PausedAccumSec
	if t.Mode == model.TimerPaused && elapsed < 0 {
		elapsed = 0
	}
	if elapsed < 0 {
		elapsed = 0
	}
	rem := t.DurationSec - elapsed
	if rem < 0 {
		return 0 // overrun ditangani layar (minta keputusan user)
	}
	return rem
}

// ElapsedSec — sudah berjalan berapa detik (untuk progress ring).
func (t *TimerState) ElapsedSec(now time.Time) int {
	e := int(now.Sub(t.StartedAt).Seconds()) - t.PausedAccumSec
	if e < 0 {
		e = 0
	}
	return e
}

// -----------------------------------------------------------------------------
// Settings accessors — AppSetting key/value (dark mode persist antar sesi).
// -----------------------------------------------------------------------------

// GetSetting membaca satu nilai setting ("" jika belum ada).
func (a *App) GetSetting(key string) string {
	var s model.AppSetting
	if err := a.DB.Where("key = ?", key).First(&s).Error; err != nil {
		return ""
	}
	return s.Value
}

// SetSetting menulis/meng-upsert nilai setting.
func (a *App) SetSetting(key, value string) error {
	s := model.AppSetting{Key: key, Value: value}
	return a.DB.Save(&s).Error
}

// LoadDarkMode — sinkronkan flag dark dari DB saat startup.
func (a *App) LoadDarkMode() {
	if a.GetSetting(model.SettDarkMode) == "false" {
		*a.DarkMode = false
	} else {
		*a.DarkMode = true // default: gelap (guideline BAGIAN 3: dark wajib)
	}
}

// ToggleDarkMode — ganti mode + persist + refresh seluruh window.
// Dipanggil oleh Settings screen dan tombol cepat di toolbar Master Map.
func (a *App) ToggleDarkMode(refreshAll func()) {
	*a.DarkMode = !*a.DarkMode
	_ = a.SetSetting(model.SettDarkMode, boolStr(*a.DarkMode))
	if refreshAll != nil {
		refreshAll()
	}
}

// -----------------------------------------------------------------------------
// Tick hook — callback 1 detik yang dipasang layar aktif (Focus Mode).
// main.go memanggil Tick() dari goroutine ticker; kita marshaling ke UI
// thread dengan fyne.Do agar mutasi widget aman.
// -----------------------------------------------------------------------------

// SetTick memasang (atau menghapus dengan nil) callback per-tick.
func (a *App) SetTick(fn func()) {
	a.tickMu.Lock()
	a.onTick = fn
	a.tickMu.Unlock()
}

// Tick dieksekusi oleh ticker app (1 Hz) — dipakai Focus Mode untuk
// me-render selisih timestamp tanpa pernah jadi sumber kebenaran.
// Catatan: pada Fyne v2.5 pembaruan widget dari goroutine lain di-refresh
// lewat canvas.Refresh (thread-safe); kita tetap memakai lock utk membaca fn.
func (a *App) Tick() {
	a.tickMu.Lock()
	fn := a.onTick
	a.tickMu.Unlock()
	if fn != nil {
		fn()
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
