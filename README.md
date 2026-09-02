# 🛡️ Sistem Audit Internal Perusahaan (Cimory Audit System)

Aplikasi Web Enterprise untuk Sistem Monitoring Audit Internal — satu alur utuh mulai dari **Inspeksi GMP** (Good Manufacturing Practice), **Pengelolaan Temuan** (*Issue*), **Tindak Lanjut** (*Follow-Up/WOWR*), **Notifikasi Real-time**, hingga **Analitik & Pelaporan** (Dashboard KPI, Custom Visualization Builder, ekspor Excel, integrasi **Power BI**) — semuanya berbasis *Role-Based Access Control* (RBAC) dinamis.

---

## 🚀 Teknologis & Arsitektur (*Tech Stack*)

### **Backend**
- **Bahasa & Framework**: Go (Golang 1.22+) dengan [Fiber v2](https://gofiber.io/) (Clean Architecture — `domain` → `usecase` → `infrastructure`/`handler`)
- **Database**: PostgreSQL 15 (GORM ORM)
- **Object Storage**: MinIO (S3 Compatible Storage untuk foto & file bukti)
- **Message Broker & Event Stream**: Apache Kafka & Zookeeper
- **Search & Analytics**: OpenSearch 2.11 & OpenSearch Dashboards
- **Ekspor Excel**: [excelize](https://github.com/xuri/excelize) — template narasi, tabel data GMP, laporan Temuan, & import master data
- **Keamanan**: JWT Authentication, AES-256 Data Encryption, Dynamic RBAC, & API Key Bearer Authentication

### **Frontend**
- **Framework**: Next.js 15 (App Router, React 19)
- **Bahasa**: TypeScript
- **Styling**: Vanilla CSS, Tailwind CSS, Dark/Light Mode, & Glassmorphism Design
- **State Management**: Zustand
- **Real-Time**: Server-Sent Events (SSE)
- **Data Visualization**: [ECharts](https://echarts.apache.org/) (Dashboard KPI & Custom Visualization Builder) & Recharts (widget bawaan)

---

## 🔄 Alur Kerja Sistem (End-to-End)

Begini perjalanan satu temuan audit dari awal sampai jadi angka di dashboard:

1. **Login & RBAC** — user login sesuai role (**Admin**, **Auditor**, **Auditee**, dan role turunan seperti Supervisor/Manager/Staff); setiap halaman & aksi dibatasi izin per modul yang diatur dinamis dari UI.
2. **Setup Master Data** — Admin menyiapkan struktur Area → Kawasan → Detail Kawasan, serta Aspek → Detail → Uraian checklist dan Klasifikasi HEI. Bisa diisi manual satu-satu, atau **import massal via Excel** untuk data dalam jumlah besar.
3. **Eksekusi Inspeksi** — Auditor menjalankan inspeksi GMP di lapangan, mengisi checklist per Uraian dengan penilaian OK/NG.
4. **Temuan (*Issue*) Otomatis Tercatat** — setiap item checklist yang gagal (NG) otomatis menjadi sebuah *Issue*, lengkap dengan foto bukti awal, PIC penanggung jawab, dan *due date*.
5. **Tindak Lanjut (*Follow-Up*)** — Auditee memperbaiki temuan dan mengunggah bukti foto follow-up + keterangan; jika perlu perbaikan fisik lebih besar, diajukan sebagai **WO/WR** (Work Order/Work Request).
6. **Verifikasi** — Auditor memverifikasi bukti follow-up (dan WO/WR bila ada) sebelum status Temuan ditutup.
7. **Notifikasi Real-time** — setiap perpindahan status (temuan baru, mendekati *due date*, follow-up masuk, WO/WR terverifikasi, dst.) langsung dikirim ke pihak terkait lewat SSE + email.
8. **Pelaporan & Ekspor** — data GMP maupun Temuan bisa difilter (facet: status, PIC, rentang tanggal, pencarian bebas teks) dan diekspor ke Excel kapan saja.
9. **Analitik** — ringkasan kepatuhan & tren tersedia lewat 6 widget KPI bawaan, atau dibangun sendiri lewat **Custom Visualization Builder** tanpa menulis SQL.
10. **Audit Trail** — seluruh aksi penting di atas dicatat via Kafka ke OpenSearch & PostgreSQL, dan/atau diteruskan ke **Power BI** eksternal lewat Public API Key untuk pelaporan korporat.

---

## ✨ Fitur Utama

### 1. Role-Based Access Control (RBAC) & Dynamic Permissions
- Multi-role dashboard: **Admin**, **Auditor**, **Auditee**, dengan role turunan (Supervisor, Manager, Staff) per plant.
- Pengaturan izin per modul & per user override secara dinamis dari UI — tidak perlu redeploy untuk ubah akses.
- Sesi otomatis me-refresh role/plant terbaru saat halaman dimuat ulang (tidak perlu logout manual setelah role diubah Admin).

### 2. Master Data & Import Excel
- Manajemen manual: Area, Kawasan, Detail Kawasan, Department, Aspek, Detail, Uraian, Klasifikasi HEI.
- **Import massal via Excel** (`PERM-MSTR-I`) untuk Aspek, Detail, Uraian, dan HEI:
  - Import berjenjang Aspek → Detail → Uraian — template hasil unduh otomatis menyertakan sheet `REFERENSI` berisi ID hasil commit sebelumnya, sehingga baris berikutnya tinggal merujuk, bukan mengetik ulang.
  - HEI independen dari ketiga master audit di atas, kategori baru terbentuk otomatis dari data yang diimpor.
  - Validasi ketat: hanya `.xlsx` resmi, maksimal 5 MB & 5.000 baris, menolak formula/macro/external link/embedded object/merged cell/sheet tambahan.
  - Commit atomik — satu baris invalid membatalkan seluruh batch, tidak ada partial import.
  - Deduplikasi di dalam file maupun database (fungsi normalisasi Postgres yang sama dengan *unique index*), termasuk deduplikasi HEI per kategori.
  - Token validasi berlaku 20 menit, sekali pakai, terikat ke satu kombinasi file/user/plant/tipe.

### 3. Inspeksi GMP & Checklist Audit
- Eksekusi inspeksi per Area/Kawasan/Detail Kawasan dengan penilaian checklist otomatis (OK/NG per Uraian).
- Persentase Kepatuhan dihitung otomatis per Detail Kawasan.
- Cutoff periode inspeksi bulanan yang bisa dikonfigurasi Admin (hari mulai/akhir custom).

### 4. Manajemen Temuan (*Issue*), Filter Lanjutan, & WOWR
- Temuan otomatis tercatat dari hasil checklist NG, lengkap foto bukti.
- Tindak lanjut perbaikan (*Follow-up*) oleh Auditee & verifikasi Auditor; eskalasi ke WO/WR (Work Order/Work Request) untuk perbaikan yang butuh proses lebih formal.
- **Filter & pencarian lanjutan** (`GET /issues/filter`) dengan facet count real-time (status, status WOWR, PIC, rentang tanggal dibuat & jatuh tempo), pencarian bebas teks, dan sort multi-kolom.
- **Ekspor hasil filter ke Excel** (`PERM-ISS-E`, permission khusus export, tersedia untuk Admin/Auditor/Manager).

### 5. Laporan & Ekspor Data GMP
- Ekspor laporan audit ke Excel dengan 2 pilihan tampilan: **Template** (narasi per inspeksi) atau **Table** (16 kolom persis tabel web Data GMP).
- Bukti foto awal temuan **dan** bukti tindak lanjut (*follow-up*) ditampilkan & diekspor berdampingan — terpisah dari bukti WO/WR yang punya makna berbeda.
- Metrik **follow-up gap days** dihitung otomatis (selisih hari kalender antara *due date* dan tanggal tindak lanjut, zona waktu Asia/Jakarta).
- Filter pencarian bebas teks pada halaman Data GMP, selain filter Area/Kawasan/tanggal yang sudah ada.

### 6. Dashboard KPI & Custom Visualization Builder
- Dashboard analitik terpisah (`/kpi`) dengan 6 widget bawaan (ringkasan, tren, ranking PIC/Auditor, WO/WR per Area, progres Area).
- Builder *drag-and-drop* ala Power BI ("Tambah Visualisasi") — kombinasikan Kategori & Nilai dari katalog measure/dimension yang sudah dikurasi developer, tanpa menulis SQL sama sekali.
- **16 jenis chart**: Bar, Bar Horizontal, Bar Bertumpuk, Line, Area, Radar, Scatter, Pie, Donut, Treemap, Funnel, Kartu Angka, Gauge, Heatmap, Sankey, dan Tabel — dengan validasi kompatibilitas otomatis per kombinasi field.
- **Drill-down interaktif**: susun beberapa Kategori berjenjang, klik langsung pada chart untuk turun ke level berikutnya (dengan breadcrumb navigasi), tanpa membangun ulang widget.
- **Opsi tampilan**: aktifkan data label, tampilkan nilai di label, dan tambahkan garis tren di atas Bar Chart.
- Widget kustom bisa diedit ulang kapan saja (judul, field, jenis chart, opsi tampilan) langsung dari dashboard, tanpa perlu dihapus & dibuat ulang.
- Setiap query dijaga *security backbone* yang otomatis membatasi hasil sesuai akses Area/Plant milik user yang login — tidak ada cara membaca data di luar cakupan akses lewat builder ini.

### 7. Integrasi Power BI via Public API Key
- Generator API Key publik di halaman Settings.
- Endpoint khusus `/api/v1/public/powerbi/data` dengan filter incremental waktu (`?since=...`) dan opsi *Single-Use* / *Multi-Use*.

### 8. Real-Time Notification & SSE Broker
- Broker Server-Sent Events (SSE) untuk notifikasi instan: temuan baru, pengingat mendekati *due date*, follow-up masuk, WO/WR terverifikasi, dan penyelesaian Kawasan — dilengkapi email untuk sebagian jenis notifikasi.
- Toggle aktif/nonaktif per jenis notifikasi.

### 9. Audit Trail & Activity Logging
- Setiap aktivitas penting dicatat via Kafka ke OpenSearch & PostgreSQL untuk keperluan audit log — termasuk perubahan permission, aksi import, dan ekspor data.

---

## 🛠️ Persyaratan Sistem (*Prerequisites*)

Sebelum menjalankan proyek, pastikan perangkat Anda telah terinstall:
- **Docker** & **Docker Compose** (Rekomendasi utama)
- **Go 1.22+** (Jika ingin menjalankan backend secara lokal)
- **Node.js 18+** & **npm** / **pnpm** (Jika ingin menjalankan frontend secara lokal)
- **PostgreSQL 15+** (Opsional jika running tanpa Docker)

---

## 🐳 Cara Menjalankan dengan Docker Compose (Rekomendasi)

Cara termudah dan tercepat untuk menjalankan seluruh ekosistem aplikasi (Database, Storage, Message Broker, Backend API, & Frontend Web).

1. **Clone Repository & Masuk ke Folder Proyek**:
   ```bash
   git clone https://github.com/bimoBintang/System-Audit-.git
   cd System-Audit-
   ```

2. **Jalankan Docker Compose**:
   ```bash
   cp .env.example .env
   # Ganti seluruh nilai CHANGE_ME dengan secret acak sebelum melanjutkan.
   docker compose up -d --build
   ```

3. **Akses Layanan**:
   - 🌐 **Frontend App**: `http://localhost:3000`
   - ⚡ **Backend API**: `http://localhost:8080`
   - 🗄️ **MinIO Console**: `http://localhost:9001` *(User: `minioadmin`, Pass: `minioadmin`)*
   - 📊 **OpenSearch Dashboards**: `http://localhost:5601`

4. **Jalankan Database Seeder (Pertama Kali Run)**:
   ```bash
   docker exec -it monitoring_backend ./main seed
   # atau jalankan script seeder lokal ke container DB
   ```

5. **(Opsional) Isi Data Dummy untuk Menguji Dashboard KPI**:
   Seeder utama sengaja tidak menyertakan data Issue/Inspection dummy (supaya lingkungan production/shared tetap bersih). Untuk mencoba Custom KPI Visualization Builder secara lokal dengan data contoh:
   ```bash
   cd backend
   go run ./cmd/seed-dummy          # isi Area/Kawasan/Inspection/Issue dummy
   go run ./cmd/seed-cleanup-dummy  # bersihkan lagi kapan saja
   ```

---

## 💻 Cara Menjalankan Secara Manual (Development Mode)

Jika Anda ingin melakukan pengembangan (*development*) secara terpisah pada Backend atau Frontend:

### **1. Persiapan Database & Infrastructure**
Jalankan container dependency (PostgreSQL, MinIO, Kafka, OpenSearch) menggunakan Docker:
```bash
docker compose up -d db minio zookeeper kafka opensearch-node1
```

---

### **2. Setup & Jalankan Backend (Golang)**

1. Masuk ke direktori `backend`:
   ```bash
   cd backend
   ```

2. Salin file environment:
   ```bash
   cp .env.example .env
   ```

3. Install dependensi Go:
   ```bash
   go mod tidy
   ```

4. Jalankan Migrasi Database & Seeder:
   ```bash
   go run ./cmd/migrate/main.go
   go run ./cmd/seed/main.go
   ```

5. Jalankan Backend Server:
   ```bash
   go run ./cmd/server/main.go
   ```
   *Backend API akan aktif di `http://localhost:8080`.*

---

### **3. Setup & Jalankan Frontend (Next.js)**

1. Buka terminal baru dan masuk ke direktori `frontend`:
   ```bash
   cd frontend
   ```

2. Install dependensi Node.js:
   ```bash
   npm install
   ```

3. Salin file environment (opsional untuk kostumisasi URL API):
   ```bash
   cp .env.example .env.local
   ```

4. Jalankan Frontend Development Server:
   ```bash
   npm run dev
   ```
   *Frontend Web akan aktif di `http://localhost:3000`.*

---

## 🔑 Akun Akses Default (Seeder)

Setelah seeder dijalankan, Anda dapat login menggunakan akun berikut:

| Role | Username / User ID | Sumber Password | Akses Utama |
| :--- | :--- | :--- | :--- |
| **Admin** | `USR-ADMIN-001` | `SEED_ADMIN_PASSWORD` | Full Access & Master Data & API Key Manager |
| **Auditor** | `USR-AUDITOR-001` | `SEED_AUDITOR_PASSWORD` | Inspeksi, Pelaporan Issue, & Verifikasi |
| **Auditee** | `USR-AUDITEE-001` | `SEED_AUDITEE_PASSWORD` | Tindak Lanjut Issue (*Follow-Up*) |

Seeder menolak password yang panjangnya kurang dari 32 karakter dan tidak pernah mencetak password ke log.

---

## 📁 Struktur Direktori Proyek

```
System-Audit-/
├── backend/                  # Source code Go Backend (Clean Architecture)
│   ├── cmd/                  # Main entrypoints (server, migrate, seed, seed-dummy, genkey)
│   ├── config/               # Load environment configuration
│   ├── internal/             # Core business logic
│   │   ├── domain/           # Domain entities & interfaces (termasuk analytics/ — katalog KPI Builder)
│   │   ├── handler/          # HTTP Request Handlers (termasuk analytichandler/ — endpoint KPI Builder)
│   │   ├── infrastructure/   # Persistence/Repositories (GORM, MinIO, Kafka)
│   │   ├── middleware/       # JWT Auth, APIKey, Logging, CORS middleware
│   │   └── usecase/          # Business logic implementation (termasuk analyticsusecase/ — query engine KPI Builder)
│   ├── router/                # Fiber route registration (v1)
│   └── templates/            # Excel report templates
├── frontend/                 # Source code Next.js 15 Frontend
│   ├── src/
│   │   ├── app/              # Next.js App Router pages & layouts (termasuk /kpi — Dashboard KPI)
│   │   ├── components/       # UI Components (Admin, Auditor, Auditee, Master, Settings)
│   │   │   ├── dashboard/kpi/    # Custom Visualization Builder & widget renderer (ECharts)
│   │   │   ├── gmp/              # Menu ekspor & tampilan bukti follow-up Data GMP
│   │   │   └── master/import/    # UI wizard import master data via Excel
│   │   ├── hooks/             # Custom React hooks (useSSE, useChunkedUpload, etc.)
│   │   ├── lib/                # API clients & utilities
│   │   ├── stores/            # Zustand state management
│   │   └── types/              # TypeScript interfaces & types
│   └── public/                # Static assets & icons
├── docs/                      # Dokumentasi tambahan (mis. master-data-import.md)
├── docker-compose.yml         # Multi-container Docker orchestration
└── README.md                  # Dokumentasi proyek
```

---

## 📄 Lisensi

Hak Cipta © 2026 Cimory Internal Audit System. Seluruh hak cipta dilindungi.
