# 🛡️ Sistem Audit Internal Perusahaan (Cimory Audit System)

Aplikasi Web Enterprise untuk Sistem Monitoring Audit Internal, mencakup Manajemen Inspeksi GMP (Good Manufacturing Practice), Pengelolaan Temuan (*Issue*), Tindak Lanjut (*Follow-Up/WOWR*), Notifikasi Real-time, serta Integrasi Data ke **Power BI** berbasis *Role-Based Access Control* (RBAC) dinamis.

---

## 🚀 Teknologis & Arsitektur (*Tech Stack*)

### **Backend**
- **Bahasa & Framework**: Go (Golang 1.22+) dengan [Fiber v2](https://gofiber.io/) (Clean Architecture)
- **Database**: PostgreSQL 15 (GORM ORM)
- **Object Storage**: MinIO (S3 Compatible Storage untuk foto & file bukti)
- **Message Broker & Event Stream**: Apache Kafka & Zookeeper
- **Search & Analytics**: OpenSearch 2.11 & OpenSearch Dashboards
- **Keamanan**: JWT Authentication, AES-256 Data Encryption, Dynamic RBAC, & API Key Bearer Authentication

### **Frontend**
- **Framework**: Next.js 15 (App Router, React 19)
- **Bahasa**: TypeScript
- **Styling**: Vanilla CSS, Tailwind CSS, Dark/Light Mode, & Glassmorphism Design
- **State Management**: Zustand
- **Real-Time**: Server-Sent Events (SSE)

---

## ✨ Fitur Utama

1. **Role-Based Access Control (RBAC) & Dynamic Permissions**:
   - Multi-role dashboard: **Admin**, **Auditor**, dan **Auditee**.
   - Pengaturan izin per modul & per user override secara dinamis dari UI.
2. **Inspeksi GMP & Checklist Audit**:
   - Manajemen Master Data (Area, Kawasan, Department, Aspek, Detail, Uraian).
   - Eksekusi inspeksi area dan penilaian checklist otomatis.
3. **Manajemen Temuan (*Issue*) & WOWR**:
   - Pelaporan temuan audit beserta foto bukti.
   - Tindak lanjut perbaikan (*Follow-up*) oleh Auditee & verifikasi Auditor.
4. **Integrasi Power BI via Public API Key**:
   - Generator API Key publik di halaman Settings.
   - Endpoint khusus `/api/v1/public/powerbi/data` dengan filter incremental waktu (`?since=...`) dan opsi *Single-Use* / *Multi-Use*.
5. **Real-Time Notification & SSE Broker**:
   - Broker Server-Sent Events (SSE) untuk pengiriman notifikasi instan saat ada temuan baru atau perubahan status.
6. **Laporan & Ekspor Data**:
   - Ekspor laporan audit ke format Excel (*GMP Report Template*).
7. **Audit Trail & Activity Logging**:
   - Setiap aktivitas penting dicatat via Kafka ke OpenSearch & PostgreSQL untuk keperluan audit log.

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
│   ├── cmd/                  # Main entrypoints (server, migrate, seed, genkey)
│   ├── config/               # Load environment configuration
│   ├── internal/             # Core business logic
│   │   ├── domain/           # Domain entities & interfaces
│   │   ├── handler/          # HTTP Request Handlers
│   │   ├── infrastructure/   # Persistence/Repositories (GORM, MinIO, Kafka)
│   │   ├── middleware/       # JWT Auth, APIKey, Logging, CORS middleware
│   │   └── usecase/          # Business logic implementation
│   ├── router/               # Fiber route registration (v1)
│   └── templates/            # Excel report templates
├── frontend/                 # Source code Next.js 15 Frontend
│   ├── src/
│   │   ├── app/              # Next.js App Router pages & layouts
│   │   ├── components/       # UI Components (Admin, Auditor, Auditee, Master, Settings)
│   │   ├── hooks/            # Custom React hooks (useSSE, useChunkedUpload, etc.)
│   │   ├── lib/              # API clients & utilities
│   │   ├── stores/           # Zustand state management
│   │   └── types/            # TypeScript interfaces & types
│   └── public/               # Static assets & icons
├── docker-compose.yml        # Multi-container Docker orchestration
└── README.md                 # Dokumentasi proyek
```

---

## 📄 Lisensi

Hak Cipta © 2026 Cimory Internal Audit System. Seluruh hak cipta dilindungi.
