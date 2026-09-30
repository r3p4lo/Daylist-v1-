// ============================================================================
// PACKAGE: engine — Dynamic Time-Boxing Engine (CORE LOGIC / LANGKAH 5)
// ----------------------------------------------------------------------------
// Prinsip: durasi per kategori BUKAN angka tetap selamanya. Engine memilih
// panjang blok berdasarkan "Bobot Kognitif" (CognitiveWeight) lalu memecah
// estimasi total tugas menjadi rantai blok + jeda istirahat.
//
// Contoh: Riset 2 jam (bobot BERAT) -> [50m, 10m break, 50m, 10m break, 20m]
//
// Input : bobot kognitif + estimasi total menit (+ kompleksitas 0..1)
// Output: []Block siap dipersist jadi model.TimeBlock (Source="ENGINE")
//         atau di-override manual user (Source="USER_OVERRIDE").
// ============================================================================
package engine

import (
"math"

"mastermap/internal/model"
)

// Block — hasil rekomendasi engine untuk satu segmen waktu.
type Block struct {
Type        model.BlockType // A (ringan) / B (deep work)
DurationMin int             // durasi dalam menit
IsBreak     bool            // true = jeda istirahat antar blok
}

// Params — masukan algoritma pemecahan waktu.
type Params struct {
Weight     model.CognitiveWeight // bobot kognitif sub-modul/kategori
TotalMin   int                   // estimasi total durasi tugas (menit)
Complexity float64               // 0.0..1.0 — penggeser intra-rentang
}

// Rentang dinamis per bobot kognitif — sesuai aturan BAGIAN 2:
//
//Ringan (WeightLight)  -> Blok A : 15-30 menit (admin/budgeting/email/olahraga)
//Sedang (WeightMedium) -> Blok B : 45-60 menit (bahasa/kursus/kreatif/desain)
//Berat  (WeightHeavy)  -> Blok B : 50-90 menit (riset berat/AI/Web3/arsitektur)
type span struct {
typ      model.BlockType
min, max int
breakMin int // panjang jeda setelah blok penuh
}

func spanFor(w model.CognitiveWeight) span {
switch w {
case model.WeightHeavy:
return span{model.BlockB, 50, 90, model.BreakLongMinutes} // 10m
case model.WeightMedium:
return span{model.BlockB, 45, 60, model.BreakShortMinutes} // 5m
default: // WeightLight
return span{model.BlockA, 15, 30, model.BreakShortMinutes} // 5m
}
}

// RecommendDuration — panjang SATU blok ideal untuk bobot tertentu.
// Mulai dari preset dasar (A=30, B=50), digeser ke arah maksimum seiring
// naiknya kompleksitas; dibulatkan ke kelipatan 5 menit agar mudah dibaca.
func RecommendDuration(w model.CognitiveWeight, complexity float64) int {
s := spanFor(w)
base := float64(s.min)
if w == model.WeightLight {
base = math.Min(float64(model.PresetBlockADuration), float64(s.max))
} else if w == model.WeightMedium {
base = math.Min(float64(model.PresetBlockBDuration), float64(s.max))
}
if complexity < 0 {
complexity = 0
}
if complexity > 1 {
complexity = 1
}
room := float64(s.max) - base
d := base + room*complexity*0.8
rounded := math.Round(d/5) * 5
if rounded < float64(s.min) {
rounded = float64(s.min)
}
if rounded > float64(s.max) {
rounded = float64(s.max)
}
return int(rounded)
}

// Plan — ALGORITMA UTAMA: pecah TotalMin menjadi rantai blok + jeda.
//  1. Panjang blok = RecommendDuration(bobot, kompleksitas).
//  2. Setelah tiap blok PENUH disisipkan jeda sesuai bobot.
//  3. Sisa (< panjang blok) jadi blok penutup tanpa jeda setelahnya.
//  4. Fragmen < 10 menit di ujung dilebur ke blok sebelumnya.
func Plan(p Params) []Block {
if p.TotalMin <= 0 {
return nil
}
s := spanFor(p.Weight)
unit := RecommendDuration(p.Weight, p.Complexity)

var out []Block
remaining := p.TotalMin
for remaining > 0 {
take := unit
if remaining < take {
take = remaining
}
if remaining-take > 0 && remaining-take < 10 && len(out) > 0 {
out[len(out)-1].DurationMin += remaining - take
return out
}
out = append(out, Block{Type: s.typ, DurationMin: take})
remaining -= take
if remaining > 0 {
out = append(out, Block{Type: s.typ, DurationMin: s.breakMin, IsBreak: true})
}
}
return out
}

// PlanWithBudget — versi sadar-energi: jika total blok kerja melampaui
// budget fokus harian (DailyPlan.EnergyBudgetMin), sisa tugas dikembalikan
// sebagai overflowMin (belum dijadwalkan hari ini). Mewujudkan aturan:
// "time-boxing mencegah satu kategori menyedot semua energi harian".
func PlanWithBudget(p Params, energyBudgetMin int) (blocks []Block, overflowMin int) {
blocks = Plan(p)
if energyBudgetMin <= 0 {
return blocks, 0
}
used, keep := 0, 0
for i, b := range blocks {
if b.IsBreak {
keep = i + 1
continue
}
if used+b.DurationMin > energyBudgetMin {
return blocks[:keep], p.TotalMin - used
}
used += b.DurationMin
keep = i + 1
}
return blocks, 0
}

// ToTimeBlocks — konversi hasil engine menjadi baris model.TimeBlock siap
// insert. source: "ENGINE" (rekomendasi diterima) | "USER_OVERRIDE".
func ToTimeBlocks(blocks []Block, dailyPlanID uint, cat model.CategoryKey, title string, subTaskID uint, source string) []model.TimeBlock {
out := make([]model.TimeBlock, 0, len(blocks))
seq := 0
for _, b := range blocks {
seq++
out = append(out, model.TimeBlock{
DailyPlanID: dailyPlanID,
SubTaskID:   subTaskID,
CategoryKey: cat,
BlockType:   b.Type,
Title:       title,
DurationMin: b.DurationMin,
SequenceNo:  seq,
IsBreak:     b.IsBreak,
Source:      source,
Status:      model.BlkPlanned,
})
}
return out
}
