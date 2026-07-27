# Infrastructure Documentation
# Sistem Audit Internal Perusahaan — Cimory Audit System

**Versi:** 1.0.0  
**Tanggal:** 2026-07-27  

---

## 1. Gambaran Umum Infrastruktur

Sistem berjalan di atas **Docker Compose** multi-container yang menyediakan seluruh ekosistem layanan. Semua container terhubung dalam jaringan bridge internal `audit_network`.

```
┌─────────────────────────────────────────────────────────────────┐
│                    Host Machine / VPS / Cloud VM                │
│                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐   ┌──────────┐  │
│  │ Frontend │    │ Backend  │    │ PostgreSQL│   │  MinIO   │  │
│  │ Next.js  │───▶│ Go Fiber │◀──▶│(Port 5432│   │(Port 9000│  │
│  │:3000     │    │:8080     │    │ internal)│   │ internal)│  │
│  └──────────┘    └────┬─────┘    └──────────┘   └──────────┘  │
│                       │                                         │
│            ┌──────────┼──────────┐                             │
│            ▼          ▼          ▼                             │
│       ┌─────────┐ ┌──────┐ ┌──────────────┐                   │
│       │  Kafka  │ │ Zoo- │ │  OpenSearch  │                   │
│       │:9092    │ │keeper│ │:9200 + :9600 │                   │
│       │internal │ │:2181 │ │              │                   │
│       └─────────┘ └──────┘ └──────────────┘                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Services Detail

### 2.1 PostgreSQL 15 (Database Utama)

| Property | Value |
|:---------|:------|
| **Image** | `postgres:15-alpine` |
| **Container Name** | `monitoring_db` |
| **Port (host:container)** | `5434:5432` |
| **Database** | `monitoring_audit` |
| **User / Password** | `postgres` / `secret` |
| **Volume** | `db_data` (persistent) |
| **Network** | `audit_network` |
| **Healthcheck** | `pg_isready -U postgres` (interval: 5s, retries: 5) |
| **Timezone** | `Asia/Jakarta` |

**Peran:** Menyimpan seluruh data bisnis sistem — user, inspeksi, issue, log, settings, notifikasi, dan API key.

---

### 2.2 MinIO (S3 Compatible Object Storage)

| Property | Value |
|:---------|:------|
| **Image** | `minio/minio` |
| **Container Name** | `monitoring_minio` |
| **Port API** | `9000:9000` |
| **Port Console** | `9001:9001` |
| **Root User / Password** | `minioadmin` / `minioadmin` |
| **Bucket** | `monitoring-audit-bucket` |
| **Volume** | `minio_data` (persistent) |
| **Network** | `audit_network` |
| **Healthcheck** | `curl http://localhost:9000/minio/health/live` |

**Peran:** Menyimpan file/foto yang diupload oleh user — foto bukti temuan issue (Issue_Photo), foto follow-up perbaikan, dan file upload terkait inspeksi.

---

### 2.3 Apache Zookeeper

| Property | Value |
|:---------|:------|
| **Image** | `bitnamilegacy/zookeeper:3.8` |
| **Container Name** | `monitoring_zookeeper` |
| **Port** | `2181:2181` |
| **Network** | `audit_network` |

**Peran:** Koordinator cluster untuk Apache Kafka. Mengelola metadata topic, partisi, dan consumer groups.

---

### 2.4 Apache Kafka 3.4 (Message Broker)

| Property | Value |
|:---------|:------|
| **Image** | `bitnamilegacy/kafka:3.4` |
| **Container Name** | `kafka` |
| **Port External** | `9092:9092` |
| **Port Internal** | `29092` (inter-container) |
| **Volume** | `kafka_data` (persistent) |
| **Network** | `audit_network` |
| **Auto Create Topics** | `true` |
| **Depends On** | `zookeeper` |
| **Healthcheck** | `kafka-topics.sh --bootstrap-server localhost:9092 --list` |

**Peran:** Message broker event-driven untuk Activity Log pipeline. Backend publish event log ke topic Kafka; consumer (OpenSearch Indexer worker) membaca event tersebut dan mengindeksnya ke OpenSearch secara async.

