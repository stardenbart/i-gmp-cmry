# Product Requirements Document (PRD)
# Sistem Audit Internal Perusahaan — Cimory Audit System

**Versi:** 1.0.0  
**Tanggal:** 2026-07-27  
**Status:** Production-Ready  
**Tim Pengembang:** Internal Engineering Team  

---

## 1. Executive Summary

Cimory Audit System adalah platform digital untuk mengelola seluruh siklus audit internal GMP (Good Manufacturing Practice) di lingkungan perusahaan manufaktur Cimory. Sistem ini menggantikan proses audit manual berbasis kertas menjadi workflow digital yang terintegrasi — mulai dari perencanaan inspeksi, penilaian checklist di lapangan, pelaporan temuan (issue), tindak lanjut perbaikan (follow-up/WOWR), hingga analitik data via Power BI.

---

## 2. Business Context & Problem Statement

### 2.1 Masalah Yang Diselesaikan

| No | Masalah Existing | Solusi dalam Sistem |
|:---|:---|:---|
| 1 | Proses audit GMP dilakukan manual dengan form kertas | Digitalisasi checklist inspeksi per area/kawasan |
| 2 | Temuan audit tidak terlacak secara real-time | Dashboard monitoring status issue dengan notifikasi SSE |
| 3 | Tindak lanjut temuan tidak terdokumentasi dengan baik | Modul Follow-Up & WOWR dengan upload foto bukti |
| 4 | Data audit tidak bisa dianalisis secara agregat | Integrasi Power BI via public API Key endpoint |
| 5 | Hak akses tidak terstruktur | RBAC dinamis: Admin, Auditor, Auditee dengan user-level override |
| 6 | Tidak ada audit trail untuk perubahan data | Activity Log otomatis di setiap endpoint berhasil |

### 2.2 Business Objectives

1. Mempercepat siklus audit dari perencanaan hingga verifikasi perbaikan
2. Meningkatkan akuntabilitas PIC temuan melalui notifikasi dan batas waktu (due date)
3. Menyediakan dashboard eksekutif berbasis data real-time untuk manajemen
4. Memudahkan integrasi data audit ke Business Intelligence (Power BI)

---

## 3. Stakeholders

| Role | Deskripsi | Tanggung Jawab Utama |
|:---|:---|:---|
| **Admin** | System Administrator | Kelola master data, user, role, permissions, settings, API Key |
| **Auditor** | Petugas Audit Internal | Eksekusi inspeksi area, penilaian checklist, pelaporan issue |
| **Auditee** | PIC Area (Person in Charge) | Upload bukti tindak lanjut, pengisian WOWR, tanda terima issue |
| **Manajemen** | Eksekutif / Pimpinan | Melihat dashboard overview, laporan summary, analitik Power BI |

---

## 4. Feature Requirements

### 4.1 Modul Autentikasi & Manajemen Akses

**FR-AUTH-01:** Sistem harus mendukung login dengan Username dan Password.  
**FR-AUTH-02:** Autentikasi menggunakan JWT (JSON Web Token) dengan expiry configurable.  
**FR-AUTH-03:** Sistem harus mencatat setiap login (sukses/gagal) ke Login_Log.  
**FR-AUTH-04:** Admin dapat mereset password user.  
**FR-AUTH-05:** User dapat mengubah password sendiri.  
**FR-AUTH-06:** Sistem mendukung lupa password via email OTP.

### 4.2 Modul RBAC & Permissions

**FR-RBAC-01:** Terdapat 3 role utama: Admin, Auditor, Auditee.  
**FR-RBAC-02:** Admin dapat mengatur permission per role secara dinamis dari UI tanpa perubahan kode.  
**FR-RBAC-03:** Admin dapat memberikan permission override spesifik per user (di atas role permission).  
**FR-RBAC-04:** Permission check berlapis: cek user override dulu → jika tidak ada, cek role permission.  
**FR-RBAC-05:** Permission codes yang tersedia: CREATE, READ, UPDATE, DELETE, APPROVE, EXPORT.

### 4.3 Modul Master Data

