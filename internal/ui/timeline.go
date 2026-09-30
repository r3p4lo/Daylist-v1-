// ============================================================================
// LAYAR: Daily Timeline — JADWAL HARI INI (BAGIAN 4 layar #4, LANGKAH 6)
// ----------------------------------------------------------------------------
// Menampilkan rantai TimeBlock hari ini sebagai garis waktu vertikal dengan
// color-coding kategori (warna identitas BAGIAN 1). Fitur utama:
//   - "Susun dengan Engine" → memanggil Dynamic Time-boxing Engine (LANGKAH 5)
//     untuk sub-task aktif pada project yang dipilih, menghormati budget
//     energi harian DailyPlan (anti satu kategori menyedot semua energi).
//   - User dapat OVERRIDE: memilih durasi manual sebelum menyimpan → tercatat
//     Source = USER_OVERRIDE (aturan BAGIAN 2: terima ATAU override).
//   - Tap blok PLANNED → masuk Focus Mode; blok ACTIVE diberi penanda.
//
// ============================================================================
package ui

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"gorm.io/gorm"

	"mastermap/internal/db"
	"mastermap/internal/engine"
	"mastermap/internal/model"
	mtheme "mastermap/internal/theme"
	"mastermap/internal/timer"
	"mastermap/internal/ui/components"
)

