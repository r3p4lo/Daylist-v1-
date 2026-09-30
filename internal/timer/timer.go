// ============================================================================
// PACKAGE: timer — Active Timer Engine (BAGIAN 4 layar #5, LANGKAH 6)
// ----------------------------------------------------------------------------
// Prinsip akurasi (sesuai desain LANGKAH 1 §1.5):
//   TIDAK ADA penghitung tick di memori. Seluruh state sesi disimpan sebagai
//   TIMESTAMP di SQLite (tabel focus_sessions & time_blocks). Nilai "sisa
//   detik" kapan pun dihitung dengan RUMUS SELISIH:
//
//     sisa = DurationSec − ( now − StartedAt ) + PausedAccumSec
//
//   Konsekuensi: walau aplikasi di-minimize, dibekukan Android, atau bahkan
//   prosesnya mati total — saat dibuka lagi, posisi timer SELALU benar,
//   karena cukup membaca StartedAt dari DB lalu menghitung selisihnya.
//
// State machine: IDLE → RUNNING ⇄ PAUSED → (BREAK) → DONE / OVERRUN / SKIPPED
// ============================================================================
package timer

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"mastermap/internal/model"
)

// View — proyeksi ringan sesi untuk dirender UI Focus Mode.
type View struct {
	BlockID      uint
	Title        string
	CategoryKey  model.CategoryKey
	DurationSec  int           // durasi rencana blok (detik)
	ElapsedSec   int           // sudah berjalan (detik)
	RemainingSec int           // sisa (0 jika overrun)
	OverrunSec   int           // kelebihan waktu di atas durasi (≥0)
	Pct          float64       // 0..1 untuk progress ring
	Mode         model.TimerMode
	IsBreak      bool          // true bila view ini adalah jeda istirahat
	BreakDurSec  int           // durasi jeda berikutnya (0 = tidak ada)
	StartedAt    time.Time
}

// Service — operasi timer terhadap TimeBlock yang aktif.
type Service struct{ gdb *gorm.DB }

func NewService(gdb *gorm.DB) *Service { return &Service{gdb: gdb} }

var ErrNoActiveSession = errors.New("tidak ada sesi fokus aktif")

