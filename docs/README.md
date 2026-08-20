# 📚 Dokumentasi Teknis — Cimory Audit System

Folder ini berisi dokumentasi teknis lengkap untuk **Sistem Audit Internal Perusahaan (Cimory Audit System)**.

---

## 📋 Daftar Dokumen

| No | File | Judul | Deskripsi |
|:---|:-----|:------|:----------|
| 01 | [01-PRD.md](./01-PRD.md) | Product Requirements Document | Kebutuhan produk, user stories, acceptance criteria |
| 02 | [02-ERD.md](./02-ERD.md) | Entity Relationship Diagram | Skema database, relasi tabel, data dictionary visual |
| 03 | [03-Infrastructure.md](./03-Infrastructure.md) | Infrastructure Documentation | Stack teknologi, Docker services, port, volume |
| 04 | [04-Business-Rules.md](./04-Business-Rules.md) | Business Rules | Aturan bisnis per modul (inspeksi, issue, WOWR, API Key) |
| 05 | [05-Validation-Rules.md](./05-Validation-Rules.md) | Validation Rules | Validasi input per DTO/endpoint |
| 06 | [06-Pipeline.md](./06-Pipeline.md) | CI/CD Pipeline | Workflow development, deployment pipeline, GitHub Actions |
| 07 | [07-System-Architecture.md](./07-System-Architecture.md) | System Architecture | Arsitektur sistem, clean architecture, sequence diagram |
| 08 | [08-API-Documentation.md](./08-API-Documentation.md) | API Documentation | Dokumentasi seluruh REST API endpoint |
| 09 | [09-Data-Dictionary.md](./09-Data-Dictionary.md) | Data Dictionary | Kamus data lengkap seluruh 25 tabel database |
| 10 | [10-IAM.md](./10-IAM.md) | Identity & Access Management | JWT auth, RBAC matrix, API Key auth, permission flow |
| 11 | [11-QA-Test-Plan.md](./11-QA-Test-Plan.md) | QA Test Plan & Test Cases | Rencana pengujian dan test case per modul |
| 12 | [12-Logging-Monitoring.md](./12-Logging-Monitoring.md) | Logging & Monitoring Plan | Logging pipeline, OpenSearch, alert rules |
| 13 | [13-Deployment-Rollback.md](./13-Deployment-Rollback.md) | Deployment & Rollback Plan | Prosedur deployment dan rollback production |
| 14 | [14-Disaster-Recovery.md](./14-Disaster-Recovery.md) | Disaster Recovery Plan | RTO/RPO, backup strategy, recovery scenarios |

---

## 🏗️ Arsitektur Singkat

```
Browser/PWA → Next.js (:3000) → Go Fiber API (:8080)
                                      │
              ┌───────────────────────┼────────────────────────┐
              ▼                       ▼                        ▼
        PostgreSQL              MinIO Storage          Apache Kafka
          (:5432)                 (:9000)                (:9092)
                                                           │
                                                    OpenSearch Worker
                                                           │
                                                    OpenSearch (:9200)
```

---

## 🚀 Quick Start

```bash
# Clone dan jalankan dengan Docker Compose
git clone https://github.com/bimoBintang/System-Audit-.git
cd System-Audit-
docker compose up -d --build

# Akses aplikasi
open http://localhost:3000
```

---

## 🔑 Default Credentials

| Role | User ID | Password |
|:-----|:--------|:---------|
| Admin | `USR-ADMIN-001` | `admin123` |
| Auditor | `USR-AUDITOR-001` | `auditor123` |
| Auditee | `USR-AUDITEE-001` | `auditee123` |
