# Validation Rules Documentation
# Sistem Audit Internal Perusahaan — Cimory Audit System

**Versi:** 1.0.0  
**Tanggal:** 2026-07-27  
**Status:** Production-Ready  
**Otoritas:** Tim Engineering & Quality Assurance Cimory  

---

## Executive Summary

Dokumen ini memuat spesifikasi lengkap aturan validasi (*validation rules*) untuk seluruh DTO (Data Transfer Object) API Request, parameter HTTP, serta respons standar pada Sistem Audit Internal Cimory. Setiap DTO divalidasi pada lapisan middleware/handler menggunakan validator Go (`go-playground/validator/v10`) sebelum data masuk ke lapisan logika bisnis (*UseCase*).

Aturan validasi disajikan dalam bentuk tabel terstruktur dengan format:
`| Field | Type | Rules | Error Message |`

---

## 1. Auth Module

Modul untuk manajemen autentikasi user, pemulihan password, dan pergantian password.

### 1.1 `LoginRequest`
Request payload untuk autentikasi user ke sistem.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `username` | String | `required` | "Username wajib diisi" |
| `password` | String | `required,min=6` | "Password wajib diisi dan minimal 6 karakter" |

### 1.2 `ForgotPasswordRequest`
Request payload untuk pengajuan OTP pemulihan password via email.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `email` | String | `required,email` | "Email wajib diisi dan harus berupa format email yang valid" |

### 1.3 `ChangePasswordRequest`
Request payload untuk penggantian password mandiri oleh user terautentikasi.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `old_password` | String | `required` | "Password lama wajib diisi" |
| `new_password` | String | `required,min=8` | "Password baru wajib diisi dan minimal 8 karakter" |

### 1.4 `AdminResetPasswordRequest`
Request payload untuk reset password user oleh Administrator.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `new_password` | String | `required,min=8` | "Password baru wajib diisi dan minimal 8 karakter" |

---

## 2. User Module

Modul pengelolaan akun dan status pengguna aplikasi.

### 2.1 `CreateUserRequest`
Request payload untuk pendaftaran akun user baru oleh Admin.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `department_id` | String | `required` | "Department ID wajib diisi" |
| `role_id` | String | `required` | "Role ID wajib diisi" |
| `username` | String | `required,min=3,max=50` | "Username wajib diisi (3 - 50 karakter)" |
| `full_name` | String | `required` | "Nama lengkap wajib diisi" |
| `email` | String | `required,email` | "Email wajib diisi dan harus berupa alamat email valid" |
| `password` | String | `required,min=8` | "Password wajib diisi dan minimal 8 karakter" |

### 2.2 `UpdateUserRequest`
Request payload untuk pembaruan profil atau status user.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `email` | String | `omitempty,email` | "Format email tidak valid" |
| `user_status` | String | `omitempty,oneof=Active Inactive Suspended` | "Status user harus salah satu dari: Active, Inactive, Suspended" |

---

## 3. Permission Module

Modul konfigurasi hak akses berbasis Role dan User Specific Override.

### 3.1 `BulkSetPermissionRequest`
Request payload untuk mengatur sekumpulan permission sekaligus pada suatu Role.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `role_id` | String | `required` | "Role ID wajib diisi" |
| `permissions` | Array | `required,dive` | "Daftar permissions wajib diisi dan valid" |

### 3.2 `PermissionToggle`
Sub-object item permission dalam array `permissions`.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `permission_id` | String | `required` | "Permission ID wajib diisi" |
| `is_allowed` | Boolean | `required` (boolean) | "Status is_allowed wajib bernilai true atau false" |

### 3.3 `BulkSetUserPermissionRequest`
Request payload untuk mengatur override permission spesifik pada suatu User.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `user_id` | String | `required` | "User ID wajib diisi" |
| `permissions` | Array | `required,dive` | "Daftar permissions override wajib diisi dan valid" |

---

## 4. Master Data Module