// FormatMMSS — tampilan countdown "MM:SS" (menit bisa > 59 utk sesi panjang).
func FormatMMSS(sec int) string {
	if sec < 0 {
		sec = 0
	}
	m, s := sec/60, sec%60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// FormatClock — jam lokal "HH:MM" untuk label timeline.
func FormatClock(t time.Time) string { return t.Format("15:04") }

// startOf —StartedAt *time.Time → time.Time nol-safe.
func startOf(blk *model.TimeBlock) time.Time {
	if blk.StartedAt != nil {
		return *blk.StartedAt
	}
	return blk.CreatedAt
}

// -----------------------------------------------------------------------------
// Start — memulai fokus pada sebuah TimeBlock.
// -----------------------------------------------------------------------------
func (s *Service) Start(blockID uint) (*View, error) {
	now := time.Now()

	var blk model.TimeBlock
	if err := s.gdb.First(&blk, blockID).Error; err != nil {
		return nil, err
	}
	// Tutup sesi aktif lain yang masih menggantung (user pindah blok).
	s.closeOpenSessions(now, true)

	start := now
	blk.StartedAt = &start
	blk.Status = model.BlkActive
	if err := s.gdb.Save(&blk).Error; err != nil {
		return nil, err
	}

	// Blok break: countdown singkat tanpa mencatat FocusSession kerja.
	if blk.IsBreak {
		return &View{
			BlockID: blk.ID, Title: blk.Title, CategoryKey: blk.CategoryKey,
			DurationSec: blk.DurationMin * 60, RemainingSec: blk.DurationMin * 60,
			Mode: model.TimerBreak, IsBreak: true, StartedAt: start,
		}, nil
	}

	sess := model.FocusSession{TimeBlockID: blk.ID, StartedAt: now}
	if err := s.gdb.Create(&sess).Error; err != nil {
		return nil, err
	}
	return s.viewFor(&blk, &sess, now), nil
}

// Pause — membekukan waktu: catat titik pause di EndedAt (sebagai anchor
// pause) dan mulai akumulasi. Resume akan menggeser StartedAt maju sejauh
// lama pause → rumus selisih tetap valid tanpa kolom tambahan.
func (s *Service) Pause() (*View, error) {
	now := time.Now()
	sess, err := s.openSession()
	if err != nil {
		return nil, err
	}
	end := now
	sess.EndedAt = &end // penanda "sedang dipause"
	s.gdb.Save(sess)
	return s.viewOfBlock(sess.TimeBlockID, now), nil
}

// Resume — melanjutkan sesi yang dipause.
func (s *Service) Resume() (*View, error) {
	now := time.Now()
	sess, err := s.pausedSession()
	if err != nil {
		return nil, err
	}
	pauseLen := int(now.Sub(*sess.EndedAt).Seconds())
	if pauseLen < 0 {
		pauseLen = 0
	}
	sess.StartedAt = sess.StartedAt.Add(time.Duration(pauseLen) * time.Second)
	sess.EndedAt = nil
	s.gdb.Save(sess)

	if blk, e := s.block(sess.TimeBlockID); e == nil {
		blk.Status = model.BlkActive
		s.gdb.Save(blk)
	}
	return s.viewOfBlock(sess.TimeBlockID, now), nil
}

// Stop — hentikan sebelum tuntas (interrupted). Blok kembali ke PLANNED.
func (s *Service) Stop() error {
	now := time.Now()
	sess, err := s.anyUnfinished()
	if err != nil {
		// Mungkin yang aktif hanya blok break.
		var blk model.TimeBlock
		if e := s.gdb.Where("status = ? AND is_break = 1", model.BlkActive).First(&blk).Error; e == nil {
			blk.Status = model.BlkSkipped
			return s.gdb.Save(&blk).Error
		}
		return err
	}
	end := now
	sess.EndedAt = &end
	sess.Interrupted = true
	s.gdb.Save(sess)

	if blk, e := s.block(sess.TimeBlockID); e == nil {
		blk.Status = model.BlkPlanned
		blk.StartedAt = nil
		s.gdb.Save(blk)
	}
	return nil
}

// Complete — tandai blok selesai; hitung OverrunMin (umpan balik kalibrasi
// engine) dan laporkan jeda istirahat berikutnya bila ada.
func (s *Service) Complete() (*View, error) {
	now := time.Now()

	// Kasus blok break: langsung DONE.
	var brk model.TimeBlock
	if e := s.gdb.Where("status = ? AND is_break = 1", model.BlkActive).First(&brk).Error; e == nil {
		end := now
		brk.EndedAt = &end
		brk.Status = model.BlkDone
		s.gdb.Save(&brk)
		return &View{BlockID: brk.ID, Title: brk.Title, CategoryKey: brk.CategoryKey,
			DurationSec: brk.DurationMin * 60, Mode: model.TimerFinished, IsBreak: true}, nil
	}

	sess, err := s.anyUnfinished()
	if err != nil {
		return nil, err
	}
	end := now
	sess.EndedAt = &end
	// Bekukan elapsed penuh ke dalam akumulasi bila sedang pause.
	s.gdb.Save(sess)

	blk, err := s.block(sess.TimeBlockID)
	if err != nil {
		return nil, err
	}
	elapsedSec := s.elapsed(blk, sess, end)
	blk.OverrunMin = elapsedSec/60 - blk.DurationMin
	blk.EndedAt = &end
	if blk.OverrunMin > 0 {
		blk.Status = model.BlkOverrun
	} else {
		blk.Status = model.BlkDone
	}
	s.gdb.Save(blk)

	v := s.viewFor(blk, sess, end)
	v.Mode = model.TimerFinished
	if nb, ok := s.NextBreak(blk.ID); ok {
		v.BreakDurSec = nb.DurationMin * 60
	}
	return v, nil
}

// NextBreak — jeda istirahat pertama setelah blok ini pada plan yang sama.
func (s *Service) NextBreak(blockID uint) (*model.TimeBlock, bool) {
	blk, err := s.block(blockID)
	if err != nil {
		return nil, false
	}
	var brk model.TimeBlock
	err = s.gdb.Where("daily_plan_id = ? AND sequence_no > ? AND is_break = 1 AND status = ?",
		blk.DailyPlanID, blk.SequenceNo, model.BlkPlanned).
		Order("sequence_no asc").First(&brk).Error
	if err != nil {
		return nil, false
	}
	return &brk, true
}

// -----------------------------------------------------------------------------
// Snapshot — rekonstruksi View APA PUN saat UI dibuka (termasuk sesudah
// app di-minimize / proses dibunuh). Inti akurasi timestamp-based.
// -----------------------------------------------------------------------------
func (s *Service) Snapshot() (*View, bool) {
	now := time.Now()
	// Prioritas 1: sesi kerja aktif/paused.
	var sess model.FocusSession
	err := s.gdb.Where("ended_at IS NULL OR (ended_at IS NOT NULL AND interrupted = 0 AND "+
		"time_block_id IN (SELECT id FROM time_blocks WHERE status = ?))", model.BlkActive).
		Order("started_at desc").First(&sess).Error
	if err == nil {
		return s.viewOfBlock(sess.TimeBlockID, now), true
	}
	// Prioritas 2: blok break yang sedang jalan.
	var blk model.TimeBlock
	if err := s.gdb.Where("status = ? AND is_break = 1", model.BlkActive).
		Order("updated_at desc").First(&blk).Error; err == nil {
		dur := blk.DurationMin * 60
		el := int(now.Sub(startOf(&blk)).Seconds())
		if el < 0 {
			el = 0
		}
		rem := dur - el
		if rem < 0 {
			rem = 0
		}
		return &View{BlockID: blk.ID, Title: blk.Title, CategoryKey: blk.CategoryKey,
			DurationSec: dur, ElapsedSec: el, RemainingSec: rem, Mode: model.TimerBreak,
			IsBreak: true, StartedAt: startOf(&blk)}, true
	}
	return nil, false
}

// anyUnfinished — sesi yang belum tuntas normal: ended_at NULL (jalan) ATAU
// paused (ended_at terisi sebagai anchor pause, blok masih ACTIVE).
func (s *Service) anyUnfinished() (*model.FocusSession, error) {
	var sess model.FocusSession
	err := s.gdb.Where("time_block_id IN (SELECT id FROM time_blocks WHERE status = ?)",
		model.BlkActive).Order("started_at desc").First(&sess).Error
	if err != nil {
		return nil, ErrNoActiveSession
	}
	return &sess, nil
}

func (s *Service) openSession() (*model.FocusSession, error) {
	var sess model.FocusSession
	err := s.gdb.Where("ended_at IS NULL AND time_block_id IN "+
		"(SELECT id FROM time_blocks WHERE status = ? AND is_break = 0)", model.BlkActive).
		Order("started_at desc").First(&sess).Error
	if err != nil {
		return nil, ErrNoActiveSession
	}
	return &sess, nil
}

func (s *Service) pausedSession() (*model.FocusSession, error) {
	var sess model.FocusSession
	err := s.gdb.Where("ended_at IS NOT NULL AND time_block_id IN "+
		"(SELECT id FROM time_blocks WHERE status = ? AND is_break = 0)", model.BlkActive).
		Order("started_at desc").First(&sess).Error
	if err != nil {
		return nil, ErrNoActiveSession
	}
	return &sess, nil
}

func (s *Service) block(id uint) (*model.TimeBlock, error) {
	var blk model.TimeBlock
	if err := s.gdb.First(&blk, id).Error; err != nil {
		return nil, err
	}
	return &blk, nil
}

func (s *Service) closeOpenSessions(now time.Time, interrupted bool) {
	var list []model.FocusSession
	s.gdb.Where("time_block_id IN (SELECT id FROM time_blocks WHERE status = ?)",
		model.BlkActive).Find(&list)
	for i := range list {
		end := now
		list[i].EndedAt = &end
		list[i].Interrupted = interrupted
		s.gdb.Save(&list[i])
	}
}

func (s *Service) viewOfBlock(blkID uint, now time.Time) *View {
	blk, err := s.block(blkID)
	if err != nil {
		return nil
	}
	var sess model.FocusSession
	if s.gdb.Where("time_block_id = ?", blkID).Order("started_at desc").First(&sess).Error == nil {
		return s.viewFor(blk, &sess, now)
	}
	return &View{BlockID: blk.ID, Title: blk.Title, CategoryKey: blk.CategoryKey,
		DurationSec: blk.DurationMin * 60, Mode: model.TimerIdle}
}

// elapsed — RUMUS SELISIH tunggal. Sesi paused: beku pada anchor EndedAt.
func (s *Service) elapsed(blk *model.TimeBlock, sess *model.FocusSession, now time.Time) int {
	ref := now
	if sess.EndedAt != nil && blk.Status == model.BlkActive {
		ref = *sess.EndedAt // sedang pause → waktu berhenti di titik pause
	}
	e := int(ref.Sub(sess.StartedAt).Seconds()) - sess.PausedAccumSec
	if e < 0 {
		e = 0
	}
	return e
}

// viewFor — bangun View lengkap dari block + sesi.
func (s *Service) viewFor(blk *model.TimeBlock, sess *model.FocusSession, now time.Time) *View {
	dur := blk.DurationMin * 60
	elapsed := s.elapsed(blk, sess, now)
	mode := model.TimerRunning
	switch {
	case blk.Status == model.BlkDone || blk.Status == model.BlkOverrun:
		mode = model.TimerFinished
	case sess.EndedAt != nil && blk.Status == model.BlkActive:
		mode = model.TimerPaused
	}
	rem, over := dur-elapsed, 0
	if rem < 0 {
		over, rem = -rem, 0
	}
	pct := float64(elapsed) / float64(dur)
	if pct > 1 {
		pct = 1
	}
	return &View{
		BlockID: blk.ID, Title: blk.Title, CategoryKey: blk.CategoryKey,
		DurationSec: dur, ElapsedSec: elapsed, RemainingSec: rem, OverrunSec: over,
		Pct: pct, Mode: mode, IsBreak: blk.IsBreak, StartedAt: sess.StartedAt,
	}
}