**Topics yang Digunakan:**
| Topic | Producer | Consumer | Tujuan |
|:------|:---------|:---------|:-------|
| `activity-log-events` | Backend API Handler | OpenSearch Indexer | Indexing activity log |

---

### 2.5 OpenSearch 2.11 (Search & Analytics)

| Property | Value |
|:---------|:------|
| **Image** | `opensearchproject/opensearch:2.11.0` |
| **Container Name** | `monitoring_opensearch` |
| **Port API** | `9200:9200` |
| **Port Performance Analyzer** | `9600:9600` |
| **Mode** | `single-node` |
| **Security Plugin** | Disabled |
| **Admin Password** | `mySecurePassword123!` |
| **JVM Heap** | `-Xms512m -Xmx512m` |
| **Volume** | `opensearch_data` (persistent) |
| **Network** | `audit_network` |

**Peran:** Menyimpan dan mengindeks activity log untuk pencarian full-text yang cepat. Admin dapat mencari log berdasarkan kata kunci, user, action, atau rentang waktu.

---

### 2.6 OpenSearch Dashboards

| Property | Value |
|:---------|:------|
| **Image** | `opensearchproject/opensearch-dashboards:2.11.0` |
| **Container Name** | `monitoring_opensearch_dashboards` |
| **Port** | `5601:5601` |
| **OpenSearch URL** | `http://opensearch-node1:9200` |
| **Network** | `audit_network` |

**Peran:** UI visualisasi untuk data di OpenSearch. Dapat digunakan tim DevOps atau Admin untuk membuat dashboard analitik, monitoring log, dan debugging.

---

### 2.7 Backend API (Go Fiber)

| Property | Value |
|:---------|:------|
| **Build Context** | `./backend` |
| **Dockerfile** | `backend/docker/backend.Dockerfile` |
| **Container Name** | `monitoring_backend` |
| **Port** | `8080:8080` |
| **Volumes** | `backend_uploads`, `backend_logs` |
| **Network** | `audit_network` |
| **Depends On** | `db`, `minio`, `kafka`, `opensearch-node1` (semua harus healthy) |

**Environment Variables:**
| Variable | Nilai Default (Docker) |
|:---------|:-----------------------|
| `APP_ENV` | `production` |
| `APP_PORT` | `8080` |
| `DB_HOST` | `db` |
| `DB_PORT` | `5432` |
| `DB_USER` | `postgres` |
| `DB_PASSWORD` | `secret` |
| `DB_NAME` | `monitoring_audit` |
| `MINIO_ENDPOINT` | `minio:9000` |
| `MINIO_ACCESS_KEY` | `minioadmin` |
| `MINIO_SECRET_KEY` | `minioadmin` |
| `MINIO_BUCKET` | `monitoring-audit-bucket` |
| `KAFKA_BROKERS` | `kafka:29092` |
| `OPENSEARCH_URL` | `http://opensearch-node1:9200` |
| `SETTING_ENCRYPTION_KEY` | AES-256 key (base64) |
| `JWT_SECRET` | 64-char hex secret |

---

### 2.8 Frontend (Next.js 15)

| Property | Value |
|:---------|:------|
| **Build Context** | `./frontend` |
| **Dockerfile** | `frontend/docker/frontend.Dockerfile` |
| **Container Name** | `frontend` |
| **Port** | `3000:3000` |
| **Network** | `audit_network` |
| **Depends On** | `backend` |

**Environment Variables:**
| Variable | Nilai |
|:---------|:------|
| `NEXT_PUBLIC_BACKEND_URL` | `http://backend:8080` |
| `INTERNAL_BACKEND_URL` | `http://backend:8080` |

---

## 3. Network Architecture

```
audit_network (bridge)
│
├── monitoring_db          → 172.x.x.2:5432
├── monitoring_minio       → 172.x.x.3:9000, :9001
├── monitoring_zookeeper   → 172.x.x.4:2181
├── kafka                  → 172.x.x.5:9092, :29092
├── monitoring_opensearch  → 172.x.x.6:9200, :9600
├── monitoring_opensearch_dashboards → 172.x.x.7:5601
├── monitoring_backend     → 172.x.x.8:8080
└── frontend               → 172.x.x.9:3000
```

