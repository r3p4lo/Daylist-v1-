// Verifikasi aturan BAGIAN 2 dengan contoh nyata dari spesifikasi.
package engine

import (
"testing"

"mastermap/internal/model"
)

func TestRisetBeratDuaJam(t *testing.T) {
got := Plan(Params{Weight: model.WeightHeavy, TotalMin: 120})
want := []int{50, 10, 50, 10, 20} // [50m, break, 50m, break, 20m]
if len(got) != len(want) {
t.Fatalf("panjang rantai = %d, ingin %d -> %+v", len(got), len(want), got)
}
for i, w := range want {
if got[i].DurationMin != w {
t.Errorf("blok[%d] = %d, ingin %d", i, got[i].DurationMin, w)
}
}
}

func TestRentangDinamis(t *testing.T) {
cases := []struct {
w          model.CognitiveWeight
cx         float64
lo, hi     int
}{
{model.WeightLight, 0, 15, 30},   // administratif/budgeting
{model.WeightMedium, 0.5, 45, 60}, // belajar bahasa/kursus
{model.WeightHeavy, 1.0, 50, 90},  // AI & Automation / Web3
}
for _, c := range cases {
d := RecommendDuration(c.w, c.cx)
if d < c.lo || d > c.hi {
t.Errorf("bobot %d cx %.1f -> %d menit, di luar [%d,%d]", c.w, c.cx, d, c.lo, c.hi)
}
}
}

func TestBudgetEnergi(t *testing.T) {
// Tugas berat 4 jam tapi budget harian hanya 100 menit fokus.
blocks, overflow := PlanWithBudget(Params{Weight: model.WeightHeavy, TotalMin: 240}, 100)
work := 0
for _, b := range blocks {
if !b.IsBreak {
work += b.DurationMin
}
}
if work > 100 {
t.Errorf("blok kerja %d melewati budget 100", work)
}
if overflow <= 0 {
t.Errorf("harus ada overflow, dapat %d", overflow)
}
}
