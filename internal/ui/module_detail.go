// ============================================================================
// LAYAR: Modul Detail — halaman per-kategori (BAGIAN 4 no.2)
// ----------------------------------------------------------------------------//
// Menampilkan sub-modul milik satu kategori + daftar quest (project) di
// dalamnya. Tombol "+" membuka form tambah-quest yang TERKUNCI pada kategori
// ini → memenuhi requirement "bisa menambah quest di setiap kategorinya".
// ============================================================================
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mastermap/internal/model"
	mtheme "mastermap/internal/theme"
	"mastermap/internal/ui/components"
)

// BuildModuleDetail — layar daftar project untuk satu kategori.
func BuildModuleDetail(sh *Shell, key model.CategoryKey) fyne.CanvasObject {
	dark := *sh.App.DarkMode
	cc := mtheme.Palette[key]

	// ---------- Header dengan identitas warna kategori -----------------------
	btnAdd := widget.NewButtonWithIcon("Quest", theme.ContentAddIcon(), func() {
		openQuickAddQuest(sh, key) // langsung ke form, kategori sudah terpilih
	})
	header := HeaderBar(cc.Icon+"  "+cc.Name, sh.Nav.Pop, btnAdd)

	// ---------- Ringkasan bobot kognitif kategori ----------------------------
	info := components.SectionTitle("Sub-Modul & Bobot Kognitif", dark)

	var subs []model.SubModule
	sh.App.DB.Where("category_key = ? AND is_active = 1", key).
		Order("sort_order").Find(&subs)

	subRows := make([]fyne.CanvasObject, 0, len(subs))
	for _, s := range subs {
		desc := canvasText(s.Description, 12, mtheme.MutedColor(dark))
		row := container.NewBorder(nil, nil, nil,
			components.WeightChip(int(s.CognitiveWeight), dark),
			container.NewVBox(canvasText(s.Name, 14, mtheme.ForegroundColor(dark)), desc))
		subRows = append(subRows, row, components.Spacer(6))
	}

	// ---------- Daftar Quest (project) milik kategori ini --------------------
	listTitle := components.SectionTitle("Daftar Quest", dark)
	var projects []model.Project
	sh.App.DB.Where("category_key = ?", key).
		Order("priority, updated_at desc").Find(&projects)

	qRows := make([]fyne.CanvasObject, 0, len(projects))
	if len(projects) == 0 {
		qRows = append(qRows, canvasText(
			"Belum ada quest. Tekan tombol \"+ Quest\" untuk mulai.",
			13, mtheme.MutedColor(dark)))
	}
	for i := range projects {
		p := projects[i]
		qRows = append(qRows, projectRow(sh, cc, &p, dark))
	}

	body := container.NewVScroll(container.NewPadded(container.NewVBox(
		info, components.Spacer(6),
		container.NewVBox(subRows...),
		components.Spacer(14), listTitle, components.Spacer(6),
		container.NewVBox(qRows...),
	)))
	return PageScaffold(header, body, nil)
}

// projectRow — satu baris quest: nama, status chip, progress checklist.
func projectRow(sh *Shell, cc *mtheme.CatColor, p *model.Project, dark bool) fyne.CanvasObject {
	var total, done int64
	sh.App.DB.Model(&model.SubTask{}).Where("project_id = ?", p.ID).Count(&total)
	sh.App.DB.Model(&model.SubTask{}).Where("project_id = ? AND is_checked = 1", p.ID).Count(&done)
	pct := 0.0
	if total > 0 {
		pct = float64(done) / float64(total)
	}

	name := canvasText(p.Name, 15, mtheme.ForegroundColor(dark))
	name.TextStyle = fyne.TextStyle{Bold: true}
	meta := canvasText(fmt.Sprintf("%s · P%d · %d/%d fase", projStatusLabel(p.Status), p.Priority, done, total),
		12, mtheme.MutedColor(dark))
	bar := components.NewProgressBar(sh.App.DarkMode, cc.Accent, pct, fmt.Sprintf("%.0f%%", pct*100))

	card := components.NewCard(sh.App.DarkMode, cc.CardTint, cc.Accent,
		container.NewVBox(
			container.NewBorder(nil, nil, nil, statusBadge(p.Status, dark),
				container.NewVBox(name, meta)),
			components.Spacer(6), bar))

	tap := widget.NewButton("", func() {
		pid := p.ID
		sh.Nav.Push(fmt.Sprintf("quest:%d", pid), p.Name, func() fyne.CanvasObject {
			return BuildProjectDetail(sh, pid)
		})
	})
	tap.Importance = widget.LowImportance
	return container.NewStack(card, container.NewMax(tap))
}

func projStatusLabel(s model.ProjectStatus) string {
	switch s {
	case model.ProjPlanned:
		return "Terencana"
	case model.ProjInProgress:
		return "Berjalan"
	case model.ProjOnHold:
		return "Ditahan"
	case model.ProjDone:
		return "Selesai"
	default:
		return "Arsip"
	}
}

func statusBadge(s model.ProjectStatus, dark bool) fyne.CanvasObject {
	col := mtheme.ForegroundColor(dark)
	switch s {
	case model.ProjInProgress:
		return components.ChipBadge("● Berjalan", greenOf(dark), tint(greenOf(dark)))
	case model.ProjDone:
		return components.ChipBadge("✓ Selesai", col, tint(col))
	default:
		return components.ChipBadge("○ "+projStatusLabel(s), mtheme.MutedColor(dark), tint(mtheme.MutedColor(dark)))
	}
}
