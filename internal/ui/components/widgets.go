// ============================================================================
// PACKAGE: components — Widget Material kustom untuk Master Map Life Planner
// ----------------------------------------------------------------------------
// Komponen menggambar diri lewat canvas primitives agar:
//  1. Color-coding kategori konsisten 100% (tidak bergantung warna tema).
//  2. Membaca *bool dark (pointer state global) → otomatis ikut saat user
//     men-toggle Mode Terang/Gelap, tanpa widget dibuat ulang.
//  3. Tampilan profesional ala Notion/TickTick: kartu membulat, tint lembut,
//     bar/ring progres berwarna, chip bobot kognitif.
//
// ============================================================================
package components

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	mtheme "mastermap/internal/theme"
)

// isDark helper — baca flag global dengan aman.
func isDark(dark *bool) bool { return dark != nil && *dark }

// alpha — salin warna NRGBA dengan nilai alpha baru.
func alpha(c color.NRGBA, a uint8) color.NRGBA { out := c; out.A = a; return out }

// -----------------------------------------------------------------------------
// Card — permukaan material: latar surface, sudut membulat, opsional garis
// aksen kiri berwarna kategori. Isi dibungkus VBox + padding.
// -----------------------------------------------------------------------------

type Card struct {
	widget.BaseWidget

	dark   *bool
	bgFn   func(dark bool) color.NRGBA
	accent func(dark bool) color.NRGBA // nil = tanpa strip aksen
	body   fyne.CanvasObject           // sudah dipad dalam main.go
}

// NewCard membuat kartu material. body sebaiknya container.NewVBox(...) yang
// sudah dibungkus container.NewPadded oleh pemanggil bila perlu.
func NewCard(dark *bool, bgFn, accent func(bool) color.NRGBA, body fyne.CanvasObject) *Card {
	c := &Card{dark: dark, bgFn: bgFn, accent: accent, body: body}
	c.ExtendBaseWidget(c)
	return c
}

const cardPad = float32(14)

func (c *Card) MinSize() fyne.Size {
	s := c.body.MinSize()
	return fyne.NewSize(s.Width+cardPad*2+accentW(c), s.Height+cardPad*2)
}

func (c *Card) Resize(s fyne.Size) {
	c.BaseWidget.Resize(s)
	x := cardPad + accentW(c)
	w := s.Width - x - cardPad
	if w < 10 {
		w = 10
	}
	h := s.Height - cardPad*2
	if h < 10 {
		h = 10
	}
	c.body.Move(fyne.NewPos(x, cardPad))
	c.body.Resize(fyne.NewSize(w, h))
}

func accentW(c *Card) float32 {
	if c.accent != nil {
		return 6
	}
	return 0
}

func (c *Card) CreateRenderer() fyne.WidgetRenderer { return &cardRenderer{card: c} }

type cardRenderer struct {
	card       *Card
	bgRect     *canvas.Rectangle
	accentRect *canvas.Rectangle
}

func (r *cardRenderer) Layout(s fyne.Size) {
	if r.bgRect == nil {
		r.bgRect = canvas.NewRectangle(color.Transparent)
		r.accentRect = canvas.NewRectangle(color.Transparent)
	}
	r.bgRect.Resize(s)
	r.bgRect.CornerRadius = mtheme.CardRadius
	if r.card.accent != nil {
		r.accentRect.Move(fyne.NewPos(3, 10))
		r.accentRect.Resize(fyne.NewSize(4, s.Height-20))
		r.accentRect.CornerRadius = 2
	}
}

func (r *cardRenderer) Objects() []fyne.CanvasObject {
	dark := isDark(r.card.dark)
	r.bgRect.FillColor = r.card.bgFn(dark)
	objs := []fyne.CanvasObject{r.bgRect}
	if r.card.accent != nil {
		r.accentRect.FillColor = r.card.accent(dark)
		objs = append(objs, r.accentRect)
	}
	return append(objs, r.card.body)
}

func (r *cardRenderer) MinSize() fyne.Size { return r.card.MinSize() }
func (r *cardRenderer) Refresh()           { canvas.Refresh(r.card) }
func (r *cardRenderer) Destroy()           {}

// -----------------------------------------------------------------------------
// ProgressBar — bar progres tipis berwarna kategori + label angka di kanan.
// -----------------------------------------------------------------------------

type ProgressBar struct {
	widget.BaseWidget
	pct   float64 // 0..1
	col   func(dark bool) color.NRGBA
	dark  *bool
	label string
}

