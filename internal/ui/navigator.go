// ============================================================================
// PACKAGE: ui — Navigator: stack layar + transisi halus antar halaman
// ----------------------------------------------------------------------------
// Peta navigasi (BAGIAN 4):
//
// MasterMap ──▶ ModulDetail(category) ──▶ ProjectDetail(project)
// Settings & Timeline via toolbar Master Map
//
// Transisi: slide horizontal ±210ms easeOutQuad (selaras theme.AnimDuration),
// terasa seperti push/pop native Android.
//
// CATATAN live dark-mode: widget standar Fyne membaca warna lewat AppTheme
// yang memegang *bool yang sama → otomatis ikut. Teks kustom (canvas.Text)
// tidak punya hook refresh, jadi tiap layar WAJIB mendaftar ke Registry agar
// dibangun ulang saat user men-toggle Mode Terang/Gelap.
// ============================================================================
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	mtheme "mastermap/internal/theme"
)

// ScreenBuilder membangun konten sebuah layar (dipanggil ulang saat ganti tema).
type ScreenBuilder func() fyne.CanvasObject

// Registry — menyimpan builder layar aktif untuk rebuild saat dark toggle.
type Registry struct {
	build   map[string]ScreenBuilder
	lastKey string
}

func NewRegistry() *Registry { return &Registry{build: map[string]ScreenBuilder{}} }

// Register menyimpan builder per kunci layar ("master", "modul:TECHNOLOGY", ...).
func (r *Registry) Register(key string, b ScreenBuilder) { r.build[key] = b; r.lastKey = key }

// RebuildLast membangun ulang layar terakhir dan mengembalikannya.
func (r *Registry) RebuildLast() fyne.CanvasObject {
	if b, ok := r.build[r.lastKey]; ok {
		return b()
	}
	return nil
}

// animJob — satu animasi geser; di-tick dari main-loop AppShell (thread UI).
type animJob struct {
	obj   fyne.CanvasObject
	from  float32
	steps int
	step  int
	done  func()
}

// screenEntry — satu entri pada stack navigasi.
type screenEntry struct {
	key   string
	title string
	b     ScreenBuilder
}

// Navigator — mengelola konten window + animasi transisi.
type Navigator struct {
	win     fyne.Window
	root    *fyne.Container
	bgRect  *canvas.Rectangle // latar global mengikuti tema (live)
	current fyne.CanvasObject
	dark    *bool
	reg     *Registry
	anim    *animJob
	stack   []screenEntry // stack halaman utk navigasi mundur (Pop)
}

// NewNavigator membuat navigator untuk satu window.
func NewNavigator(win fyne.Window, dark *bool, reg *Registry) *Navigator {
	bg := canvas.NewRectangle(mtheme.BackgroundColor(*dark))
	root := container.NewStack(bg)
	n := &Navigator{win: win, root: root, dark: dark, bgRect: bg, reg: reg}
	win.SetContent(root)
	return n
}

// Push memasang layar baru (dengan animasi bila sudah ter-layout).
func (n *Navigator) Push(key, title string, b ScreenBuilder) {
	n.reg.Register(key, b)
	n.stack = append(n.stack, screenEntry{key: key, title: title, b: b})
	content := b()
	if title != "" {
		n.win.SetTitle("Master Map • " + title)
	}
	n.transition(content)
}

// Reload membangun ulang layar aktif tanpa animasi (toggle tema / setelah CRUD).
func (n *Navigator) Reload() {
	if c := n.reg.RebuildLast(); c != nil {
		n.swap(c)
	}
}

func (n *Navigator) transition(next fyne.CanvasObject) {
	size := n.root.Size()
	if size.Width < 10 || size.Height < 10 || n.current == nil {
		n.swap(next)
		return
	}
	next.Move(fyne.NewPos(size.Width, 0))
	next.Resize(size)
	n.root.Add(next)
	n.anim = &animJob{obj: next, from: size.Width, steps: 24, done: func() { n.swap(next) }}
}

// TickAnim dipanggil ticker AppShell (~60fps) — semua mutasi objek terjadi
// di thread UI sehingga aman. Mengembalikan true bila animasi masih berjalan.
func (n *Navigator) TickAnim() bool {
	if n.anim == nil {
		return false
	}
	a := n.anim
	a.step++
	t := float32(a.step) / float32(a.steps)
	ease := 1 - (1-t)*(1-t) // easeOutQuad — halus ala Material
	a.obj.Move(fyne.NewPos(a.from*(1-ease), 0))
	canvas.Refresh(n.root)
	if a.step >= a.steps {
		n.anim = nil
		a.done()
	}
	return n.anim != nil
}

func (n *Navigator) swap(c fyne.CanvasObject) {
	if c == nil {
		return
	}
	n.bgRect.FillColor = mtheme.BackgroundColor(*n.dark)
	n.root.Objects = []fyne.CanvasObject{n.bgRect, c}
	c.Move(fyne.NewPos(0, 0))
	c.Resize(n.root.Size())
	n.current = c
	canvas.Refresh(n.root)
}

// Pop kembali ke layar sebelumnya dengan animasi geser terbalik.
func (n *Navigator) Pop() {
	if len(n.stack) == 0 {
		return
	}
	last := n.stack[len(n.stack)-1]
	n.stack = n.stack[:len(n.stack)-1]
	prev := n.stack[len(n.stack)-1]
	n.reg.Register(prev.key, prev.b)
	content := prev.b()
	n.win.SetTitle("Master Map • " + prev.title)
	size := n.root.Size()
	if size.Width < 10 || n.current == nil {
		n.swap(content)
		return
	}
	// Geser masuk dari kiri — terasa seperti "mundur" dalam stack.
	content.Move(fyne.NewPos(-size.Width, 0))
	content.Resize(size)
	n.root.Add(content)
	n.anim = &animJob{obj: content, from: -size.Width, steps: 24, done: func() { n.swap(content) }}
	_ = last
}

// HeaderBar — app bar material: tombol kembali + judul + aksi kanan.
func HeaderBar(title string, onBack func(), actions ...fyne.CanvasObject) fyne.CanvasObject {
	left := []fyne.CanvasObject{}
	if onBack != nil {
		btn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), onBack)
		btn.Importance = widget.LowImportance
		left = append(left, btn)
	}
	lbl := widget.NewLabelWithStyle(title, fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true})
	left = append(left, lbl)
	right := container.NewHBox(actions...)
	bar := container.NewBorder(nil, nil, container.NewHBox(left...), right)
	return container.NewPadded(bar)
}

// PageScaffold — kerangka standar layar: header + konten scrollable + FAB.
func PageScaffold(header, body, fab fyne.CanvasObject) fyne.CanvasObject {
	if fab != nil {
		return container.NewBorder(header, nil, nil, fab,
			container.NewPadded(body))
	}
	return container.NewBorder(header, nil, nil, nil,
		container.NewPadded(body))
}

// Spacer — tinggi tetap untuk ritme vertikal antar kartu (grid 4-pt).
func Spacer(h float32) fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.Resize(fyne.NewSize(1, h))
	r.SetMinSize(fyne.NewSize(1, h))
	return r
}
