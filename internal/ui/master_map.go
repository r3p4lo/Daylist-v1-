// ============================================================================
// LAYAR: Master Map — HOME (BAGIAN 3)
// ----------------------------------------------------------------------------
// Tampilan "Peta Induk": satu layar yang merangkum KEENAM modul kehidupan.
// Prinsip UX yang diimplementasikan di sini:
//   • Color-coding konsisten → tiap kartu memakai tint + strip aksen kategori.
//   • Hierarki visual        → judul besar, statistik ringkas, bar progres.
//   • Quick action           → FAB "+ Tambah Quest" selalu tersedia; setiap
//                              quest baru DAPAT dibuat dari dalam kategori
//                              masing-masing (dialog pilih sub-modul per modul).
//   • Dark/Light mode        → seluruh teks kustom dibangun ulang lewat
//                              Registry saat toggle (lihat navigator.go).
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

	"mastermap/internal/model"
	"mastermap/internal/state"
	mtheme "mastermap/internal/theme"
	"mastermap/internal/ui/components"
)

// Shell — dependensi bersama semua layar.
type Shell struct {
	App *state.App
	Nav *Navigator
	Win fyne.Window
}

// toggleTheme — ganti Mode Terang/Gelap lalu bangun ulang layar aktif,
// sehingga seluruh teks kustom (canvas.Text) ikut warna mode baru.
func (sh *Shell) toggleTheme() {
	sh.App.ToggleDarkMode(func() { sh.Nav.Reload() })
}

// longDate — label tanggal humanis berbahasa Indonesia di header Home.
func longDate() string {
	idn := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	names := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli",
		"Agustus", "September", "Oktober", "November", "Desember"}
	now := time.Now()
	return fmt.Sprintf("%s, %d %s %d", idn[int(now.Weekday())], now.Day(), names[now.Month()-1], now.Year())
}

// catStats — agregasi ringan untuk kartu kategori.
type catStats struct {
	subs     int64
	projects int64
	done     int64
	total    int64
	hours    float64 // akumulasi fokus nyata (jam)
}

func statsFor(gdb *gorm.DB, key model.CategoryKey) catStats {
	var s catStats
	gdb.Model(&model.SubModule{}).Where("category_key = ? AND is_active = 1", key).Count(&s.subs)
	gdb.Model(&model.Project{}).Where("category_key = ?", key).Count(&s.projects)
	gdb.Model(&model.SubTask{}).
		Joins("JOIN projects ON projects.id = sub_tasks.project_id").
		Where("projects.category_key = ?", key).Count(&s.total)
	gdb.Model(&model.SubTask{}).
		Joins("JOIN projects ON projects.id = sub_tasks.project_id").
		Where("projects.category_key = ? AND sub_tasks.is_checked = 1", key).Count(&s.done)
	var agg struct{ A int64 }
	gdb.Model(&model.Project{}).Select("COALESCE(SUM(actual_total_sec),0) AS a").
		Where("category_key = ?", key).Scan(&agg)
	s.hours = float64(agg.A) / 3600
	return s
}

// BuildMasterMap membuat layar home. Dipanggil ulang oleh Registry saat tema
// berganti, sehingga warna teks kustom ikut mode baru.
func BuildMasterMap(sh *Shell) fyne.CanvasObject {
	dark := *sh.App.DarkMode

	// ---------- Header: greeting + tanggal + tombol tema & settings ----------
	dateTxt := canvas.NewText(longDate(), mtheme.MutedColor(dark))
	dateTxt.TextSize = 13
	title := widget.NewLabelWithStyle("Master Map", fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true})

	btnTheme := widget.NewButtonWithIcon("", theme.MediaPhotoIcon(), func() { sh.toggleTheme() })
	btnTheme.Importance = widget.LowImportance
	btnSettings := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		sh.Nav.Push("settings", "Pengaturan", func() fyne.CanvasObject { return BuildSettings(sh) })
	})
	btnSettings.Importance = widget.LowImportance

	header := container.NewBorder(nil, nil,
		container.NewVBox(title, dateTxt),
		container.NewHBox(btnTheme, btnSettings))

	// ---------- Kartu 6 kategori (grid responsif) ----------------------------
	cards := make([]fyne.CanvasObject, 0, len(mtheme.OrderedPalette))
	for _, cc := range mtheme.OrderedPalette {
		cards = append(cards, categoryCard(sh, cc))
	}
	grid := container.NewGridWrap(fyne.NewSize(340, 150), cards...)

	list := container.NewVScroll(container.NewPadded(grid))

	// ---------- FAB: tambah quest langsung dari Home (pilih kategori dulu) ---
	fab := widget.NewButtonWithIcon("Tambah Quest", theme.ContentAddIcon(), func() {
		openQuickAddQuest(sh, "")
	})
	fab.Importance = widget.HighImportance

	body := PageScaffold(header, list, fab)
	return body
}