**FR-MASTER-01:** Admin dapat mengelola Departemen (CRUD).  
**FR-MASTER-02:** Admin dapat mengelola Area audit (CRUD).  
**FR-MASTER-03:** Admin dapat mengelola Kawasan (sub-area) dan Detail Kawasan (CRUD).  
**FR-MASTER-04:** Admin dapat mengelola Aspek, Detail, dan Uraian audit (hierarki 3-level) (CRUD).  
**FR-MASTER-05:** Setiap Uraian memiliki Standard Score yang digunakan dalam kalkulasi nilai inspeksi.  
**FR-MASTER-06:** Admin dapat mengatur System Settings (SMTP, aplikasi, enkripsi) dari UI.

### 4.4 Modul PIC Mapping

**FR-PIC-01:** Admin dapat memetakan User (Auditee) ke Kawasan tertentu sebagai PIC.  
**FR-PIC-02:** PIC Mapping digunakan untuk assignment issue otomatis saat inspeksi.

### 4.5 Modul Inspeksi

**FR-INSP-01:** Auditor dapat membuat Inspection Header untuk area/kawasan/detail kawasan tertentu.  
**FR-INSP-02:** Sistem membuat Inspection Result otomatis dari daftar Uraian yang aktif di area terkait.  
**FR-INSP-03:** Auditor dapat mengisi nilai (OK/NG/NA) dan keterangan per Uraian.  
**FR-INSP-04:** Status inspeksi: Draft → Ongoing → Completed → Approved.  
**FR-INSP-05:** Setelah Completed, sistem otomatis membuat Issue untuk setiap Uraian yang NG.  
**FR-INSP-06:** Issue otomatis di-assign ke PIC berdasarkan PIC Mapping kawasan.  
**FR-INSP-07:** Sistem mengirim notifikasi ke PIC saat issue baru dibuat.

### 4.6 Modul Issue Management

**FR-ISSUE-01:** Auditor dapat membuat Issue manual (di luar inspeksi).  
**FR-ISSUE-02:** Issue memiliki status lifecycle: Open → InProgress → PendingValidation → Closed/Verified.  
**FR-ISSUE-03:** PIC (Auditee) dapat mengupload foto bukti perbaikan (Follow-Up Photo).  
**FR-ISSUE-04:** Auditor dapat memverifikasi bukti dan mengubah status issue.  
**FR-ISSUE-05:** Sistem mendukung delegasi issue ke user lain.  
**FR-ISSUE-06:** Issue dapat memiliki label dan due date.

### 4.7 Modul WOWR (Work Order / Work Request)

**FR-WOWR-01:** Issue yang membutuhkan perbaikan teknis dapat ditandai `NeedsWOWR = true`.  
**FR-WOWR-02:** PIC dapat mengisi Work Order ID (WO_ID) dan Work Request ID (WR_ID).  
**FR-WOWR-03:** Status WOWR: None → PendingValidation → Verified/Rejected.  
**FR-WOWR-04:** Auditor dapat memverifikasi atau menolak WOWR.

### 4.8 Modul Laporan & Ekspor

**FR-RPT-01:** Admin/Auditor dapat mengeksport data inspeksi ke Excel (GMP Report Template).  
**FR-RPT-02:** Admin dapat melihat statistik dashboard: total inspeksi, issue per status, skor rata-rata per area.  
**FR-RPT-03:** Sistem menyediakan endpoint public `/api/v1/public/powerbi/data` untuk integrasi Power BI.

### 4.9 Modul API Key (Power BI Integration)

**FR-APIKEY-01:** Admin dapat membuat API Key dari halaman Settings.  
**FR-APIKEY-02:** API Key mendukung mode Single-Use (sekali pakai lalu expire) atau Multi-Use.  
**FR-APIKEY-03:** API Key dikomputer dengan SHA-256 sebelum disimpan ke database (tidak pernah disimpan plaintext).  
**FR-APIKEY-04:** Endpoint Power BI mendukung filter incremental waktu via parameter `?since=<datetime>`.  
**FR-APIKEY-05:** Autentikasi menggunakan `Authorization: Bearer MAKEY_...` atau `?api_key=MAKEY_...`.  
**FR-APIKEY-06:** Admin dapat mencabut (revoke) API Key kapan saja.

### 4.10 Modul Notifikasi

