// cmd/seed-dummy/main.go
// Jalankan dengan: go run ./cmd/seed-dummy
//
// LOCAL/DEV TESTING ONLY. Mengisi Area/Kawasan/Inspection/Issue dummy
// (seeds.SeedDummyData, sudah ada di kode tapi tidak pernah dipanggil dari
// pipeline seed utama — seeds.Run malah memanggil CleanupDummyData supaya
// data ini tidak ikut ke lingkungan yang di-share) — dipakai untuk menguji
// Custom KPI Visualization Builder (measures/dimensions perlu ada baris
// Issue nyata untuk di-agregasi, kalau tidak semua chart tampil "Belum ada
// data"). Jangan dijalankan di database yang datanya harus tetap bersih;
// bersihkan lagi dengan `go run ./cmd/seed-cleanup-dummy` (lihat file
// seed-cleanup-dummy/main.go) setelah selesai menguji.
package main

import (
	"log"

	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/seeds"
)

func main() {
	cfg := config.Load()
	db, err := config.OpenDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	seeds.SeedDummyData(db)
	log.Println("✅ Dummy data seeded — Custom KPI Builder sekarang punya data untuk diagregasi.")
}
