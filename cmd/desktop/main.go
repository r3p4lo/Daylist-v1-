// ============================================================================
// ENTRY POINT DESKTOP (dev preview) — bootstrap Fyne app + theme engine.
// Di Android nanti file ini jadi referensi App struct untuk gomobile bind.
// ============================================================================
package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"mastermap/internal/db"
	"mastermap/internal/state"
	mtheme "mastermap/internal/theme"
	"mastermap/internal/ui"
)

func main() {
	// 1. Database: buka + migrasi + seed Master Map (idempotent).
	gdb, err := db.Open("mastermap.db")
	if err != nil {
		log.Fatalf("gagal membuka database: %v", err)
	}

	// 2. State global: flag dark mode dipakai bersama Theme & semua widget.
	dark := true // default gelap (guideline BAGIAN 3)
	a := &state.App{DB: gdb, DarkMode: &dark}
	a.LoadDarkMode()

	// 3. Fyne app + custom theme engine (live switch via *bool).
	appn := app.NewWithID("com.mastermap.lifeplanner")
	appn.Settings().SetTheme(mtheme.For(&dark))
	w := appn.NewWindow("Master Map Life Planner")
	w.Resize(fyne.NewSize(420, 780)) // simulasi layar HP

	// 4. Navigator + layar pertama: Master Map (home).
	reg := ui.NewRegistry()
	nav := ui.NewNavigator(w, &dark, reg)
	sh := &ui.Shell{App: a, Nav: nav, Win: w}
	nav.Push("master", "Master Map", func() fyne.CanvasObject { return ui.BuildMasterMap(sh) })

	// 5. Ticker animasi transisi (~60fps) — murni UI thread, aman.
	done := make(chan struct{})
	tk := time.NewTicker(16 * time.Millisecond)
	go func() {
		for {
			select {
			case <-tk.C:
				_ = nav.TickAnim()
			case <-done:
				return
			}
		}
	}()

	w.ShowAndRun()
	close(done)
}