// BuildTimeline — layar jadwal hari ini.
func BuildTimeline(sh *Shell) fyne.CanvasObject {
	dark := *sh.App.DarkMode
	gdb := sh.App.DB

	plan, err := db.TodayDailyPlan(gdb)
	if err != nil {
		plan = &model.DailyPlan{PlanDate: time.Now().Format("2006-01-02"), EnergyBudgetMin: 240}
	}

	var blocks []model.TimeBlock
	gdb.Where("daily_plan_id = ?", plan.ID).Order("sequence_no asc").Find(&blocks)

	// Ringkasan pemakaian energi hari ini.
	workMin := 0
	doneMin := 0
	for _, b := range blocks {
		if !b.IsBreak {
			workMin += b.DurationMin
			if b.Status == model.BlkDone || b.Status == model.BlkOverrun {
				doneMin += b.DurationMin
			}
		}
	}

	title := widget.NewLabelWithStyle("Timeline Hari Ini", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	sub := canvas.NewText(longDate(), mtheme.MutedColor(dark))
	sub.TextSize = 13

	budgetTxt := canvas.NewText(
		fmt.Sprintf("Energi terpakai: %d / %d menit · Selesai: %d menit", workMin, plan.EnergyBudgetMin, doneMin),
		mtheme.ForegroundColor(dark))
	budgetTxt.TextSize = 13

	btnTheme := widget.NewButtonWithIcon("", theme.MediaPhotoIcon(), func() { sh.toggleTheme() })
	btnTheme.Importance = widget.LowImportance
	header := container.NewBorder(nil, nil, container.NewVBox(title, sub, Spacer(4), budgetTxt),
		container.NewHBox(btnTheme))

	// ---------- Daftar blok (atau empty-state) ----------
	// Blok dibangun sebagai kartu VBox dalam scroll — kontrol tampilan penuh
	// (emoji + warna kategori) lebih presisi daripada widget.List.
	rows := make([]fyne.CanvasObject, 0, len(blocks)+2)
	if len(blocks) == 0 {
		empty := components.SectionTitle("Belum ada blok waktu hari ini", dark)
		hint := canvas.NewText("Tekan tombol \"Susun dengan Engine\" untuk membuat time-block dari sub-task aktif.",
			mtheme.MutedColor(dark))
		hint.TextSize = 13
		rows = append(rows, empty, hint, Spacer(8))
	}
	for i := range blocks {
		rows = append(rows, timelineRow(sh, &blocks[i]))
	}
	body := PageScaffold(header, container.NewVScroll(container.NewPadded(container.NewVBox(rows...))), nil)

	// ---------- FAB: generate dari engine ----------
	fab := widget.NewButtonWithIcon("Susun dengan Engine", theme.ViewRefreshIcon(), func() {
		openEngineDialog(sh, plan)
	})
	fab.Importance = widget.HighImportance
	return fabWrap(body, fab)
}

// fabWrap — menempelkan FAB ke scaffold via PageScaffold ulang.
func fabWrap(body fyne.CanvasObject, fab *widget.Button) fyne.CanvasObject {
	if c, ok := body.(*fyne.Container); ok && len(c.Objects) > 0 {
		c.Objects[len(c.Objects)-1] = fab // slot FAB pada PageScaffold
		canvas.Refresh(c)
	}
	return body
}

// timelineRow — satu entri garis waktu: titik warna + strip durasi + aksi.
func timelineRow(sh *Shell, b *model.TimeBlock) fyne.CanvasObject {
	dark := *sh.App.DarkMode
	cc := mtheme.Palette[b.CategoryKey]
	accent := cc.Accent(dark)

	// Titik warna kategori pada rail kiri.
	dot := canvas.NewCircle(accent)
	dot.Resize(fyne.NewSize(12, 12))

	timeLbl := canvas.NewText(fmt.Sprintf("%d mnt", b.DurationMin), mtheme.MutedColor(dark))
	timeLbl.TextSize = 12

	name := b.Title
	if b.IsBreak {
		name = "☕ " + name + " (istirahat)"
	}
	lblName := canvas.NewText(name, mtheme.ForegroundColor(dark))
	lblName.TextSize = 15
	lblName.TextStyle = fyne.TextStyle{Bold: !b.IsBreak}

	status := statusBlockLabel(b.Status)
	badge := components.ChipBadge(status, accent, cc.ChipBg(dark))

	row1 := container.NewBorder(nil, nil, dot, badge,
		container.NewPadded(lblName))
	row2 := container.NewPadded(container.NewHBox(timeLbl, Spacer(8)))

	// Aksi: mulai fokus / lihat hasil.
	var act fyne.CanvasObject
	switch b.Status {
	case model.BlkPlanned, model.BlkSkipped:
		act = widget.NewButtonWithIcon("Fokus ▶", theme.MediaPlayIcon(), func() {
			svc := timer.NewService(sh.App.DB)
			if _, err := svc.Start(b.ID); err != nil {
				showToast(sh, "Gagal memulai: "+err.Error())
				return
			}
			pushFocus(sh)
		})
		act.(*widget.Button).Importance = widget.HighImportance
	case model.BlkActive:
		act = widget.NewButton("Lanjutkan Fokus ⏸▶", func() { pushFocus(sh) })
		act.(*widget.Button).Importance = widget.HighImportance
	default:
		act = widget.NewButton("✓ selesai ("+fmt.Sprintf("%d mnt", b.DurationMin)+")", func() {})
		act.(*widget.Button).Importance = widget.LowImportance
	}

	card := components.NewCard(sh.App.DarkMode, cc.CardTint, cc.Accent,
		container.NewVBox(row1, row2, act))
	return card
}

func statusBlockLabel(s model.BlockStatus) string {
	switch s {
	case model.BlkActive:
		return "AKTIF"
	case model.BlkDone:
		return "SELESAI"
	case model.BlkOverrun:
		return "OVERRUN"
	case model.BlkSkipped:
		return "LEWAT"
	default:
		return "TERJADWAL"
	}
}

// openEngineDialog — inti integrasi UI ↔ Engine (LANGKAH 5):
// pilih project → sub-task → bobot diambil dari CognitiveWeight sub-modul →
// user memasukkan estimasi total menit → pratinjau rantai blok → simpan
// (ENGINE) atau override durasi manual (USER_OVERRIDE).
func openEngineDialog(sh *Shell, plan *model.DailyPlan) {
	gdb := sh.App.DB
	dark := *sh.App.DarkMode

	var projects []model.Project
	gdb.Where("status <> ?", model.ProjDone).Order("updated_at desc").Find(&projects)
	if len(projects) == 0 {
		showToast(sh, "Belum ada project aktif — tambah quest dulu ya.")
		return
	}

	projNames := make([]string, len(projects))
	for i, p := range projects {
		projNames[i] = p.Name
	}

	var selProject int
	estEntry := widget.NewEntry()
	estEntry.SetPlaceHolder("Estimasi total menit (mis. 120)")

	lblPreview := canvas.NewText("Rantai blok akan tampil di sini…", mtheme.MutedColor(dark))
	lblPreview.TextSize = 13

	overrideSel := widget.NewSelect([]string{"Terima rekomendasi Engine", "Override manual (menit/blok)"}, nil)
	overrideSel.SetSelectedIndex(0)
	manualEntry := widget.NewEntry()
	manualEntry.SetPlaceHolder("Durasi manual per blok, mis. 40")

	formItems := []*widget.FormItem{
		widget.NewFormItem("Project", widget.NewSelect(projNames, func(s string) {
			for i, p := range projects {
				if p.Name == s {
					selProject = i
					break
				}
			}
		})),
		widget.NewFormItem("Total (menit)", estEntry),
		widget.NewFormItem("Sumber durasi", overrideSel),
		widget.NewFormItem("Override", manualEntry),
		widget.NewFormItem("Pratinjau", lblPreview),
	}

	dlg := dialog.NewForm("⚙️ Susun Time-Block Otomatis",
		"Susun", "Batal", formItems, func(ok bool) {
			if !ok {
				return
			}
			total, _ := strconv.Atoi(estEntry.Text)
			if total <= 0 {
				showToast(sh, "Isi estimasi total menit lebih dulu.")
				return
			}
			p := projects[selProject]
			weight := weightForProject(gdb, p.ID)

			source := "ENGINE"
			var blocks []engine.Block
			if overrideSel.Selected == "Override manual (menit/blok)" {
				man, _ := strconv.Atoi(manualEntry.Text)
				if man < 5 {
					showToast(sh, "Override minimal 5 menit per blok.")
					return
				}
				source = "USER_OVERRIDE"
				blocks = manualChain(weight, total, man)
			} else {
				var overflow int
				blocks, overflow = engine.PlanWithBudget(engine.Params{
					Weight: weight, TotalMin: total, Complexity: complexityOf(p),
				}, plan.EnergyBudgetMin)
				if overflow > 0 {
					showToast(sh, fmt.Sprintf("Budget energi habis — %d menit ditunda besok.", overflow))
				}
			}
			if len(blocks) == 0 {
				showToast(sh, "Tidak ada blok dihasilkan.")
				return
			}
			tbs := engine.ToTimeBlocks(blocks, plan.ID, p.CategoryKey, p.Name, firstSubTaskID(gdb, p.ID), source)
			seqBase := nextSeq(gdb, plan.ID)
			for i := range tbs {
				tbs[i].SequenceNo = seqBase + i
			}
			if err := gdb.Create(&tbs).Error; err != nil {
				showToast(sh, "Gagal menyimpan: "+err.Error())
				return
			}
			sh.Nav.Reload()
			showToast(sh, fmt.Sprintf("✔ %d blok disimpan (%s)", len(tbs), source))
		}, sh.Win)
	dlg.Resize(fyne.NewSize(420, 520))
	dlg.Show()
}

// manualChain — rantai blok seragam utk override user (+ jeda sesuai bobot).
func manualChain(w model.CognitiveWeight, total, perBlock int) []engine.Block {
	brk := model.BreakShortMinutes
	if w == model.WeightHeavy {
		brk = model.BreakLongMinutes
	}
	var out []engine.Block
	typ := model.BlockA
	if w != model.WeightLight {
		typ = model.BlockB
	}
	rem := total
	for rem > 0 {
		take := perBlock
		if rem < take {
			take = rem
		}
		out = append(out, engine.Block{Type: typ, DurationMin: take})
		rem -= take
		if rem >= 10 {
			out = append(out, engine.Block{Type: typ, DurationMin: brk, IsBreak: true})
		}
	}
	return out
}

// weightForProject — bobot kognitif dari sub-modul project (fallback: sedang).
func weightForProject(gdb *gorm.DB, pid uint) model.CognitiveWeight {
	var p model.Project
	if gdb.First(&p, pid).Error != nil {
		return model.WeightMedium
	}
	var sm model.SubModule
	if gdb.First(&sm, p.SubModuleID).Error == nil {
		return model.CognitiveWeight(sm.CognitiveWeight)
	}
	return model.WeightMedium
}

func complexityOf(p model.Project) float64 {
	switch p.Priority {
	case 1:
		return 0.8
	case 2:
		return 0.5
	case 3:
		return 0.25
	default:
		return 0.1
	}
}

func firstSubTaskID(gdb *gorm.DB, pid uint) uint {
	var t model.SubTask
	if gdb.Where("project_id = ? AND is_checked = ?", pid, false).Order("phase_order asc").First(&t).Error == nil {
		return t.ID
	}
	return 0
}

func nextSeq(gdb *gorm.DB, planID uint) int {
	var maxSeq int64
	gdb.Model(&model.TimeBlock{}).Where("daily_plan_id = ?", planID).
		Select("COALESCE(MAX(sequence_no),0)").Scan(&maxSeq)
	return int(maxSeq) + 1
}
