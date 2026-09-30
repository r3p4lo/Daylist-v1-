#!/usr/bin/env bash
# ============================================================================
# build_apk.sh — SATU PERINTAH build APK "Master Map Life Planner"
# Cara pakai : ./build_apk.sh
# Hasil      : ./fyne-cross/dist/android/mastermap.apk
# Prasyarat  : Docker berjalan + fyne-cross terinstall:
#              go install github.com/fyne-io/fyne-cross@latest
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")"

APP_ID="com.mastermap.life"
ICON="Icon.png"

echo "🔍 [1/4] Verifikasi toolchain..."
if ! docker info >/dev/null 2>&1; then
  echo "❌ Docker tidak berjalan. Nyalakan Docker dulu, lalu ulangi."
  exit 1
fi
command -v fyne-cross >/dev/null 2>&1 || {
  echo "📦 fyne-cross belum ada — menginstall..."
  go install github.com/fyne-io/fyne-cross@latest
}

echo "🧪 [2/4] Unit test & vet (agar APK tidak membawa bug)..."
go vet ./... && go test ./internal/engine/

# Buat ikon placeholder 512x512 jika belum ada (PNG biru royal #1565C0)
if [ ! -f "$ICON" ]; then
  echo "🎨 [3/4] Ikon belum ada — membuat $ICON placeholder..."
  python3 - "$ICON" <<'PY'
import sys, zlib, struct
w = h = 512
def chunk(t, d):
    c = t + d
    return struct.pack(">I", len(d)) + c + struct.pack(">I", zlib.crc32(c))
raw = b"".join(b"\x00" + bytes([0x15, 0x65, 0xC0, 255]) * w for _ in range(h))
png = (b"\x89PNG\r\n\x1a\n"
       + chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 6, 0, 0, 0))
       + chunk(b"IDAT", zlib.compress(raw, 9))
       + chunk(b"IEND", b""))
open(sys.argv[1], "wb").write(png)
PY
else
  echo "🎨 [3/4] Ikon $ICON sudah ada — dipakai apa adanya."
fi

echo "🚀 [4/4] fyne-cross build android/arm64 + arm (±5–15 menit build pertama)..."
fyne-cross android \
  --arch=arm64,arm \
  --name="Master Map Life Planner" \
  --app-id="$APP_ID" \
  --icon="$ICON" \
  ./cmd/desktop

APK=$(ls fyne-cross/dist/android/*.apk 2>/dev/null | head -1)
echo ""
echo "✅ SELESAI! APK: ${APK:-fyne-cross/dist/android/}"
echo "   Pasang ke HP : adb install -r \"$APK\""
echo "   Optimasi     : unzip -p apk assets -> hapus arsitektur 'arm' bila HP arm64 saja"