Modul pengelolaan entitas dasar hirarki audit, lokasi, dan struktur organisasi.

### 4.1 `CreateAreaRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `area_name` | String | `required,max=100` | "Nama area wajib diisi (maksimal 100 karakter)" |

### 4.2 `CreateKawasanRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `area_id` | String | `required` | "Area ID wajib diisi" |
| `kawasan_name` | String | `required,max=100` | "Nama kawasan wajib diisi (maksimal 100 karakter)" |

### 4.3 `CreateDetailKawasanRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `kawasan_id` | String | `required` | "Kawasan ID wajib diisi" |
| `detail_kawasan_name` | String | `required` | "Nama detail kawasan wajib diisi" |

### 4.4 `CreateAspekRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `area_id` | String | `required` | "Area ID wajib diisi" |
| `aspek_name` | String | `required,max=100` | "Nama aspek audit wajib diisi (maksimal 100 karakter)" |

### 4.5 `CreateDetailRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `aspek_id` | String | `required` | "Aspek ID wajib diisi" |
| `detail_name` | String | `required,max=150` | "Nama detail aspek wajib diisi (maksimal 150 karakter)" |

### 4.6 `CreateUraianRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `detail_id` | String | `required` | "Detail ID wajib diisi" |
| `uraian_text` | String | `required,max=500` | "Teks uraian checklist wajib diisi (maksimal 500 karakter)" |
| `standard_score` | Integer | `required,min=0` | "Skor standar wajib diisi dan bernilai non-negatif" |

### 4.7 `CreateDepartmentRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `department_name` | String | `required,max=100` | "Nama departemen wajib diisi (maksimal 100 karakter)" |

---

## 5. Inspection Module

Modul inisiasi dan pengisian lembar kerja inspeksi audit.

### 5.1 `CreateInspectionRequest`
Request payload untuk inisiasi header inspeksi baru.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `area_id` | String | `required` | "Area ID wajib diisi" |
| `kawasan_id` | String | `required` | "Kawasan ID wajib diisi" |
| `detail_kawasan_id` | String | `required` | "Detail Kawasan ID wajib diisi" |

### 5.2 `UpdateResultRequest`
Request payload untuk memperbarui nilai satu item checklist audit.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `checking` | String | `required,oneof=OK NG NA` | "Nilai checking wajib diisi dan harus salah satu dari: OK, NG, NA" |
| `nilai` | Integer | `min=0` | "Nilai skor minimal 0" |
| `keterangan` | String | `max=255` | "Keterangan tambahan maksimal 255 karakter" |

---

## 6. Issue Module

Modul pelaporan, tindakan perbaikan, dan pelimpahan temuan audit.

### 6.1 `CreateIssueRequest`
Request payload untuk membuat temuan issue secara manual.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `result_id` | String | `required` | "Result ID inspeksi wajib diisi" |
| `issue_pic_user_id` | String | `required` | "User ID PIC Penanggungjawab wajib diisi" |
| `due_date` | String / Date | `optional` (format ISO/RFC3339) | "Format due date harus YYYY-MM-DD atau RFC3339" |

### 6.2 `UpdateIssueRequest`
Request payload untuk memperbarui status dan atribut penanganan issue.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `issue_status` | String | `oneof=Open InProgress PendingValidation Closed Verified` | "Status issue harus salah satu dari: Open, InProgress, PendingValidation, Closed, Verified" |
| `label` | String | `optional` | "Format label tidak valid" |
| `keterangan` | String | `optional` | "Keterangan penanganan berlebihan" |
| `needs_wowr` | Boolean | `optional` | "Nilai needs_wowr harus boolean" |
| `wo_id` | String | `optional` | "Work Order ID tidak valid" |
| `wr_id` | String | `optional` | "Work Request ID tidak valid" |
| `wowr_status` | String | `oneof=None PendingValidation Verified Rejected` | "Status WOWR harus salah satu dari: None, PendingValidation, Verified, Rejected" |

