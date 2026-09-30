// ============================================================================
// LAYAR: Quest Detail — sub-task modular + checklist (BAGIAN 4 no.3)
// ----------------------------------------------------------------------------
// Menampilkan 5 fase kerja sebuah project (Riset → Desain → Implementasi →
// Testing → Dokumentasi) sebagai checklist berwarna sesuai kategori induknya.
// Time-boxing per task akan dipasang oleh engine (LANGKAH 5).
// ============================================================================
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mastermap/internal/model"
	mtheme "mastermap/internal/theme"
	"mastermap/internal/ui/components"
)

// BuildProjectDetail — layar detail satu quest berdasarkan ID.
func BuildProjectDetail(sh *Shell, pid uint) fyne.CanvasObject {
	dark := *sh.App.DarkMode

	var p model.Project
	if err := sh.App.DB.First(&p, pid).Error; err != nil {
		return container.NewCenter(widget.NewLabel("Quest tidak ditemukan."))
	}
	cc := mtheme.Palette[p.CategoryKey]

	header := HeaderBar(cc.Icon+"  "+p.Name, sh.Nav.Pop,
		widget.NewButtonWithIcon("", theme.ListIcon(), func() { showFolderTemplate(sh, &p) }))

	// Ringkasan progress keseluruhan quest.
	var total, done int64
	sh.App.DB.Model(&model.SubTask{}).Where("project_id = ?", p.ID).Count(&total)
	sh.App.DB.Model(&model.SubTask{}).Where("project_id = ? AND is_checked = 1", p.ID).Count(&done)
	pct := 0.0
	if total > 0 {
		pct = float64(done) / float64(total)
	}

	ring := components.NewRing(72, pct, cc.Accent(dark), fmt.Sprintf("%.0f%%", pct*100))
	summ := container.NewBorder(nil, nil, ring, nil,
		container.NewVBox(
			canvasText(projStatusLabel(p.Status)+" · Prioritas P"+fmt.Sprint(p.Priority),
				13, mtheme.ForegroundColor(dark)),
			canvasText(p.Description, 12, mtheme.MutedColor(dark))))

	// Checklist sub-task (fase modular).
	title := components.SectionTitle("Fase Kerja Modular", dark)
	var tasks []model.SubTask
	sh.App.DB.Where("project_id = ?", p.ID).Order("phase_order").Find(&tasks)

	rows := make([]fyne.CanvasObject, 0, len(tasks))
	for i := range tasks {
		rows = append(rows, subTaskRow(sh, cc, &tasks[i], dark))
	}

	body := container.NewVScroll(container.NewPadded(container.NewVBox(
		summ, components.Spacer(14), title, components.Spacer(6),
		container.NewVBox(rows...))))
	return PageScaffold(header, body, nil)
}

// subTaskRow — checkbox + bobot + tombol time-box (placeholder LANGKAH 5).
func subTaskRow(sh *Shell, cc *mtheme.CatColor, t *model.SubTask, dark bool) fyne.CanvasObject {
	chk := widget.NewCheck("", func(v bool) {
		sh.App.DB.Model(t).Updates(map[string]any{"is_checked": v})
		sh.Nav.Reload()
	})
	chk.SetChecked(t.IsChecked)

	name := canvasText(fmt.Sprintf("%d. %s", t.PhaseOrder, t.Name), 14, mtheme.ForegroundColor(dark))
	meta := canvasText(fmt.Sprintf("Bobot: %s · Estimasi: %d mnt", weightLabel(int(t.Complexity)), t.EstimateMin),
		11, mtheme.MutedColor(dark))

	card := components.NewCard(sh.App.DarkMode, cc.CardTint, cc.Accent,
		container.NewBorder(nil, nil, chk,
			components.WeightChip(int(t.Complexity), dark),
			container.NewVBox(name, meta)))
	return card
}

func weightLabel(w int) string {
	switch w {
	case 1:
		return "Ringan (Blok A)"
	case 3:
		return "Berat (Blok B)"
	default:
		return "Menengah"
	}
}

// showFolderTemplate — info struktur folder standar project (docs/, scripts/…).
func showFolderTemplate(sh *Shell, p *model.Project) {
	txt := p.FolderTemplate
	if txt == "" {
		txt = model.DefaultFolderTemplate
	}
	dialog.ShowInformation("Struktur Folder Project", txt, sh.Win)
}