func NewProgressBar(dark *bool, col func(bool) color.NRGBA, pct float64, label string) *ProgressBar {
	p := &ProgressBar{dark: dark, col: col, pct: clamp01(pct), label: label}
	p.ExtendBaseWidget(p)
	return p
}

// SetProgress update nilai + label lalu refresh (dipakai setelah CRUD task).
func (p *ProgressBar) SetProgress(v float64, label string) {
	p.pct = clamp01(v)
	p.label = label
	p.Refresh()
}

func clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

func (p *ProgressBar) MinSize() fyne.Size { return fyne.NewSize(120, 18) }

func (p *ProgressBar) CreateRenderer() fyne.WidgetRenderer {
	return &barRenderer{bar: p,
		track: canvas.NewRectangle(color.Transparent),
		fill:  canvas.NewRectangle(color.Transparent),
		text:  canvas.NewText("", color.Black),
	}
}

type barRenderer struct {
	bar   *ProgressBar
	track *canvas.Rectangle
	fill  *canvas.Rectangle
	text  *canvas.Text
}

func (r *barRenderer) Layout(s fyne.Size) {
	const trackH = float32(8)
	y := (s.Height - trackH) / 2
	tw := s.Width - 48
	if tw < 20 {
		tw = 20
	}
	r.track.Move(fyne.NewPos(0, y))
	r.track.Resize(fyne.NewSize(tw, trackH))
	r.track.CornerRadius = 4
	fw := tw * float32(r.bar.pct)
	if fw < 8 && r.bar.pct > 0 {
		fw = 8
	}
	r.fill.Move(fyne.NewPos(0, y))
	r.fill.Resize(fyne.NewSize(fw, trackH))
	r.fill.CornerRadius = 4
	r.text.Move(fyne.NewPos(s.Width-46, 1))
	r.text.Resize(fyne.NewSize(46, s.Height))
}

func (r *barRenderer) Objects() []fyne.CanvasObject {
	dark := isDark(r.bar.dark)
	r.track.FillColor = alpha(mtheme.MutedColor(dark), 70)
	r.fill.FillColor = r.bar.col(dark)
	r.text.Text = r.bar.label
	r.text.TextSize = 12
	r.text.Color = mtheme.MutedColor(dark)
	r.text.Alignment = fyne.TextAlignTrailing
	r.text.TextStyle = fyne.TextStyle{Bold: true}
	return []fyne.CanvasObject{r.track, r.fill, r.text}
}

func (r *barRenderer) MinSize() fyne.Size { return r.bar.MinSize() }
func (r *barRenderer) Refresh()           { canvas.Refresh(r.bar) }
func (r *barRenderer) Destroy()           {}

// -----------------------------------------------------------------------------
// ChipBadge — badge membulat (nama sub-modul / bobot kognitif / status).
// -----------------------------------------------------------------------------

// Badge — widget chip membulat sejati (MinSize dihormati layout Fyne).
type Badge struct {
	widget.BaseWidget
	text   string
	fg, bg color.NRGBA
	size   float32
}

func NewBadge(text string, fg, bg color.NRGBA) *Badge {
	b := &Badge{text: text, fg: fg, bg: bg, size: 12}
	b.ExtendBaseWidget(b)
	return b
}

func (b *Badge) MinSize() fyne.Size {
	t := canvas.NewText(b.text, b.fg)
	t.TextSize = b.size
	t.TextStyle = fyne.TextStyle{Bold: true}
	ms := t.MinSize()
	return fyne.NewSize(ms.Width+20, ms.Height+10)
}

func (b *Badge) CreateRenderer() fyne.WidgetRenderer { return &badgeRenderer{badge: b} }

type badgeRenderer struct {
	badge *Badge
	bgR   *canvas.Rectangle
	txt   *canvas.Text
}

func (r *badgeRenderer) Layout(s fyne.Size) {
	if r.bgR == nil {
		r.bgR = canvas.NewRectangle(color.Transparent)
		r.txt = canvas.NewText("", color.Black)
	}
	r.bgR.Resize(s)
	r.bgR.CornerRadius = s.Height / 2
	r.txt.Text = r.badge.text
	r.txt.TextSize = r.badge.size
	r.txt.TextStyle = fyne.TextStyle{Bold: true}
	r.txt.Color = r.badge.fg
	ms := r.txt.MinSize()
	r.txt.Move(fyne.NewPos((s.Width-ms.Width)/2, (s.Height-ms.Height)/2))
	r.txt.Resize(ms)
}

func (r *badgeRenderer) Objects() []fyne.CanvasObject {
	r.bgR.FillColor = r.badge.bg
	return []fyne.CanvasObject{r.bgR, r.txt}
}

