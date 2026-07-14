# Sistem Audit Internal Perusahaan

Aplikasi untuk sistem monitoring audit internal, mencakup manajemen temuan (issue) dan tindak lanjut, berbasis Role-Based Access Control (RBAC) dinamis.

## Struktur Proyek
- `backend/`: API Server menggunakan Golang (Clean Architecture)
- `frontend/`: (Reserved untuk UI Web App)

## Persyaratan Backend
- Go 1.22+
- PosgreSQL 
- Make (opsional untuk helper scripts)
- Docker & Docker Compose

## Cara Menjalankan (Local Development)

1. Masuk ke direktori backend
   ```bash
   cd backend
   ```
2. Copy env file
   ```bash
   cp .env.example .env
   ```
3. Sesuaikan koneksi database di `.env`. Pastikan database `monitoring_audit` sudah dibuat di MySQL Anda.
4. Download dependencies
   ```bash
   go mod tidy
   ```
5. Jalankan migrasi dan seeder
   ```bash
   go run ./cmd/migrate/main.go
   go run ./cmd/seed/main.go
   ```
6. Jalankan server
   ```bash
   go run ./cmd/server/main.go
   ```
API akan berjalan di `http://localhost:8080`.
7. jalankan gen untuk mendapatkan encryption key
   go run ./cmd/genkey/main.go
8. update  .env dengan encryption key
   SETTING_ENCRYPTION_KEY=
   MINIO_ALLOWED_IPS=IP,IP,IP
   DB_SSLMODE=disable
   
## Menjalankan dengan Docker

Anda bisa menggunakan Docker Compose dari root proyek (tempat file `docker-compose.yml` berada).

```bash
docker-compose up -d
```
Docker akan:
1. Menjalankan container MySQL (`monitoring_db`)
2. Menjalankan container Backend API (`monitoring_backend`)
3. Membuat volume untuk data DB, upload foto, dan log.

> **Note:** Pada saat docker up pertama kali, migrasi dan seeder *belum* dijalankan secara otomatis. Anda bisa mengeksekusinya ke dalam container backend:
```bash
docker exec -it monitoring_backend ./main
# Atau setup command khusus untuk migrate/seed di dalam docker entrypoint
```

## Fitur Utama Backend
- Autentikasi JWT
- Role & Permission dinamis (Admin dapat mengatur izin role tanpa mengubah kode)
- Master Data Terpusat (Area, Kawasan, Department)
- Template Audit (Aspek, Detail, Uraian)
- Eksekusi Audit (Header, Result)
- Manajemen Issue dan Bukti Tindak Lanjut (Photo Upload)
- Audit Trail / Activity Logging untuk semua aksi sistem.
