// ============================================================================
// HELPERS UI — util kecil yang dipakai lintas layar
// ============================================================================
package ui

import (
	"image/color"

	"fyne.io/fyne/v2/canvas"

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
