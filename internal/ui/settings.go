// ============================================================================
// LAYAR: Pengaturan (BAGIAN 4 no.6) — toggle Mode Terang/Gelap, dll.
// ============================================================================
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	mtheme "mastermap/internal/theme"
	"mastermap/internal/ui/components"
)

// BuildSettings — layar pengaturan; switch tema adalah fitur wajib sesuai
// BAGIAN 3 ("Dark Mode WAJIB"). Preferensi disimpan ke DB via state.App.
func BuildSettings(sh *Shell) fyne.CanvasObject {
	dark := *sh.App.DarkMode

	header := HeaderBar("Pengaturan", sh.Nav.Pop)

	swTheme := widget.NewCheck("Mode Gelap (Dark Mode)", func(on bool) {
		if on != *sh.App.DarkMode {
			sh.App.ToggleDarkMode(func() { sh.Nav.Reload() })
		}
	})
	swTheme.SetChecked(dark)

	notif := widget.NewCheck("Notifikasi pengingat blok waktu", nil)
	notif.SetChecked(sh.App.GetSetting("notif_enabled") == "true")
	notif.OnChanged = func(v bool) {
		b := "false"
		if v {
			b = "true"
		}
		_ = sh.App.SetSetting("notif_enabled", b)
	}

	backup := widget.NewButton("Cadangkan Database (ekspor .db)", func() {
		// Implementasi nyata di LANGKAH 7 (ekspor file SQLite).
		showToast(sh, "Backup akan tersedia setelah build APK (LANGKAH 7).")
	})

	card := components.NewCard(sh.App.DarkMode,
		func(d bool) color.NRGBA { return mtheme.SurfaceColor(d) },
		func(d bool) color.NRGBA { return mtheme.ForegroundColor(d) },
		container.NewVBox(
			swTheme, components.Spacer(10), notif, components.Spacer(14), backup))

	body := container.NewVScroll(container.NewPadded(card))
	return PageScaffold(header, body, nil)
}

// showToast — konfirmasi singkat ala snackbar Material.
func showToast(sh *Shell, msg string) {
	dialog.ShowInformation("Master Map", msg, sh.Win)
}
