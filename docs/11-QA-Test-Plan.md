# QA Test Plan - Cimory Audit System

## 1. Test Objectives
- Ensure all API endpoints meet functional and non-functional requirements.
- Validate Role-Based Access Control (RBAC) across all modules.
- Ensure automated Issue creation and handling works flawlessly.
- Guarantee system performance and data security.

## 2. Test Scope
**In-Scope:**
- Authentication & Authorization
- RBAC validation
- Inspection management lifecycle
- Issue tracking & Follow-up workflows
- API Key management & Power BI integration
- SSE Notifications
- Basic Security and Performance

**Out-of-Scope:**
- UI/UX layout testing (Frontend)
- Network infrastructure testing
- Hardware device testing

## 3. Test Types
- **Unit Testing:** Focus on individual logic (e.g., hash generation, score calculation).
- **Integration Testing:** API endpoint behavior with DB.
- **E2E Testing:** Complete user journey (Inspection -> Issue -> Resolve).
- **Performance Testing:** Load testing critical endpoints.
- **Security Testing:** Auth bounds, RBAC, SQL Injection checks.

## 4. Test Environment Setup
- **App Server:** Go Fiber running on `localhost:8080`
- **Database:** PostgreSQL on `localhost:5432`
- **Storage:** MinIO running on `localhost:9000`
- **Search:** OpenSearch on `localhost:9200`
- **Mock Data:** DB seeded with 1 Admin, 1 Auditor, 1 Auditee, basic master data.

---

## 5. Detailed Test Cases

### AUTH MODULE
| TC-ID | Test Name | Module | Type | Precondition | Steps | Expected Result | Priority |
|---|---|---|---|---|---|---|---|
| TC-AUTH-001 | Login dengan kredensial valid | Auth | Positive | User exists, active | POST `/auth/login` with valid uname/pass | HTTP 200 + JWT token returned | High |
| TC-AUTH-002 | Login dengan password salah | Auth | Negative | User exists | POST `/auth/login` with invalid pass | HTTP 401 Unauthorized | High |
| TC-AUTH-003 | Login dengan user Inactive | Auth | Negative | User IsActive=false | POST `/auth/login` | HTTP 401 Unauthorized | High |
| TC-AUTH-004 | Request endpoint protected tanpa JWT | Auth | Negative | None | GET `/users` without Auth header | HTTP 401 Unauthorized | High |
| TC-AUTH-005 | Request dengan JWT expired | Auth | Negative | Have expired token | GET `/users` with expired token | HTTP 401 Unauthorized | High |
| TC-AUTH-006 | Forgot password dengan email valid | Auth | Positive | User email exists | POST `/auth/forgot-password` | HTTP 200 + OTP Email sent | Medium |
| TC-AUTH-007 | Change password dengan old password salah| Auth | Negative | Logged in | PUT `/users/me/change-password` | HTTP 400 Bad Request | Medium |

### RBAC MODULE
| TC-ID | Test Name | Module | Type | Precondition | Steps | Expected Result | Priority |
|---|---|---|---|---|---|---|---|
| TC-RBAC-001 | Auditor akses endpoint DELETE /users/:id | RBAC | Negative | Logged as Auditor | DELETE `/users/2` | HTTP 403 Forbidden | High |
| TC-RBAC-002 | Admin akses semua endpoint | RBAC | Positive | Logged as Admin | Call various CRUD endpoints | HTTP 200 Success | High |
| TC-RBAC-003 | User permission override allow | RBAC | Positive | Override exists | Auditee accessing restricted read | HTTP 200 Success | Medium |
| TC-RBAC-004 | User permission override deny | RBAC | Negative | Override exists | Admin accessing denied endpoint | HTTP 403 Forbidden | High |

### INSPECTION MODULE
| TC-ID | Test Name | Module | Type | Precondition | Steps | Expected Result | Priority |
|---|---|---|---|---|---|---|---|
| TC-INSP-001 | Auditor buat inspeksi baru | Insp | Positive | Auditor login | POST `/inspections` | HTTP 201, inspection_results generated | High |
| TC-INSP-002 | Auditor update result (OK/NG/NA) | Insp | Positive | Inspection exists | PUT `/inspections/1/results/1` | HTTP 200 Success | High |
| TC-INSP-003 | Complete inspection | Insp | Positive | Has NG results | POST `/inspections/1/complete` | Issues automatically created for NG | High |
| TC-INSP-004 | Complete inspection sudah Completed | Insp | Negative | Insp is Completed | POST `/inspections/1/complete` | HTTP 400 Bad Request | Medium |
| TC-INSP-005 | Auditee coba buat inspeksi | Insp | Negative | Auditee login | POST `/inspections` | HTTP 403 Forbidden | High |

