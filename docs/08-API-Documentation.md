# API Documentation

## Introduction
**Base URL:** `http://localhost:8080/api/v1`  
**Authentication:** JWT Bearer token via `Authorization: Bearer <token>` header (unless noted as Public or API Key).

## Standard Response Format
All API responses follow a standard structure:

**Success Response:**
```json
{
  "status": "success",
  "message": "Operation successful",
  "data": { ... }
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "Error message description",
  "code": 400
}
```

## Pagination
For endpoints returning lists, standard query parameters are used:
- `page` (integer): Page number, defaults to 1.
- `limit` (integer): Number of items per page, defaults to 10.
- `search` (string): Search keywords.

---

## 1. AUTH MODULE (`/auth`)

### POST /auth/login
- **Description:** Login with username and password.
- **Authentication:** Public
- **Request Body:** `{ "username": "string", "password": "password123" }`
- **Response:** `{ "token": "jwt-token-string", "user": { "userID": 1, "username": "admin", "fullName": "Admin", "roleID": 1, "departmentID": 1 } }`
- **Errors:** 400 Bad Request, 401 Unauthorized

### POST /auth/logout
- **Description:** Logout current user.
- **Authentication:** JWT
- **Response:** `{ "message": "Logged out successfully" }`
- **Errors:** 401 Unauthorized

### GET /auth/me
- **Description:** Get current logged in user information.
- **Authentication:** JWT
- **Response:** User data object.
- **Errors:** 401 Unauthorized

### POST /auth/forgot-password
- **Description:** Send reset OTP to email.
- **Authentication:** Public
- **Request Body:** `{ "email": "user@example.com" }`
- **Response:** Success message.
- **Errors:** 400 Bad Request, 404 Not Found

### POST /auth/reset-password
- **Description:** Reset password using OTP token.
- **Authentication:** Public
- **Request Body:** `{ "token": "otp-string", "newPassword": "newPassword123" }`
- **Response:** Success message.
- **Errors:** 400 Bad Request

---

## 2. USERS MODULE (`/users`)

### GET /users
- **Description:** List all users (paged, searchable).
- **Authentication:** JWT
- **Permission:** MOD-USR READ
- **Query:** `page`, `limit`, `search`
- **Errors:** 401 Unauthorized, 403 Forbidden

### GET /users/filter
- **Description:** Get filtered users without pagination for dropdowns.
- **Authentication:** JWT
- **Permission:** MOD-USR READ

### POST /users
- **Description:** Create a new user.
- **Authentication:** JWT
- **Permission:** MOD-USR CREATE
- **Request Body:** `{ "departmentID": 1, "roleID": 2, "username": "johnd", "fullName": "John Doe", "email": "john@email.com", "password": "password" }`

### GET /users/:id
- **Description:** Get specific user details by ID.
- **Authentication:** JWT
- **Permission:** MOD-USR READ

### PUT /users/:id
- **Description:** Update user information.
- **Authentication:** JWT
- **Permission:** MOD-USR UPDATE

### DELETE /users/:id
- **Description:** Delete or deactivate a user.
- **Authentication:** JWT
- **Permission:** MOD-USR DELETE

### PUT /users/:id/change-password
- **Description:** Change own password.
- **Authentication:** JWT
- **Request Body:** `{ "oldPassword": "...", "newPassword": "..." }`

### PUT /users/:id/reset-password
- **Description:** Admin reset user password.
- **Authentication:** JWT
- **Permission:** MOD-USR UPDATE
- **Request Body:** `{ "newPassword": "..." }`

### GET /users/:id/permissions
- **Description:** Get specific permission overrides for a user.
- **Authentication:** JWT
- **Permission:** MOD-USR READ

### PUT /users/:id/permissions
- **Description:** Set permission overrides for a user.
- **Authentication:** JWT
- **Permission:** MOD-USR UPDATE
- **Request Body:** `{ "userID": 1, "permissions": [ { "permissionID": 2, "isAllowed": true } ] }`

---

## 3. MASTER DATA MODULE (`/master`)

### GET /master/departments
- **Description:** List departments.
- **Authentication:** JWT
- **Permission:** MOD-MSTR READ

### POST /master/departments
- **Description:** Create department.
- **Authentication:** JWT
- **Permission:** MOD-MSTR CREATE

### PUT /master/departments/:id
- **Description:** Update department.
- **Authentication:** JWT
- **Permission:** MOD-MSTR UPDATE

