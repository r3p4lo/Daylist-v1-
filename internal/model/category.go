// ============================================================================
// PACKAGE: model — Master Map Registry (Peta Induk)
// ----------------------------------------------------------------------------
// 6 kategori adalah TETAP (fixed seed) karena menjadi fondasi color-coding UI,
// bobot kognitif engine, dan stabilitas relasi database.
// Yang DINAMIS (dikelola user lewat CRUD) adalah: SubModule, Project, SubTask.
// ============================================================================
package model

import "time"

// Category — representasi satu dari 6 pilar Master Map.
type Category struct {
	Key          CategoryKey `gorm:"primaryKey;type:text"` // mis. "TECHNOLOGY"
	Name         string      `gorm:"not null"`             // "Technology"
	Icon         string      `gorm:"not null"`             // emoji "⚙️"
	ColorHex     string      `gorm:"not null"`             // "#1565C0" (light mode)
	ColorDarkHex string      `gorm:"not null"`             // varian agar tetap terbaca di dark mode
	Tagline      string      // deskripsi singkat untuk subtitle kartu
	SortOrder    int         `gorm:"default:0"` // urutan tampil di Master Map

	SubModules []SubModule `gorm:"foreignKey:CategoryKey;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Category) TableName() string { return "categories" }

// -----------------------------------------------------------------------------
// SEED MASTER MAP — sumber kebenaran tunggal warna & identitas kategori.
// ColorDarkHex dipilih lebih terang (tint) supaya kontras >= 4.5:1 di atas
// background gelap (#121212) sesuai guideline Material Dark Theme.
// -----------------------------------------------------------------------------
var SeedCategories = []Category{
	{
		Key: CatLearning, Name: "Learning", Icon: "🎓",
		ColorHex: "#2E7D32", ColorDarkHex: "#66BB6A", // Emerald Green
		Tagline: "Riset, Belajar, Kursus & Progress", SortOrder: 1,
	},
	{
		Key: CatTechnology, Name: "Technology", Icon: "⚙️",
		ColorHex: "#1565C0", ColorDarkHex: "#42A5F5", // Royal Blue
		Tagline: "Web3, AI, Development & Tools", SortOrder: 2,
	},
	{
		Key: CatFinance, Name: "Finance", Icon: "💰",
		ColorHex: "#E65100", ColorDarkHex: "#FF9800", // Deep Orange
		Tagline: "Budgeting, Investasi & Dashboard", SortOrder: 3,
	},
	{
		Key: CatCareer, Name: "Career", Icon: "💼",
		ColorHex: "#6A1B9A", ColorDarkHex: "#AB47BC", // Deep Purple
		Tagline: "Persiapan Jepang, Visa & Interview", SortOrder: 4,
	},
	{
		Key: CatLifestyle, Name: "Lifestyle", Icon: "❤️",
		ColorHex: "#C62828", ColorDarkHex: "#EF5350", // Crimson
		Tagline: "Kesehatan, Rutinitas & Hobi", SortOrder: 5,
	},
	{
		Key: CatPersonalGrowth, Name: "Personal Growth", Icon: "🌱",
		ColorHex: "#00695C", ColorDarkHex: "#26A69A", // Teal/Cyan
		Tagline: "Meditasi, Refleksi & Disiplin", SortOrder: 6,
	},
}

// SeedSubModules — sub-modul awal per kategori (sesuai BAGIAN 1).
// User dapat menambah/mengubah lewat Modul Detail View (LANGKAH 4).
var SeedSubModules = map[CategoryKey][]struct {
	Name   string
	Weight CognitiveWeight
}{
	CatLearning: {
		{"Riset", WeightHeavy},
		{"Belajar (JLPT, dll)", WeightMedium},
		{"Kursus Online", WeightMedium},
		{"Catatan & Progress", WeightLight},
	},
	CatTechnology: {
		{"Web3 & Blockchain", WeightHeavy},
		{"AI & Automation", WeightHeavy},
		{"Web/App Development", WeightHeavy},
		{"Tools & Infrastruktur", WeightMedium},
	},
	CatFinance: {
		{"Budgeting", WeightLight},
		{"Investasi", WeightMedium},
		{"Portfolio", WeightMedium},
		{"Financial Dashboard", WeightMedium},
	},
	CatCareer: {
		{"Persiapan Jepang", WeightMedium},
		{"Dokumen & Visa", WeightLight},
		{"Bahasa Jepang", WeightMedium},
		{"Interview & Contract", WeightHeavy},
	},
	CatLifestyle: {
		{"Olahraga & Kesehatan", WeightLight},
		{"Pola Hidup", WeightLight},
		{"Qiyamul-lail", WeightMedium},
		{"Hobi & Refreshing", WeightLight},
	},
	CatPersonalGrowth: {
		{"Meditasi / Zazen", WeightLight},
		{"Refleksi", WeightMedium},
		{"Filosofi & Agama", WeightMedium},
		{"Karakter & Disiplin", WeightMedium},
	},
}