### 6.3 `CreateDelegateRequest`
Request payload untuk mendelengasikan penanganan issue ke user lain.

| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `delegate_user_id` | String | `required` | "User ID penerima delegasi wajib diisi" |

---

## 7. API Key Module

Modul pembuatan token akses untuk Power BI / sistem analitik.

### 7.1 `CreateAPIKeyRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `name` | String | `required,max=100` | "Nama API Key wajib diisi (3 - 100 karakter)" |
| `is_single_use` | Boolean | `optional` | "Tipe single use harus berupa boolean" |

---

## 8. PIC Mapping Module

Modul pemetaan penanggung jawab terhadap lokasi kawasan.

### 8.1 `CreatePICMappingRequest`
| Field | Type | Rules | Error Message |
|:---|:---|:---|:---|
| `kawasan_id` | String | `required` | "Kawasan ID wajib diisi" |
| `user_id` | String | `required` | "User ID PIC wajib diisi" |
| `kategori_pic` | String | `optional` | "Kategori PIC tidak valid" |

---

## 9. General HTTP Validation Rules

Aturan validasi umum untuk parameter HTTP URL, Query, dan Filter.

| Parameter Type | Target Field | Validation Rule | Error Action / Response |
|:---|:---|:---|:---|
| **Path Parameter** | ID (`:id`, `:area_id`, etc.) | Non-empty string, valid identifier string | 400 Bad Request ("Invalid ID parameter") |
| **Query Param** | `page` | Integer, `min=1`, Default = `1` | Otomatis dipaksa ke default `1` jika invalid |
| **Query Param** | `limit` | Integer, `min=1, max=100`, Default = `20` | Otomatis dipaksa ke max `100` jika >100 |
| **Query Param** | `search` | String, `max=255` karakter | Truncate / 400 Bad Request jika >255 char |
| **Query Param** | `since` / Date Filters | Format RFC3339 (`2006-01-02T15:04:05Z`) | 400 Bad Request ("Invalid datetime format, use RFC3339") |

---

## 10. File Upload Validation Rules

Validasi keamanan dan batas teknis pengunggahan berkas media.

| Aspect | Constraint | Technical Enforcement | Error Response |
|:---|:---|:---|:---|
| **MIME Type Allowed** | `image/jpeg`, `image/png`, `image/jpg` | Inspected via Content-Type header & magic bytes | 400 Bad Request ("Only JPEG and PNG images are allowed") |
| **Max File Size** | 10 MB (10.485.760 Bytes) | Checked by Multipart Reader & Middleware Limit | 413 Payload Too Large ("File size exceeds limit of 10MB") |
| **Filename Security** | Dilarang mengandung path traversal (`../`, `/`, `\`) | Sanitized by `filepath.Base()` & regex clean | 400 Bad Request ("Invalid filename structure") |
| **Form Field Name** | Harus bernama `'photo'` atau `'file'` | Form-data boundary field check | 400 Bad Request ("Missing upload file field 'photo' or 'file'") |

---

## 11. Standard HTTP Response Formats

Seluruh API pada Sistem Audit Cimory menghasilkan format JSON standar yang konsisten.

### 11.1 Standard Success Response (200 OK / 201 Created)
```json
{
  "success": true,
  "status_code": 200,
  "message": "Data retrieved successfully",
  "data": {
    "id": "INSP-2026-001",
    "status": "Completed"
  }
}
```

### 11.2 Standard Error Response (400 / 401 / 403 / 404 / 500)
```json
{
  "success": false,
  "status_code": 400,
  "message": "Validation failed: password minimal 8 karakter"
}
```

### 11.3 Standard Paginated Response (200 OK)
```json
{
  "success": true,
  "status_code": 200,
  "message": "Success fetching paginated list",
  "data": [
    {
      "area_id": "AREA-001",
      "area_name": "Pengolahan Susu Pasteurisasi"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```