### DELETE /master/departments/:id
- **Description:** Delete department.
- **Authentication:** JWT
- **Permission:** MOD-MSTR DELETE

### GET /master/roles
- **Description:** List roles.
- **Authentication:** JWT
- **Permission:** MOD-MSTR READ

### POST /master/roles
- **Description:** Create a new role.
- **Authentication:** JWT
- **Permission:** MOD-MSTR CREATE

### PUT /master/roles/:id/permissions
- **Description:** Bulk set permissions for a role.
- **Authentication:** JWT
- **Permission:** MOD-MSTR UPDATE
- **Request Body:** `{ "roleID": 1, "permissions": [ { "permissionID": 1, "isAllowed": true } ] }`

### GET /master/area
- **Description:** List areas.
- **Permission:** MOD-MSTR READ

### POST /master/area
- **Description:** Create area.
- **Permission:** MOD-MSTR CREATE

### PUT /master/area/:id
- **Description:** Update area.
- **Permission:** MOD-MSTR UPDATE

### DELETE /master/area/:id
- **Description:** Delete area.
- **Permission:** MOD-MSTR DELETE

### GET /master/kawasan
- **Description:** List kawasan.

### POST /master/kawasan
- **Description:** Create kawasan.

### GET /master/detail-kawasan
- **Description:** List detail kawasan.

### POST /master/detail-kawasan
- **Description:** Create detail kawasan.

### GET /master/aspek
- **Description:** List aspek.

### POST /master/aspek
- **Description:** Create aspek.

### GET /master/details
- **Description:** List detail.

### POST /master/details
- **Description:** Create detail.

### GET /master/urain
- **Description:** List uraian.

### POST /master/urain
- **Description:** Create uraian.

### GET /master/settings
- **Description:** Get system settings.
- **Authentication:** JWT

### PUT /master/settings
- **Description:** Update system settings.
- **Authentication:** JWT
- **Permission:** Admin

---

## 4. PIC MAPPING MODULE (`/pic-mappings`)

### GET /pic-mappings
- **Description:** List PIC mappings.
- **Permission:** MOD-PIC READ

### POST /pic-mappings
- **Description:** Create PIC mapping.
- **Permission:** MOD-PIC CREATE

### PUT /pic-mappings/:id
- **Description:** Update PIC mapping.
- **Permission:** MOD-PIC UPDATE

### DELETE /pic-mappings/:id
- **Description:** Delete PIC mapping.
- **Permission:** MOD-PIC DELETE

---

## 5. INSPECTION MODULE (`/inspections`)

### GET /inspections
- **Description:** List inspections (paged, filter by area/status).
- **Permission:** MOD-INSP READ

### POST /inspections
- **Description:** Create inspection header.
- **Permission:** MOD-INSP CREATE
- **Request Body:** `{ "areaID": 1, "kawasanID": 2, "detailKawasanID": 3 }`

### GET /inspections/:id
- **Description:** Get inspection detail with results.

### PUT /inspections/:id
- **Description:** Update inspection status.
- **Permission:** MOD-INSP UPDATE

### DELETE /inspections/:id
- **Description:** Delete inspection.
- **Permission:** MOD-INSP DELETE

### GET /inspections/:id/results
- **Description:** Get all inspection results.

### PUT /inspections/:id/results/:resultID
- **Description:** Update inspection result.
- **Permission:** MOD-INSP UPDATE
- **Request Body:** `{ "checking": "OK", "nilai": 100, "keterangan": "Good" }`
*(Checking options: OK|NG|NA)*

### POST /inspections/:id/complete
- **Description:** Complete inspection (creates Issues automatically for NG).
- **Permission:** MOD-INSP UPDATE

### POST /inspections/:id/approve
- **Description:** Approve inspection.
- **Permission:** MOD-INSP APPROVE

### GET /inspections/:id/export
- **Description:** Export inspection to Excel.
- **Permission:** MOD-INSP EXPORT

### GET /inspections/filter
- **Description:** Get filtered inspections without pagination.

### GET /inspections/checklist
- **Description:** Get checklist items based on Area/Kawasan.

---

## 6. ISSUE MODULE (`/issues`)

### GET /issues
- **Description:** List issues (paged, filter by status/area/PIC).
- **Permission:** MOD-ISS READ

