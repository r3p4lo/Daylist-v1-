// ============================================================================
// PACKAGE: theme — Custom Theme Engine "Master Map Life Planner"
// ----------------------------------------------------------------------------
// Implementasi fyne.Theme penuh dengan 2 mode: TERANG & GELAP, mengikuti
// Material Design principles:
//   - card-based layout : warna surface & background terpisah (lihat Surface*)
//   - rounded corners   : SizeNameInputRadius/SelectionRadius diperbesar
//   - proper spacing    : padding mengikuti grid 4-pt
//   - tipografi         : sans-serif bawaan Fyne (bersih, tanpa aset tambahan)
//
// KUNCI DESAIN: field `dark` berupa POINTER → komponen UI cukup membaca
// satu sumber kebenaran (state.App.DarkMode) dan seluruh aplikasi ikut berubah
// saat user men-toggle Mode Terang/Gelap di Settings (LANGKAH 6).
// ============================================================================
package theme

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// -----------------------------------------------------------------------------
// Surface palette global (netral, bukan warna kategori) — dipakai juga oleh
// komponen UI kustom (kartu, chip, timeline) agar tidak menebak warna tema.
// -----------------------------------------------------------------------------
const (
	LightBackground uint32 = 0xFAFAFA // halaman light
	LightSurface    uint32 = 0xFFFFFF // kartu light
	DarkBackground  uint32 = 0x121212 // halaman dark (standar Material)
	DarkSurface     uint32 = 0x1E1E1E // kartu dark
	DarkSurfaceAlt  uint32 = 0x2A2A2A // elemen elevated (input/dialog dark)
)

// AppTheme — implementasi fyne.Theme dua-mode.
type AppTheme struct {
	dark *bool // pointer ke state global → live switch tanpa restart app
}

var _ fyne.Theme = (*AppTheme)(nil)

// NewAppTheme membuat instance theme yang membaca kondisi dark lewat pointer.
func NewAppTheme(dark *bool) *AppTheme { return &AppTheme{dark: dark} }

// For — konveniens untuk main.go & settings screen.
func For(dark *bool) fyne.Theme { return NewAppTheme(dark) }

func (t *AppTheme) isDark() bool { return t.dark != nil && *t.dark }

// Color — palet netral per mode. Warna KATEGORI tidak lewat sini; komponen
// kategori memakai theme.Palette langsung (color-coding eksplisit per BAGIAN 1).
func (t *AppTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	dark := t.isDark()
	switch n {
	case theme.ColorNameBackground:
		if dark {
			return grey(DarkBackground)
		}
		return grey(LightBackground)
	case theme.ColorNameButton, theme.ColorNameMenuBackground,
		theme.ColorNameOverlayBackground, theme.ColorNameInputBackground:
		if dark {
			return grey(DarkSurfaceAlt)
		}
		return grey(LightSurface)
	case theme.ColorNamePrimary, theme.ColorNameHyperlink, theme.ColorNameFocus:
		// Aksen global = Royal Blue Technology — netral & profesional.
		if dark {
			return parseHex("#42A5F5", 255)
		}
		return parseHex("#1565C0", 255)
	case theme.ColorNameForeground:
		if dark {
			return grey(0xEDEDED)
		}
		return grey(0x212121)
	case theme.ColorNameDisabled:
		if dark {
			return grey(0x616161)
		}
		return grey(0xBDBDBD)
	case theme.ColorNameSeparator, theme.ColorNameInputBorder:
		if dark {
			return grey(0x3A3A3A)
		}
		return grey(0xE0E0E0)
	case theme.ColorNameHeaderBackground:
		if dark {
			return grey(DarkSurface)
		}
		return grey(LightSurface)
	case theme.ColorNameHover:
		if dark {
			return grey(0x333333)
		}
		return grey(0xF0F0F0)
	case theme.ColorNamePlaceHolder:
		if dark {
			return grey(0x8A8A8A)
		}
		return grey(0x9E9E9E)
	case theme.ColorNameScrollBar:
		if dark {
			return grey(0x424242)
		}
		return grey(0xCFCFCF)
	default:
		// Nama lain (selection, danger, success, dsb.) → delegasi ke tema
		// bawaan Fyne dengan variant yang benar sesuai mode kita.
		return theme.DefaultTheme().Color(n, variantOf(dark))
	}
}

// Font — font bawaan Fyne (sans-serif clean). Tidak menambah aset font berarti
// APK jauh lebih kecil (relevan untuk LANGKAH 7).
func (t *AppTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(s)
}

// Icon — ikon bawaan Fyne (Material-style, satu bahasa visual di seluruh app).
func (t *AppTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

// Size — tuning Material: radius membulat, spacing grid 4-pt, skala tipografi.
func (t *AppTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNamePadding:
		return 8 // jarak antar elemen (kelipatan 4)
	case theme.SizeNameInnerPadding:
		return 12 // padding dalam kartu/input
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 10 // sudut membulat ala Notion/TickTick
	case theme.SizeNameScrollBar:
		return 5
	case theme.SizeNameScrollBarSmall:
		return 3
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14 // body text standar mobile
	case theme.SizeNameHeadingText:
		return 24 // judul halaman
	case theme.SizeNameSubHeadingText:
		return 18 // judul kartu
	case theme.SizeNameCaptionText:
		return 12 // label kecil / persen progress
	default:
		return theme.DefaultTheme().Size(n)
	}
}

// -----------------------------------------------------------------------------
// Transisi halus antar halaman (BAGIAN 3): durasi animasi slide/fade custom
// terasa "snappy tapi smooth" ala Things 3 / TickTick. Dipakai navigator.go.
// -----------------------------------------------------------------------------
const AnimDuration = 220 * time.Millisecond

// CardRadius — radius sudut kartu Material yang dipakai komponen kustom.
const CardRadius = float32(14)

// -----------------------------------------------------------------------------
// Helper internal
// -----------------------------------------------------------------------------

func grey(hexv uint32) color.NRGBA {
	return color.NRGBA{uint8(hexv >> 16), uint8((hexv >> 8) & 0xFF), uint8(hexv & 0xFF), 255}
}

func variantOf(dark bool) fyne.ThemeVariant {
	if dark {
		return theme.VariantDark
	}
	return theme.VariantLight
}

// SurfaceColor — bantu komponen kustom memilih warna latar kartu sesuai mode.
func SurfaceColor(dark bool) color.NRGBA {
	if dark {
		return grey(DarkSurface)
	}
	return grey(LightSurface)
}

// BackgroundColor — warna halaman sesuai mode.
func BackgroundColor(dark bool) color.NRGBA {
	if dark {
		return grey(DarkBackground)
	}
	return grey(LightBackground)
}

// ForegroundColor — warna teks utama sesuai mode.
func ForegroundColor(dark bool) color.NRGBA {
	if dark {
		return grey(0xEDEDED)
	}
	return grey(0x212121)
}

// MutedColor — warna teks sekunder/caption sesuai mode.
func MutedColor(dark bool) color.NRGBA {
	if dark {
		return grey(0x9E9E9E)
	}
	return grey(0x757575)
}
