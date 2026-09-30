// ============================================================================
// LAYAR: Focus Mode — ACTIVE TIMER (BAGIAN 4 layar #5, LANGKAH 6)
// ----------------------------------------------------------------------------
// UI minimalis fullscreen-style saat time-block berjalan. Prinsip desain:
//   - Hening & fokus: hanya ada warna kategori, ring progres, countdown besar.
//   - Akurasi lintas-minimize: angka TIDAK dihitung dari tick UI, melainkan
//     dari SELISIH TIMESTAMP di DB (timer.Service.viewFor):
//     sisa = DurationSec − (now − StartedAt − PausedAccumSec)
//     Jika Android membunuh proses, sesi tetap tersimpan; saat app dibuka
//     lagi, Snapshot() merekonstruksi kondisi RUNNING/PAUSED dengan benar.
//   - Ticker 1 detik me-"render" selisih — bukan menjadi sumber kebenaran.
//
// ============================================================================
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mastermap/internal/model"
	mtheme "mastermap/internal/theme"
	"mastermap/internal/timer"
	"mastermap/internal/ui/components"
)

// pushFocus — navigasi ke Focus Mode + pasang builder di Registry agar layar
// ini ikut ter-rebuild bila user men-toggle tema atau Reload dipanggil.
func pushFocus(sh *Shell) {
	sh.Nav.Push("focus", "Fokus", func() fyne.CanvasObject { return BuildFocusMode(sh) })
}

// BuildFocusMode — membangun layar timer dari snapshot Service (bukan state UI).
func BuildFocusMode(sh *Shell) fyne.CanvasObject {
	dark := *sh.App.DarkMode
	svc := timer.NewService(sh.App.DB)
	view, ok := svc.Snapshot()
	if !ok {
		// Tidak ada sesi aktif → kembali rapi ke Timeline.
		showToast(sh, "Tidak ada sesi fokus aktif.")
		return BuildTimeline(sh)
	}

	cc := mtheme.Palette[view.CategoryKey]
	if cc == nil {
		cc = mtheme.Palette["TECHNOLOGY"]
	}
	accent := cc.Accent(dark)

	// ---- Header sederhana: kategori + judul blok --------------------------
	catLbl := canvasText(string(view.CategoryKey), 13, accent)
	catLbl.TextStyle = fyne.TextStyle{Bold: true}
	titleLbl := canvasText(view.Title, 17, mtheme.ForegroundColor(dark))
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	head := container.NewVBox(
		container.NewCenter(catLbl),
		container.NewCenter(titleLbl),
	)

	// ---- Ring progres + countdown besar -----------------------------------
	timeLbl := canvasText(timer.FormatMMSS(view.RemainingSec), 56, mtheme.ForegroundColor(dark))
	timeLbl.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	timeLbl.Alignment = fyne.TextAlignCenter

	modeLbl := canvasText(focusModeLabel(view), 13, mtheme.MutedColor(dark))
	modeLbl.Alignment = fyne.TextAlignCenter

	ring := components.NewRing(230, view.Pct, accent, "")
	core := container.NewStack(ring,
		container.NewPadded(container.NewCenter(
			container.NewVBox(timeLbl, modeLbl))))

	// ---- Info mulai / estimasi selesai ------------------------------------
	startInfo := canvasText(fmt.Sprintf("Mulai %s • target %d menit",
		timer.FormatClock(view.StartedAt), view.DurationSec/60), 12, mtheme.MutedColor(dark))
	startInfo.Alignment = fyne.TextAlignCenter

	// ---- Tombol aksi -------------------------------------------------------
	btnPause := widget.NewButtonWithIcon("Jeda", theme.MediaPauseIcon(), func() {
		v, err := svc.Pause()
		if err != nil {
			showToast(sh, err.Error())
			return
		}
		_ = v
		sh.Nav.Reload() // rebuild → tampilan beku pada waktu pause
	})
	btnResume := widget.NewButtonWithIcon("Lanjut", theme.MediaPlayIcon(), func() {
		if _, err := svc.Resume(); err != nil {
			showToast(sh, err.Error())
			return
		}
		sh.Nav.Reload()
	})
	btnDone := widget.NewButtonWithIcon("Selesai ✓", theme.ConfirmIcon(), func() {
		v, err := svc.Complete()
		if err != nil {
			showToast(sh, err.Error())
			return
		}
		onFocusFinished(sh, svc, v)
	})
	btnDone.Importance = widget.HighImportance

	btnStop := widget.NewButtonWithIcon("Batalkan", theme.CancelIcon(), func() {
		dialog.ShowConfirm("Batalkan sesi fokus?",
			"Sesi akan dicatat sebagai terputus (data tidak hilang).",
			func(yes bool) {
				if !yes {
					return
				}
				_ = svc.Stop()
				sh.Nav.Pop()
			}, sh.Win)
	})
	btnStop.Importance = widget.LowImportance

	var actions *fyne.Container
	switch view.Mode {
	case model.TimerPaused:
		actions = container.NewHBox(btnResume, btnDone, btnStop)
	default: // RUNNING / BREAK
		actions = container.NewHBox(btnPause, btnDone, btnStop)
	}

	body := container.NewVBox(
		Spacer(8), head, Spacer(16),
		container.NewCenter(core), Spacer(8),
		container.NewCenter(startInfo), Spacer(20),
		container.NewCenter(actions),
	)

	header := HeaderBar("🎯 Focus Mode", func() { sh.Nav.Pop() })
	page := PageScaffold(header, body, nil)

	// Update countdown per detik TANPA mengubah sumber kebenaran (DB).
	sh.App.SetTick(func() {
		v, ok := svc.Snapshot()
		if !ok || v == nil {
			return
		}
		timeLbl.Text = timer.FormatMMSS(v.RemainingSec)
		if v.OverrunSec > 0 {
			timeLbl.Text = "+" + timer.FormatMMSS(v.OverrunSec)
		}
		ring.SetProgress(v.Pct, "")
		canvas.Refresh(page)
	})
	return page
}

// onFocusFinished — pasca blok kerja selesai: tawarkan jeda istirahat otomatis
// (rantai engine: [50m, ☕10m, ...]) atau kembali ke Timeline.
func onFocusFinished(sh *Shell, svc *timer.Service, v *timer.View) {
	showToast(sh, "🎉 Blok selesai! Total fokus tercatat.")
	if blk, hasNext := svc.NextBreak(v.BlockID); hasNext {
		dialog.ShowConfirm("Lanjut istirahat? ☕",
			fmt.Sprintf("Jeda %d menit sudah terjadwal setelah blok ini.", blk.DurationMin),
			func(yes bool) {
				if yes {
					if _, err := svc.Start(blk.ID); err != nil {
						showToast(sh, err.Error())
					}
				}
				sh.Nav.Reload()
			}, sh.Win)
		return
	}
	sh.Nav.Reload()
}

func focusModeLabel(v *timer.View) string {
	switch v.Mode {
	case model.TimerRunning:
		return "SEDANG BERFOKUS"
	case model.TimerPaused:
		return "DIJEDA"
	case model.TimerBreak:
		return "ISTIRAHAT — tarik napas"
	case model.TimerFinished:
		return "SELESAI ✔"
	default:
		return ""
	}
}
