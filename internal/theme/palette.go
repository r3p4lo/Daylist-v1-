// ============================================================================
// PACKAGE: theme — Color Palette Master Map (sumber kebenaran warna tunggal)
// ----------------------------------------------------------------------------
// Prinsip UX: color-coding HARUS konsisten di seluruh aplikasi (kartu Master
// Map, chip sub-modul, timeline, timer). Karena itu semua warna didefinisikan
// SEKALI di sini dan dibaca oleh theme engine, komponen UI, maupun repository.
//
// Setiap kategori punya 2 varian:
//   - Light : warna identitas asli (sesuai BAGIAN 1 spesifikasi)
//   - Dark  : TINT yang lebih terang agar kontras >= 4.5:1 di atas background
//     gelap #121212 (Material Dark Theme guideline).
//
// ============================================================================
package theme

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"mastermap/internal/model"
)

// CatColor — set warna lengkap untuk satu kategori.
type CatColor struct {
	Key       model.CategoryKey
	Name      string
	Icon      string
	Base      color.NRGBA // warna identitas (light mode) — dipakai aksen/teks
	Dark      color.NRGBA // varian tint untuk dark mode
	Fill      color.NRGBA // solid: header kartu / ring progres
	OnFill    color.NRGBA // warna teks/icon DI ATAS Fill (kontras AA dihitung)
	ChipLight color.NRGBA // background chip subtle (light mode)
	ChipDark  color.NRGBA // background chip subtle (dark mode)
	TintLight color.NRGBA // latar kartu lembut (light mode)
	TintDark  color.NRGBA // latar kartu lembut (dark mode)
}

// parseHex mengonversi "#RRGGBB" → NRGBA dengan alpha a.
func parseHex(hex string, a uint8) color.NRGBA {
	digits := make([]uint8, 0, 6)
	for _, ch := range hex {
		switch {
		case ch >= '0' && ch <= '9':
			digits = append(digits, uint8(ch-'0'))
		case ch >= 'a' && ch <= 'f':
			digits = append(digits, uint8(ch-'a'+10))
		case ch >= 'A' && ch <= 'F':
			digits = append(digits, uint8(ch-'A'+10))
		}
	}
	if len(digits) != 6 {
		return color.NRGBA{R: 0x9E, G: 0x9E, A: a, B: 0x9E} // fallback abu-abu
	}
	return color.NRGBA{
		R: digits[0]<<4 | digits[1],
		G: digits[2]<<4 | digits[3],
		B: digits[4]<<4 | digits[5],
		A: a,
	}
}

// -----------------------------------------------------------------------------
// Fungsi turunan warna (dipakai otomatis saat membangun palette, JANGAN hardcode
// hitam/putih manual) — menjamin kontras teks selalu memenuhi WCAG AA.
// -----------------------------------------------------------------------------

// luminance — kecerahan relatif WCAG (0..1).
func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	toLin := func(v float64) float64 {
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	lr, lg, lb := toLin(float64(r)/65535), toLin(float64(g)/65535), toLin(float64(b)/65535)
	return 0.2126*lr + 0.7152*lg + 0.0722*lb
}

// contrastRatio — rasio kontras WCAG antara dua warna.
func contrastRatio(a, b color.Color) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// readableOn — pilih putih atau hitam (#111) sebagai warna teks di atas bg,
// berdasarkan rasio kontras mana yang lebih baik (>= 4.5:1 prioritas).
func readableOn(bg color.Color) color.NRGBA {
	white := color.NRGBA{255, 255, 255, 255}
	black := color.NRGBA{17, 17, 17, 255}
	if contrastRatio(white, bg) >= contrastRatio(black, bg) {
		return white
	}
	return black
}

// mix — mencampur warna fg ke bg dengan proporsi t (0..1). Untuk tint/chip.
func mix(bg, fg color.NRGBA, t float64) color.NRGBA {
	cl := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return color.NRGBA{
		R: cl(float64(bg.R)*(1-t) + float64(fg.R)*t),
		G: cl(float64(bg.G)*(1-t) + float64(fg.G)*t),
		B: cl(float64(bg.B)*(1-t) + float64(fg.B)*t),
		A: 255,
	}
}

// -----------------------------------------------------------------------------
// PALETTE — 6 kategori sesuai spesifikasi BAGIAN 1.
// Base/Dark diambil dari kolom ColorHex/ColorDarkHex seed database sehingga
// DB dan UI tidak pernah bisa "berbeda warna".
// -----------------------------------------------------------------------------
var Palette = map[model.CategoryKey]*CatColor{}

// OrderedPalette — urutan tampil di Master Map.
var OrderedPalette []*CatColor

func init() {
	lightBG := color.NRGBA{0xFA, 0xFA, 0xFA, 255} // surface light
	darkBG := color.NRGBA{0x1E, 0x1E, 0x1E, 255}  // surface dark (card)

	for _, seed := range model.SeedCategories {
		base := parseHex(seed.ColorHex, 255)
		dark := parseHex(seed.ColorDarkHex, 255)
		cc := &CatColor{
			Key:       seed.Key,
			Name:      seed.Name,
			Icon:      seed.Icon,
			Base:      base,
			Dark:      dark,
			Fill:      base,
			OnFill:    readableOn(base),
			ChipLight: mix(lightBG, base, 0.14),
			ChipDark:  mix(darkBG, dark, 0.22),
			TintLight: mix(lightBG, base, 0.08),
			TintDark:  mix(darkBG, dark, 0.14),
		}
		Palette[seed.Key] = cc
	}
	for _, k := range model.AllCategories {
		OrderedPalette = append(OrderedPalette, Palette[k])
	}
}

// Accent — warna aksen kategori yang aktif terhadap tema saat ini.
func (c *CatColor) Accent(dark bool) color.NRGBA {
	if dark {
		return c.Dark
	}
	return c.Base
}

// ChipBg — background chip subtle sesuai mode.
func (c *CatColor) ChipBg(dark bool) color.NRGBA {
	if dark {
		return c.ChipDark
	}
	return c.ChipLight
}

// CardTint — latar kartu lembut sesuai mode.
func (c *CatColor) CardTint(dark bool) color.NRGBA {
	if dark {
		return c.TintDark
	}
	return c.TintLight
}

// TextOnFill alias agar pemanggil komponen jelas maksudnya.
func (c *CatColor) TextOnFill() color.NRGBA { return c.OnFill }

// NewText — helper membuat canvas.Text berwarna (dipakai banyak komponen).
func NewText(str string, size float32, col color.Color) *canvas.Text {
	t := canvas.NewText(str, col)
	t.TextSize = size
	t.Alignment = fyne.TextAlignLeading
	return t
}