**Port Exposure ke Host:**
| Service | Host Port | Container Port |
|:--------|:----------|:---------------|
| Frontend Web App | `3000` | `3000` |
| Backend API | `8080` | `8080` |
| PostgreSQL | `5434` | `5432` |
| MinIO API | `9000` | `9000` |
| MinIO Console | `9001` | `9001` |
| Kafka | `9092` | `9092` |
| OpenSearch | `9200` | `9200` |
| OpenSearch Dashboards | `5601` | `5601` |
| Zookeeper | `2181` | `2181` |

---

## 4. Volume Persistence

| Volume Name | Service | Path di Container | Data Yang Disimpan |
|:------------|:--------|:------------------|:-------------------|
| `db_data` | PostgreSQL | `/var/lib/postgresql/data` | Seluruh data database |
| `minio_data` | MinIO | `/data` | File/foto yang diupload |
| `kafka_data` | Kafka | `/bitnami/kafka` | Topic data & offset |
| `opensearch_data` | OpenSearch | `/usr/share/opensearch/data` | Index log & analytics |
| `backend_uploads` | Backend | `/app/uploads` | File upload lokal (fallback) |
| `backend_logs` | Backend | `/app/logs` | Log file aplikasi |

---

## 5. Health Check Summary

| Container | Health Command | Interval | Timeout | Retries |
|:----------|:---------------|:---------|:--------|:--------|
| `monitoring_db` | `pg_isready -U postgres` | 5s | 5s | 5 |
| `monitoring_minio` | `curl -f http://localhost:9000/minio/health/live` | 10s | 5s | 5 |
| `kafka` | `kafka-topics.sh --list` | 20s | 10s | 15 |
| `monitoring_opensearch` | `curl -s -I http://localhost:9200` | 10s | 5s | 10 |

> **Note:** Backend container hanya akan start setelah semua dependency di atas berstatus `healthy`.

---

## 6. Technology Stack Summary

| Layer | Teknologi | Versi |
|:------|:----------|:------|
| **Runtime** | Go (Golang) | 1.25+ |
| **Web Framework** | Fiber | v2.52.x |
| **ORM** | GORM | v1.25.x |
| **Database Driver** | pgx (PostgreSQL) | v5.6.x |
| **Authentication** | JWT (golang-jwt) | v5.2.x |
| **Encryption** | AES-256 / bcrypt | stdlib |
| **Object Storage Client** | minio-go | v7.2.x |
| **Message Broker Client** | segmentio/kafka-go | v0.4.x |
| **Search Client** | opensearch-go | v2.3.x |
| **Excel Export** | excelize | v2.11.x |
| **Validation** | go-playground/validator | v10.22.x |
| **Logger** | Uber Zap | v1.27.x |
| **Frontend Framework** | Next.js | 15.x (App Router) |
| **UI Library** | React | 19.x |
| **Language** | TypeScript | 5.x |
| **State Management** | Zustand | latest |
| **HTTP Client** | Axios | latest |
| **Container Runtime** | Docker | 24+ |
| **Orchestration** | Docker Compose | v2 |
| **Database** | PostgreSQL | 15-alpine |
| **Object Storage** | MinIO | latest |
| **Message Broker** | Apache Kafka | 3.4 |
| **Search Engine** | OpenSearch | 2.11.0 |

---

## 7. Minimum Hardware Requirements

### Development / Staging
| Resource | Minimum |
|:---------|:--------|
| CPU | 4 Core |
| RAM | 8 GB |
| Disk | 50 GB SSD |
| Network | 10 Mbps |

### Production
| Resource | Minimum |
|:---------|:--------|
| CPU | 8 Core |
| RAM | 16 GB |
| Disk | 100 GB SSD (scalable) |
| Network | 100 Mbps |

> **Note:** OpenSearch memerlukan setidaknya 4 GB RAM. Sesuaikan `OPENSEARCH_JAVA_OPTS` jika RAM terbatas.