### POST /issues
- **Description:** Create manual issue.
- **Permission:** MOD-ISS CREATE
- **Request Body:** `{ "resultID": 12, "issuePICUserID": 3, "dueDate": "2026-12-31", "label": "Major", "keterangan": "Issue detail" }`

### GET /issues/:id
- **Description:** Get issue detail with photos.

### PUT /issues/:id
- **Description:** Update issue details.
- **Permission:** MOD-ISS UPDATE
- **Request Body:** `{ "issueStatus": "InProgress", "label": "Minor", "keterangan": "Updated info", "dueDate": "2026-12-31", "needsWOWR": true, "WO_ID": "WO-123", "WR_ID": "WR-456", "wowrStatus": "Open" }`

### DELETE /issues/:id
- **Description:** Delete issue.
- **Permission:** Admin

### POST /issues/:id/photos
- **Description:** Upload issue photo (multipart).
- **Permission:** MOD-ISS UPDATE
- **Form fields:** `photo` (file), `photoType` (Initial|FollowUp|WOWR), `followUpDate`, `jumlahFollowUp`

### GET /issues/:id/photos
- **Description:** Get issue photos.

### POST /issues/:id/delegate
- **Description:** Delegate issue to another user.
- **Request Body:** `{ "delegateUserID": 5 }`

### GET /issues/filter
- **Description:** Get filtered issues for dropdown/reports.

### GET /issues/followup/filter
- **Description:** Get filtered follow-ups.

---

## 7. NOTIFICATIONS MODULE (`/notifications`)

### GET /notifications
- **Description:** List user notifications (paged).
- **Query:** `page`, `limit`

### PUT /notifications/:id/read
- **Description:** Mark specific notification as read.

### PUT /notifications/read-all
- **Description:** Mark all notifications as read.

---

## 8. LOGS MODULE (`/logs`)

### GET /logs/activity
- **Description:** List activity logs (search via OpenSearch).
- **Permission:** MOD-LOG READ
- **Query:** `page`, `limit`, `search`, `userID`, `dateFrom`, `dateTo`

### GET /logs/login
- **Description:** List login logs.
- **Permission:** MOD-LOG READ

---

## 9. SEARCH MODULE (`/search`)

### GET /search
- **Description:** Global search via OpenSearch.
- **Query:** `q` (search query), `type` (activity|inspection|issue)

---

## 10. DASHBOARD MODULE (`/dashboard`)

### GET /dashboard/stats
- **Description:** Get aggregate statistics for dashboard.
- **Response:** `{ "totalInspections": 10, "totalIssues": 5, "issuesByStatus": {}, "scoreByArea": {}, "recentActivity": [] }`

### GET /dashboard/preview-export
- **Description:** Preview data to be exported.

### GET /dashboard/export
- **Description:** Export dashboard data to Excel.

### GET /dashboard/auditor-detail
- **Description:** Detail performance per Auditor.

### GET /dashboard/pic-detail
- **Description:** Detail performance per PIC.

---

## 11. SSE MODULE (`/sse`)

### GET /sse/notifications
- **Description:** SSE stream for real-time notifications.
- **Header:** `Authorization: Bearer <token>`
- **Response:** `text/event-stream`

---

## 12. API KEY MODULE (`/api-keys`) - Admin only

### GET /api-keys
- **Description:** List API keys created by current admin.

### POST /api-keys
- **Description:** Create new API key.
- **Request Body:** `{ "name": "string", "isSingleUse": boolean }`
- **Response:** `{ "keyID": "uuid", "name": "PowerBI", "rawToken": "MAKEY_...", "prefix": "MAKEY_", "isSingleUse": false, "createdAt": "..." }` *(rawToken shown ONCE)*

### DELETE /api-keys/:id
- **Description:** Revoke API key.

---

## 13. PUBLIC MODULE (`/public`) - API Key Auth

### GET /public/powerbi/data
- **Description:** Get massive data dump for Power BI reporting.
- **Authentication:** `Authorization: Bearer MAKEY_...` OR `?api_key=MAKEY_...`
- **Query:** `since=2026-01-01T00:00:00Z` (RFC3339 format)
- **Response:** JSON with arrays: `inspection_headers`, `inspection_results`, `issues`, `follow_ups`, `issue_photos`.

---

## 14. UPLOAD MODULE (`/uploads`)

### POST /uploads
- **Description:** Upload file (chunked support).
- **Form Data:** `file` (multipart), `inspection_id`, `file_type`

### GET /uploads/:id
- **Description:** Get upload info / metadata.