// categoryCard — satu kartu modul pada Peta Induk.
func categoryCard(sh *Shell, cc *mtheme.CatColor) fyne.CanvasObject {
	dark := *sh.App.DarkMode
	st := statsFor(sh.App.DB, cc.Key)
	pct := 0.0
	if st.total > 0 {
		pct = float64(st.done) / float64(st.total)
	}

	icon := canvas.NewText(cc.Icon, cc.Accent(dark))
	icon.TextSize = 22
	name := canvas.NewText(cc.Name, mtheme.ForegroundColor(dark))
	name.TextSize = 16
	name.TextStyle = fyne.TextStyle{Bold: true}
	meta := canvas.NewText(
		fmt.Sprintf("%d project · %d sub-modul · %.1f jam fokus", st.projects, st.subs, st.hours),
		mtheme.MutedColor(dark))
	meta.TextSize = 12

	bar := components.NewProgressBar(sh.App.DarkMode, cc.Accent, pct,
		fmt.Sprintf("%.0f%%", pct*100))

	headRow := container.NewBorder(nil, nil, icon, nil,
		container.NewVBox(name, meta))

	body := container.NewVBox(headRow, components.Spacer(8), bar)
	card := components.NewCard(sh.App.DarkMode, cc.CardTint, cc.Accent, body)

	// Klik kartu → buka halaman detail modul (kategori) tersebut.
	btnTap := widget.NewButton("", func() {
		key := string(cc.Key)
		sh.Nav.Push("modul:"+key, cc.Name, func() fyne.CanvasObject {
			return BuildModuleDetail(sh, cc.Key)
		})
	})
	btnTap.Importance = widget.LowImportance
	// Tombol transparan mengisi seluruh kartu → kartu bisa diklik di mana saja.
	return container.NewStack(card, container.NewMax(btnTap))
}

// openQuickAddQuest — dialog pembuatan QUEST baru. Jika preferCat kosong,
// user memilih kategori lebih dulu; jika terisi (dari halaman modul), dialog
// langsung ke form dan sub-modul dibatasi pada kategori itu → memenuhi
// requirement "bisa menambah quest di setiap kategorinya".
func openQuickAddQuest(sh *Shell, preferCat model.CategoryKey) {
	win := sh.Win
	showForm := func(catKey model.CategoryKey) {
		var subs []model.SubModule
		sh.App.DB.Where("category_key = ? AND is_active = 1", catKey).
			Order("sort_order").Find(&subs)
		subNames := make([]string, len(subs))
		for i, s := range subs {
			subNames[i] = s.Name
		}
		if len(subNames) == 0 {
			subNames = []string{"(tanpa sub-modul)"}
		}

		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Nama quest, mis. \"Kursus Go Concurrency\"")
		descEntry := widget.NewMultiLineEntry()
		descEntry.SetPlaceHolder("Definition of done / catatan (opsional)")
		subSel := widget.NewSelect(subNames, nil)
		subSel.SetSelected(subNames[0])
		estEntry := widget.NewEntry()
		estEntry.SetPlaceHolder("Estimasi total menit (mis. 240)")
		priSel := widget.NewSelect([]string{"P1 Kritis", "P2 Tinggi", "P3 Normal", "P4 Santai"}, nil)
		priSel.SetSelected("P3 Normal")

		form := widget.NewForm(
			widget.NewFormItem("Judul Quest", nameEntry),
			widget.NewFormItem("Sub-Modul", subSel),
			widget.NewFormItem("Prioritas", priSel),
			widget.NewFormItem("Estimasi (menit)", estEntry),
			widget.NewFormItem("Deskripsi", descEntry),
		)

		dlg := dialog.NewForm("➕ Quest Baru — kategori "+catName(catKey), "Simpan", "Batal", form, win)
		dlg.Resize(fyne.NewSize(minI(win.Canvas().Size().Width-40, 520), 480))
		dlg.OnSubmitted = func(_ fyne.CanvasObject) {
			p, _ := strconv.Atoi(priSel.Selected[:2][1:])
			est, _ := strconv.Atoi(estEntry.Text)
			proj := model.Project{
				SubModuleID:       selectedSubID(subs, subSel.Selected),
				CategoryKey:       catKey,
				Name:              nameEntry.Text,
				Description:       descEntry.Text,
				Status:            model.ProjPlanned,
				Priority:          p,
				EstimatedTotalMin: est,
			}
			proj.FolderTemplate = model.DefaultFolderTemplate
			if err := sh.App.DB.Create(&proj).Error; err != nil {
				dialog.ShowError(err, win)
				return
			}
			// Auto-generate 5 fase modular (Riset→Desain→Impl→Test→Dokumentasi)
			for k, phase := range model.DefaultSubTaskPhases {
				t := model.SubTask{ProjectID: proj.ID, Name: phase, PhaseOrder: k + 1,
					Status: model.TaskTodo, Complexity: model.WeightMedium}
				sh.App.DB.Create(&t)
			}
			sh.Nav.Reload()
			dialog.ShowInformation("Tersimpan",
				"Quest \""+proj.Name+"\" dibuat lengkap dengan 5 fase kerjanya.", win)
		}
		dlg.Show()
	}

	if preferCat != "" {
		showForm(preferCat)
		return
	}
	// Pilih kategori dulu — daftar berwarna sesuai identitas modul.
	items := make([]string, 0, len(mtheme.OrderedPalette))
	keys := make([]model.CategoryKey, 0, len(mtheme.OrderedPalette))
	for _, cc := range mtheme.OrderedPalette {
		items = append(items, cc.Icon+"  "+cc.Name)
		keys = append(keys, cc.Key)
	}
	sel := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(items[i])
		})
	prompt := dialog.NewCustom("Pilih Kategori Quest", "Batal",
		container.NewCenter(sel), win)
	sel.OnSelected = func(id widget.ListItemID) {
		prompt.Hide()
		showForm(keys[id])
	}
	prompt.Show()
}

func selectedSubID(subs []model.SubModule, name string) uint {
	for _, s := range subs {
		if s.Name == name {
			return s.ID
		}
	}
	return 0
}

func catName(k model.CategoryKey) string {
	if cc, ok := mtheme.Palette[k]; ok {
		return cc.Name
	}
	return string(k)
}

func minI(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
