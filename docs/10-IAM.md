# Identity & Access Management (IAM)
# Sistem Audit Internal Perusahaan — Cimory Audit System

**Versi:** 1.0.0  
**Tanggal:** 2026-07-27  

---

## 1. Arsitektur IAM

Sistem menggunakan **dua lapisan autentikasi** dan **RBAC dinamis dua-tingkat** untuk otorisasi:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    REQUEST AUTHENTICATION FLOW                       │
│                                                                      │
│   Incoming Request                                                   │
│         │                                                            │
│         ▼                                                            │
│   ┌──────────┐   API Key?   ┌─────────────────────┐                │
│   │  Auth    │─────YES─────▶│ APIKeyBearerMiddleware│               │
│   │  Check   │              │ (Public Endpoints)   │                │
│   └────┬─────┘              └──────────┬───────────┘               │
│        │ JWT                           │ Valid                      │
│        ▼                               ▼                            │
│   ┌──────────┐              ┌──────────────────┐                   │
│   │ JWT Auth │              │  PowerBI Handler │                   │
│   │Middleware│              └──────────────────┘                   │
│   └────┬─────┘                                                      │
│        │ Authenticated                                              │
│        ▼                                                            │
│   ┌──────────────────────────────────┐                             │
│   │     PERMISSION MIDDLEWARE        │                             │
│   │                                  │                             │
│   │  1. Check User_Permission        │                             │
│   │     (user-level override)        │                             │
│   │         │                        │                             │
│   │         ▼ Not found              │                             │
│   │  2. Check Role_Permission        │                             │
│   │     (role-level baseline)        │                             │
│   │         │                        │                             │
│   │         ▼ Allow/Deny             │                             │
│   └──────────────────────────────────┘                             │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. Roles (Peran Pengguna)

| Role ID | Nama Role | Deskripsi |
|:--------|:----------|:----------|
| `ROLE-001` | **Admin** | Administrator sistem — akses penuh ke seluruh modul |
| `ROLE-002` | **Auditor** | Petugas audit — dapat membuat inspeksi dan melaporkan issue |
| `ROLE-003` | **Auditee** | PIC area yang diaudit — dapat melakukan follow-up issue |
| `ROLE-004` | **Supervisor** | Supervisor — memantau proses audit dan approval |
| `ROLE-005` | **Manager** | Manajer — akses laporan dan dashboard |

---

## 3. Modules (Modul Sistem)

| Module ID | Nama Modul | Deskripsi |
|:----------|:-----------|:----------|
| `MOD-USR` | User Management | Manajemen pengguna sistem |
| `MOD-ROLE` | Role & Permission Management | Pengaturan hak akses |
| `MOD-MSTR` | Master Data | Data master audit (area, aspek, uraian, dll.) |
| `MOD-PIC` | PIC Mapping | Pemetaan Person in Charge per kawasan |
| `MOD-INSP` | Inspection | Modul eksekusi inspeksi/audit |
| `MOD-ISS` | Issue & Follow Up | Manajemen temuan dan tindak lanjut |
| `MOD-LOG` | Logging & Audit Trail | Activity log & login log |

---

## 4. Permission Codes

Setiap modul memiliki kombinasi *permission codes* berikut:

| Code | Nama | Tindakan Yang Dijaga |
|:-----|:-----|:--------------------|
| `CREATE` | Create | POST endpoints — membuat data baru |
| `READ` | Read | GET endpoints — membaca data |
| `UPDATE` | Update | PUT/PATCH endpoints — mengubah data |
| `DELETE` | Delete | DELETE endpoints — menghapus data |
| `APPROVE` | Approve | Approval action (finalisasi inspeksi) |
| `EXPORT` | Export | Download/ekspor data ke Excel |

---

## 5. Default Permission Matrix Per Role

> ✅ = Diizinkan | ❌ = Tidak diizinkan | — = Tidak berlaku

| Modul | Permission | Admin | Auditor | Auditee | Supervisor | Manager |
|:------|:-----------|:-----:|:-------:|:-------:|:----------:|:-------:|
| **MOD-USR** | READ | ✅ | ❌ | ❌ | ✅ | ❌ |
| **MOD-USR** | CREATE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-USR** | UPDATE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-USR** | DELETE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-ROLE** | READ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-ROLE** | UPDATE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-MSTR** | READ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **MOD-MSTR** | CREATE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-MSTR** | UPDATE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-MSTR** | DELETE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-PIC** | READ | ✅ | ✅ | ❌ | ✅ | ❌ |
| **MOD-PIC** | CREATE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-PIC** | UPDATE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-PIC** | DELETE | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MOD-INSP** | READ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **MOD-INSP** | CREATE | ✅ | ✅ | ❌ | ❌ | ❌ |
| **MOD-INSP** | UPDATE | ✅ | ✅ | ❌ | ❌ | ❌ |
| **MOD-INSP** | APPROVE | ✅ | ✅ | ❌ | ✅ | ❌ |
| **MOD-INSP** | EXPORT | ✅ | ✅ | ❌ | ✅ | ✅ |
| **MOD-ISS** | READ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **MOD-ISS** | CREATE | ✅ | ✅ | ❌ | ❌ | ❌ |
| **MOD-ISS** | UPDATE | ✅ | ✅ | ✅ | ❌ | ❌ |
| **MOD-LOG** | READ | ✅ | ❌ | ❌ | ✅ | ✅ |

