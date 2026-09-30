// migrate — util CLI desktop untuk menyiapkan & memverifikasi skema database
// selama pengembangan (tanpa perlu menjalankan emulator Android).
package main

import (
	"fmt"
	"os"

	"mastermap/internal/db"
	"mastermap/internal/model"
)

func main() {
	path := "mastermap.db"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	gdb, err := db.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	var cats []model.Category
	gdb.Order("sort_order").Find(&cats)
	fmt.Printf("✅ Skema siap di %q — Master Map ter-seed:\n", path)
	for _, c := range cats {
		var nSub, nProj int64
		gdb.Model(&model.SubModule{}).Where("category_key = ?", c.Key).Count(&nSub)
		gdb.Model(&model.Project{}).Where("category_key = ?", c.Key).Count(&nProj)
		fmt.Printf("  %s %-15s %s (light %s | dark %s) — %d sub-modul, %d project\n",
			c.Icon, c.Name, c.ColorHex, c.ColorHex, c.ColorDarkHex, nSub, nProj)
	}

	fmt.Println("\nTabel yang dibuat:")
	type tbl struct{ Name string }
	var rows []tbl
	gdb.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Scan(&rows)
	for _, r := range rows {
		fmt.Println("  •", r.Name)
	}
}
