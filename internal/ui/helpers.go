// ============================================================================
// HELPERS UI — util kecil yang dipakai lintas layar
// ============================================================================
package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	mtheme "mastermap/internal/theme"
)

// canvasText — teks kustom dengan ukuran eksplisit (memakai helper palette).
func canvasText(s string, size float32, col color.Color) *canvas.Text {
	t := mtheme.NewText(s, size, col)
	return t
}

// greenOf / tint — warna bantu untuk badge status.
func greenOf(dark bool) color.NRGBA {
	if cc := mtheme.Palette["LEARNING"]; cc != nil {
		return cc.Accent(dark)
	}
	return color.NRGBA{R: 0x2E, G: 0x7D, B: 0x32, A: 0xFF}
}

// tint — versi transparan (18%) dari warna untuk background chip.
func tint(c color.NRGBA) color.NRGBA {
	out := c
	out.A = 0x2E
	return out
}

// showToast — "toast" ala Android: pesan singkat non-blocking yang ditempel
// di atas root stack navigator dan hilang sendiri setelah 3 detik.
func showToast(sh *Shell, msg string) {
	dark := *sh.App.DarkMode
	lbl := mtheme.NewText(msg, 14, mtheme.ForegroundColor(dark))
	lbl.Alignment = fyne.TextAlignCenter
	bg := canvas.NewRectangle(mtheme.SurfaceColor(dark))
	bg.CornerRadius = 10
	card := container.NewStack(bg, container.NewPadded(lbl))

	root := sh.Win.Content().(*fyne.Container) // root = container.NewStack(bgRect, ...)
	root.Add(card)
	sz := sh.Win.Canvas().Size()
	card.Move(fyne.NewPos(16, sz.Height-96))
	card.Resize(fyne.NewSize(sz.Width-32, 56))
	canvas.Refresh(root)

	time.AfterFunc(3*time.Second, func() {
		for i, o := range root.Objects {
			if o == card {
				root.Objects = append(root.Objects[:i], root.Objects[i+1:]...)
				break
			}
		}
		canvas.Refresh(root)
	})
}