---

## 6. RBAC Decision Flow (Logic Detail)

```go
// Pseudocode: PermissionMiddleware logic
func CheckAccess(userID, roleID, moduleID, permCode string) bool {
    // Step 1: Check user-level override
    override := User_Permission.CheckOverride(userID, moduleID, permCode)
    
    if override != nil {
        if *override == true {
            return ALLOW
        }
        // Special: if READ denied but UPDATE allowed → still READ allowed
        if permCode == "READ" {
            updateOverride := User_Permission.CheckOverride(userID, moduleID, "UPDATE")
            if updateOverride != nil && *updateOverride == true {
                return ALLOW
            }
        }
        return DENY (403)
    }

    // Step 2: Check role-level permission
    allowed := Role_Permission.HasPermission(roleID, moduleID, permCode)
    
    if allowed {
        return ALLOW
    }
    
    // Special fallback: READ via UPDATE role permission
    if permCode == "READ" {
        allowedViaUpdate := Role_Permission.HasPermission(roleID, moduleID, "UPDATE")
        if allowedViaUpdate {
            return ALLOW
        }
    }
    
    return DENY (403)
}
```

---

## 7. JWT Authentication Detail

### 7.1 Token Structure
```json
{
  "user_id": "USR-ADMIN-001",
  "username": "admin",
  "role_id": "ROLE-001",
  "exp": 1753500000,
  "iat": 1753413600
}
```

### 7.2 Token Generation
- **Algorithm:** HMAC-SHA256 (`HS256`)
- **Secret Key:** Dikonfigurasi via `JWT_SECRET` env variable
- **Expiry:** Dikonfigurasi via `JWT_EXPIRED_HOURS` (default: 24 jam)

### 7.3 Token Extraction
```
Priority Order:
1. Authorization header: "Bearer <token>"
2. Query parameter: "?token=<token>"
```

### 7.4 Protected vs Public Endpoints

| Type | Middleware | Contoh Endpoint |
|:-----|:-----------|:----------------|
| **Public** | Tidak ada | `POST /api/v1/auth/login` |
| **Public (API Key)** | `APIKeyBearerMiddleware` | `GET /api/v1/public/powerbi/data` |
| **Protected (JWT Only)** | `AuthMiddleware` | `GET /api/v1/auth/me` |
| **Protected (JWT + Permission)** | `AuthMiddleware + PermissionMiddleware` | `DELETE /api/v1/users/:id` |
| **Protected (Admin Only)** | `AuthMiddleware + RequireAdmin` | `GET /api/v1/api-keys` |

---

## 8. API Key Authentication (Power BI)

### 8.1 Cara Kerja
1. Admin membuat API Key dari Settings UI
2. Raw token digenerate: `MAKEY_<48-char hex>`
3. SHA-256 hash dari raw token disimpan ke DB (bukan raw token)
4. Raw token ditampilkan **hanya sekali** kepada admin — tidak dapat dilihat lagi
5. Saat request masuk: middleware ekstrak token → hitung SHA-256 → cari di DB

### 8.2 Mode Penggunaan
| Mode | Behavior |
|:-----|:---------|
| **Single-Use** | Setelah satu kali digunakan berhasil, key dinonaktifkan (`IsActive = false`, `UsedAt = now`) |
| **Multi-Use** | Key dapat digunakan berulang kali selama `IsActive = true` dan belum expired |

### 8.3 Token Extraction Priority
```
1. Authorization header: "Bearer MAKEY_..."
2. Query param: ?api_key=MAKEY_...
3. Query param: ?token=MAKEY_...
4. Query param: ?key=MAKEY_...
```

### 8.4 Error Responses
| Kondisi | HTTP Status | Pesan |
|:--------|:------------|:------|
| Token tidak ada | 401 | "API Key is missing" |
| Token tidak ditemukan di DB | 401 | "invalid API key" |
| Key sudah digunakan (single-use) | 401 | "API key has already been used on ..." |
| Key dicabut (revoke) | 401 | "API key has been revoked" |
| Key expired | 401 | "API key has expired" |
| DB error | 401 | "database query error: ..." |

---

## 9. Password Policy

| Aturan | Spesifikasi |
|:-------|:------------|
| **Panjang minimum** | 8 karakter |
| **Hashing algorithm** | bcrypt (cost factor: default) |
| **Reset mekanisme** | Admin reset via endpoint, atau via email OTP (forgot password) |
| **Perubahan password** | User dapat mengubah password sendiri dengan verifikasi old password |

---

## 10. User Account Status

| Status | Deskripsi | Dapat Login? |
|:-------|:----------|:------------|
| `Active` | Akun aktif normal | ✅ Ya |
| `Inactive` | Dinonaktifkan sementara | ❌ Tidak |
| `Suspended` | Ditangguhkan oleh admin | ❌ Tidak |

> **Catatan:** Validasi status akun dilakukan di `AuthUseCase.Login()` sebelum JWT token digenerate.

---

## 11. Audit Trail untuk Akses

Setiap request berhasil dari user terautentikasi dicatat ke `Activity_Log`:

| Field | Value |
|:------|:------|
| `UserID` | ID user dari JWT claims |
| `ActivityAction` | READ / CREATE / UPDATE / DELETE |
| `TableAffected` | Nama entitas dari URL path |
| `IPAddress` | IP address client |
| `ActivityCreatedAt` | Timestamp aksi |