### ISSUE MODULE
| TC-ID | Test Name | Module | Type | Precondition | Steps | Expected Result | Priority |
|---|---|---|---|---|---|---|---|
| TC-ISS-001 | PIC upload foto follow-up | Issue | Positive | PIC login, Issue Open| POST `/issues/1/photos` | Status -> PendingValidation | High |
| TC-ISS-002 | Auditor verify issue | Issue | Positive | Status PendingVal | Auditor verifies follow up | Status -> Verified/Resolved | High |
| TC-ISS-003 | PIC update issue yang bukan miliknya | Issue | Negative | PIC login | PUT `/issues/2` | HTTP 403/404 | High |
| TC-ISS-004 | Upload file non-image | Issue | Negative | PIC login | Upload .pdf as photo | HTTP 400 Bad Request | Medium |
| TC-ISS-005 | Upload file > 10MB | Issue | Negative | PIC login | Upload 15MB file | HTTP 400 Payload Too Large | Medium |

### API KEY MODULE
| TC-ID | Test Name | Module | Type | Precondition | Steps | Expected Result | Priority |
|---|---|---|---|---|---|---|---|
| TC-APIKEY-001 | Request /public/powerbi/data dg API key | API | Positive | Key generated | GET `/public/powerbi/data?api_key=...`| HTTP 200 JSON Data | High |
| TC-APIKEY-002 | Request dengan API key invalid | API | Negative | Wrong key | GET `/public/powerbi/data` | HTTP 401 Unauthorized | High |
| TC-APIKEY-003 | Request single-use key yang sudah dipakai | API | Negative | Key used once | GET with single-use key | HTTP 401 Unauthorized | Medium |
| TC-APIKEY-004 | Request dengan revoked key | API | Negative | Key deleted | GET with deleted key | HTTP 401 Unauthorized | High |
| TC-APIKEY-005 | Filter dengan ?since= param | API | Positive | Data exists | GET `...?since=2026-01-01` | Filtered data returned | High |

### NOTIFICATION MODULE
| TC-ID | Test Name | Module | Type | Precondition | Steps | Expected Result | Priority |
|---|---|---|---|---|---|---|---|
| TC-NOTIF-001 | SSE stream aktif dan terima notif | Notif | Positive | SSE Connected | Trigger event (e.g. Issue assign) | Message received on stream | High |
| TC-NOTIF-002 | Mark as read | Notif | Positive | Unread notif exists | PUT `/notifications/1/read` | IsRead changed to true | Medium |

### PERFORMANCE & SECURITY
| TC-ID | Test Name | Module | Type | Precondition | Steps | Expected Result | Priority |
|---|---|---|---|---|---|---|---|
| TC-PERF-001 | Login endpoint response time | Perf | Positive | Load 10 req/s | Measure POST `/auth/login` | Response < 200ms | High |
| TC-PERF-002 | List inspections response | Perf | Positive | 100+ records | Measure GET `/inspections` | Response < 500ms | High |
| TC-PERF-003 | Dashboard stats response | Perf | Positive | Massive data | Measure GET `/dashboard/stats` | Response < 1000ms | High |
| TC-SEC-001 | SQL Injection di search param | Sec | Negative | None | GET `/users?search=' OR 1=1--` | Graceful failure, no breach | High |
| TC-SEC-002 | Password tidak plaintext | Sec | Positive | User exists | Check DB row directly | Password is hashed (Bcrypt) | High |
| TC-SEC-003 | API Key tidak bisa dilihat 2x | Sec | Positive | Key generated | Check API Key list response | Token only shows prefix/hash | High |
| TC-SEC-004 | CORS checking | Sec | Negative | Wrong Origin | Request from untrusted origin | Blocked by CORS | High |
| TC-SEC-005 | Encrypted System Setting | Sec | Positive | SMTP set | Check System_Setting DB | SMTP Pass is encrypted | High |

---

## 6. Test Execution Checklist
- [ ] Database Schema Migrated
- [ ] Seed Data Inserted
- [ ] Environment Variables Configured
- [ ] Run Unit Tests (`go test ./...`)
- [ ] Run API Tests (Postman/Newman)
- [ ] Security Scan (Static Analysis)
- [ ] Load Testing (K6/JMeter)
- [ ] Defect Triage & Resolution

## 7. Bug Severity Classification
- **Critical:** System crash, data loss, security breach, core flow completely blocked.
- **Major:** Core feature broken but has workaround, performance severely degraded.
- **Minor:** Non-critical feature broken, edge cases failing.
- **Trivial:** Typos, minor UI/UX issues, very rare edge cases.

## 8. Test Coverage Targets
- **Code Coverage:** Minimum 80% on core services and utilities.
- **API Coverage:** 100% of defined endpoints must have at least one positive and one negative test.
- **Role Coverage:** 100% matrix test for Admin, Auditor, and Auditee roles across all modules.
