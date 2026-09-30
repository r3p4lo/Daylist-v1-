# 🗺️ Master Map Life Planner

Aplikasi produktivitas Android (Go + Fyne v2 + SQLite) dengan konsep **"Prinsip Modular + Satu Peta Induk"**: semua proyek dipetakan dalam satu master map berisi 6 kategori berwarna, lengkap dengan **Dynamic Time-boxing Engine** dan **Focus Mode timer yang akurat meski app di-minimize**.

| | Kategori | Warna (Light / Dark) |
|---|---|---|
| 🎓 | Learning | `#2E7D32` / varian emerald terang |
| ⚙️ | Technology | `#1565C0` / varian royal blue terang |
| 💰 | Finance | `#E65100` / varian deep orange terang |
| 💼 | Career | `#6A1B9A` / varian purple terang |
| ❤️ | Lifestyle | `#C62828` / varian crimson terang |
| 🌱 | Personal Growth | `#00695C` / varian teal terang |

**Fitur utama**
- ✅ Master Map Home: 6 kartu kategori + progress bar per kategori
- ✅ Tambah "Quest" (Project) di setiap kategori — auto-generate 5 fase sub-task (Riset → Desain → Implementasi → Testing → Dokumentasi)
- ✅ Modul Detail & Project Detail (CRUD project/sub-task, checklist)
- ✅ Dynamic Time-boxing: rekomendasi durasi Blok A (30 mnt) / Blok B (50 mnt) berdasarkan Bobot Kognitif, bisa override manual, dengan jeda istirahat (contoh: riset 120 mnt → `[50m, ☕10m, 50m, ☕10m, 20m]`)
- ✅ Daily Timeline color-coded + Focus Mode countdown (basis timestamp, tahan minimize/restart)
- ✅ Mode Terang / Gelap (live switch, preferensi tersimpan di SQLite)
- ✅ Storage lokal murni (SQLite, driver Go-native `glebarez/sqlite` → gomobile-safe, tanpa cgo)

---

## 📁 Struktur Proyek (modular)

```
mastermap/
├── cmd/
│   ├── desktop/main.go      # Jalankan versi desktop (dev/test UI)
│   └── migrate/main.go      # Verifier skema DB + seed awal
├── internal/
│   ├── model/               # Struct GORM: Category→SubModule→Project→SubTask→TimeBlock
│   ├── db/                  # Open(), AutoMigrate, PRAGMA WAL, seed idempotent 6 kategori
│   ├── engine/              # ⭐ Dynamic Time-boxing Engine (+ unit test)
│   ├── timer/               # Timer berbasis timestamp selisih (akurat lintas minimize)
│   ├── theme/               # Custom Theme Engine: palette 6 warna + dark/light
│   ├── state/               # AppState terpusat (DB, darkMode *bool, timer)
│   └── ui/                  # Layar: master_map, module_detail, project_detail,
│                            #        timeline, focus_mode, settings, navigator (transisi slide)
├── FyneApp.toml             # Metadata app (ID, ikon) untuk fyne package
├── go.mod
└── README.md
```

---

# 🚀 CARA INSTALASI — JADI 1 PERINTAH, TINGGAL JALANKAN

## Opsi A — Desktop (paling cepat, untuk mencoba UI)

Prasyarat: **Go 1.22+** dan toolchain C dasar (`gcc`; Linux juga butuh `xorg-dev`, macOS sudah siap via Xcode CLT).

```bash
git clone <repo-anda> mastermap && cd mastermap   # atau langsung masuk folder proyek
go run ./cmd/desktop                              # ← SELESAI, aplikasi terbuka
```

Database dibuat otomatis di `~/.mastermap/mastermap.db` beserta seed 6 kategori, 24 sub-modul, dan 6 project contoh. Tidak ada langkah migrasi manual.

### Prasyarat OS Linux (sekali seumur mesin):
```bash
sudo apt update && sudo apt install -y gcc git make pkg-config xorg-dev libglu1-mesa-dev
```

## Opsi B — Build APK Android — 1 skrip, tinggal jalankan

Prasyarat besar (sekali saja): [Android SDK](https://developer.android.com/studio) + NDK, lalu set env:
```bash
export ANDROID_HOME=$HOME/Android/Sdk
export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/ndk/<versi>/toolchains/llvm/prebuilt/linux-x86_64/bin
```

Alat build: **fyne-cross** (Docker) — paling stabil untuk gomobile:
```bash
go install github.com/fyne-io/fyne-cross@latest   # butuh Docker berjalan
./build_apk.sh                                    # ← SATU PERINTAH: hasil di ./fyne-cross/dist/android/*.apk
```

Alternatif tanpa Docker (gomobile langsung):
```bash
go install golang.org/x/mobile/cmd/gomobile@latest && gomobile init
fyne package -os android -appID com.mastermap.life -icon Icon.png -o MasterMap.apk ./cmd/desktop
```

Pasang ke HP: `adb install MasterMap.apk` (atau salin file APK).

---

## 🔧 Perintah Berguna Lainnya

```bash
go test ./...            # unit test engine time-boxing
go vet ./...             # static check
go run ./cmd/migrate     # cek isi & skema database (kategori/project/plan hari ini)
```

Lokasi DB: desktop `~/.mastermap/mastermap.db` • Android `files:///data/data/<appID>/mastermap.db` (Settings → Backup mengekspor JSON ke folder Documents).

## 🧠 Catatan Arsitektur Singkat

- **Timer**: `sisa = DurationMin×60 − (now − StartedAt) + PausedAccumSec` — kebenaran angka selalu dari DB, UI hanya merender; app di-kill Android pun sesi fokus tetap pulih akurat.
- **Engine**: bobot kognitif (Ringan 15–30 mnt / Sedang 45–60 / Berat 50–90) menentukan panjang blok & jeda; `PlanWithBudget` memotong rantai sesuai budget energi harian agar satu kategori tidak menyedot semua waktu.
- **Tema**: satu sumber kebenaran `*bool` dark-mode → seluruh widget (kartu, bar, chip) ikut berganti instan, preferensi persist di tabel `app_settings`.
- **Modularitas data**: setiap tabel anak menyimpan `category_key` → query & operasi per kategori terisolasi, tidak saling mengganggu.