**FR-NOTIF-01:** Sistem mengirim notifikasi in-app saat: issue baru dibuat, status issue berubah, bukti follow-up diupload.  
**FR-NOTIF-02:** Notifikasi dikirim secara real-time via Server-Sent Events (SSE).  
**FR-NOTIF-03:** User dapat menandai notifikasi sebagai sudah dibaca.  
**FR-NOTIF-04:** User dapat melihat daftar notifikasi dengan pagination.

### 4.11 Modul Audit Trail & Logging

**FR-LOG-01:** Setiap request berhasil (2xx) dari user terautentikasi dicatat ke Activity_Log.  
**FR-LOG-02:** Activity log disimpan ke PostgreSQL dan diindeks ke OpenSearch secara async via Kafka.  
**FR-LOG-03:** Admin dapat mencari dan memfilter activity log via OpenSearch.

---

## 5. Non-Functional Requirements (NFR)

| ID | Kategori | Requirement |
|:---|:---|:---|
| NFR-01 | Performance | Response time API < 500ms pada kondisi normal |
| NFR-02 | Performance | Dashboard statistik harus ter-cache atau ter-aggregate efisien |
| NFR-03 | Security | Semua endpoint kecuali `/auth/login`, `/auth/forgot-password`, dan `/public/powerbi/data` harus memerlukan autentikasi |
| NFR-04 | Security | Password disimpan dengan bcrypt hashing |
| NFR-05 | Security | Setting sensitif (SMTP password, encryption key) disimpan terenkripsi (AES-256) |
| NFR-06 | Availability | Target uptime 99.5% |
| NFR-07 | Scalability | Arsitektur event-driven (Kafka) memungkinkan horizontal scaling consumer |
| NFR-08 | Observability | Semua log diindeks ke OpenSearch; metrics dapat diintegrasikan dengan Prometheus/Grafana |
| NFR-09 | Data Integrity | Semua foreign key constraints harus enforced di database layer |
| NFR-10 | Audit | Seluruh perubahan data sensitif harus terlacak di Activity_Log |

---

## 6. User Stories (Utama)

```
US-001: Sebagai Auditor, saya ingin membuat sesi inspeksi baru untuk area tertentu
        agar saya dapat menilai kondisi GMP secara sistematis.

US-002: Sebagai Auditor, saya ingin melihat semua checklist uraian yang perlu dinilai
        dalam satu sesi inspeksi sehingga saya tidak melewatkan poin penilaian.

US-003: Sebagai Auditor, saya ingin menyelesaikan inspeksi dan sistem otomatis
        membuat issue untuk setiap uraian yang dinilai NG.

US-004: Sebagai Auditee (PIC), saya ingin menerima notifikasi real-time
        ketika ada issue baru yang di-assign ke saya.

US-005: Sebagai Auditee, saya ingin mengupload foto bukti perbaikan
        sebagai bukti tindak lanjut terhadap issue yang di-assign.

US-006: Sebagai Admin, saya ingin mengatur permission setiap role
        dari UI tanpa perlu mengubah kode program.

US-007: Sebagai Admin, saya ingin membuat API Key untuk integrasi Power BI
        agar tim BI dapat menarik data audit secara otomatis.

US-008: Sebagai Manajemen, saya ingin melihat dashboard statistik
        jumlah issue per status dan skor rata-rata per area.
```

---

## 7. Out of Scope (v1.0)

- Mobile native application (iOS/Android) — saat ini web responsive + PWA
- Integrasi ERP/SAP untuk sinkronisasi Work Order
- Otomatisasi eskalasi issue berdasarkan due date (planned v1.1)
- Multi-tenant / multi-perusahaan

---

## 8. Acceptance Criteria

| Feature | Acceptance Criteria |
|:---|:---|
| Login | User dapat login dengan username/password valid dan mendapat JWT token |
| Inspeksi | Auditor dapat membuat, mengisi, dan menyelesaikan sesi inspeksi; issue terbentuk otomatis |
| Issue Follow-Up | PIC dapat mengupload foto dan status issue berubah menjadi PendingValidation |
| API Key | Admin membuat key, gunakan key di endpoint `/public/powerbi/data`, data kembali 200 OK |
| RBAC | User dengan role Auditee tidak bisa mengakses endpoint dengan permission CREATE untuk modul Master |