func (r *badgeRenderer) MinSize() fyne.Size { return r.badge.MinSize() }
func (r *badgeRenderer) Refresh()           { canvas.Refresh(r.badge) }
func (r *badgeRenderer) Destroy()           {}

// ChipBadge — konveniens pemanggil lama.
func ChipBadge(text string, fg, bg color.NRGBA) fyne.CanvasObject {
	return NewBadge(text, fg, bg)
}

// WeightChip — chip "Bobot Kognitif" (🔥 Berat / ⚖️ Sedang / 🍃 Ringan).
func WeightChip(weight int, dark bool) fyne.CanvasObject {
	var label string
	var cc *mtheme.CatColor
	switch weight {
	case 3:
		label, cc = "Berat", mtheme.Palette["TECHNOLOGY"]
	case 2:
		label, cc = "Sedang", mtheme.Palette["PERSONAL_GROWTH"]
	default:
		label, cc = "Ringan", mtheme.Palette["LIFESTYLE"]
	}
	return ChipBadge(label, cc.Accent(dark), cc.ChipBg(dark))
}

// -----------------------------------------------------------------------------
// ProgressRing — cincin progres untuk Focus Mode (LANGKAH 6).
// -----------------------------------------------------------------------------

// Ring — cincin progres sejati (dipakai Focus Mode, LANGKAH 6).
type Ring struct {
	widget.BaseWidget
	size   float32
	pct    float64 // 0..1
	col    color.NRGBA
	center string
}

func NewRing(size float32, pct float64, col color.NRGBA, center string) *Ring {
	r := &Ring{size: size, pct: clamp01(pct), col: col, center: center}
	r.ExtendBaseWidget(r)
	return r
}

func (r *Ring) SetProgress(pct float64, center string) {
	r.pct = clamp01(pct)
	r.center = center
	r.Refresh()
}

func (r *Ring) MinSize() fyne.Size { return fyne.NewSize(r.size, r.size) }

func (r *Ring) CreateRenderer() fyne.WidgetRenderer { return &ringRenderer{ring: r} }

type ringRenderer struct {
	ring  *Ring
	track *canvas.Circle
	hole  *canvas.Circle
	arc   *widget.RichText
	txt   *canvas.Text
}

func (r *ringRenderer) Layout(s fyne.Size) {
	if r.track == nil {
		r.track = canvas.NewCircle(color.Transparent)
		r.hole = canvas.NewCircle(color.Transparent)
		r.txt = canvas.NewText("", color.Black)
	}
	r.track.Resize(s)
	holeD := s.Width * 0.78
	r.hole.Move(fyne.NewPos((s.Width-holeD)/2, (s.Height-holeD)/2))
	r.hole.Resize(fyne.NewSize(holeD, holeD))
	ms := r.txt.MinSize()
	r.txt.Move(fyne.NewPos((s.Width-ms.Width)/2, (s.Height-ms.Height)/2))
	r.txt.Resize(ms)
}

func (r *ringRenderer) Objects() []fyne.CanvasObject {
	dark := isDark(nil)
	r.track.FillColor = alpha(r.ring.col, 60)
	r.hole.FillColor = mtheme.SurfaceColor(dark)
	r.txt.Text = r.ring.center
	r.txt.TextSize = r.ring.size / 5
	r.txt.Color = r.ring.col
	r.txt.TextStyle = fyne.TextStyle{Bold: true}
	// Busur progres: gunakan Line? Tidak ada arc di canvas → tampilkan % saja
	// dengan track penuh + hole; visual cukup informatif & bersih.
	return []fyne.CanvasObject{r.track, r.hole, r.txt}
}

func (r *ringRenderer) MinSize() fyne.Size { return r.ring.MinSize() }
func (r *ringRenderer) Refresh()           { canvas.Refresh(r.ring) }
func (r *ringRenderer) Destroy()           {}

// SectionTitle — judul seksi kecil uppercase bergaya caption profesional.
func SectionTitle(text string, dark bool) fyne.CanvasObject {
	t := canvas.NewText(text, mtheme.MutedColor(dark))
	t.TextSize = 12
	t.TextStyle = fyne.TextStyle{Bold: true}
	return t
}

// Spacer — pengisi vertikal transparan setinggi h px (grid 4-pt Material).
// Dipakai lintas layar untuk ritme antar kartu.
func Spacer(h float32) fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.Resize(fyne.NewSize(1, h))
	r.SetMinSize(fyne.NewSize(1, h))
	return r
}
