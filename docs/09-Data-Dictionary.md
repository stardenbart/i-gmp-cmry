# Data Dictionary - Cimory Audit System

This document outlines the data structures for all 25 core tables in the database schema.

---

## 1. Department_Master
**Purpose:** Stores master list of departments in the organization.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| DepartmentID | INT | No | Auto Increment | PK | Unique identifier |
| DepartmentName | VARCHAR(100) | No | | Unique | Name of department |
| IsActive | BOOLEAN | No | TRUE | | Soft delete flag |
| CreatedAt | TIMESTAMP | No | CURRENT_TIMESTAMP| | |
| UpdatedAt | TIMESTAMP | No | CURRENT_TIMESTAMP| | |

## 2. Role_Master
**Purpose:** Master list of user roles (e.g., Admin, Auditor, Auditee).
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| RoleID | INT | No | Auto Increment | PK | Unique identifier |
| RoleName | VARCHAR(50) | No | | Unique | Role name |
| IsActive | BOOLEAN | No | TRUE | | Active status |

## 3. Users
**Purpose:** Stores all system users and their credentials.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| UserID | INT | No | Auto Increment | PK | Unique identifier |
| Username | VARCHAR(50) | No | | Unique | Login username |
| PasswordHash | VARCHAR(255) | No | | | Bcrypt hash |
| FullName | VARCHAR(100) | No | | | User's full name |
| Email | VARCHAR(100) | No | | Unique | User email address |
| DepartmentID | INT | No | | FK (Department_Master) | User's department |
| RoleID | INT | No | | FK (Role_Master) | User's primary role |
| IsActive | BOOLEAN | No | TRUE | | |

## 4. Module_Master
**Purpose:** Defines system modules for RBAC.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| ModuleID | INT | No | Auto Increment | PK | |
| ModuleCode | VARCHAR(50) | No | | Unique | e.g. MOD-USR, MOD-INSP |
| ModuleName | VARCHAR(100) | No | | | Human readable name |

## 5. Permission_Master
**Purpose:** List of available actions (CREATE, READ, UPDATE, DELETE) per module.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| PermissionID | INT | No | Auto Increment | PK | |
| ModuleID | INT | No | | FK (Module_Master) | |
| Action | VARCHAR(50) | No | | | e.g. CREATE, READ |

## 6. Role_Permission
**Purpose:** Core RBAC mapping of Permissions to Roles.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| RoleID | INT | No | | PK, FK (Role_Master) | |
| PermissionID | INT | No | | PK, FK (Permission_Master)| |
| IsAllowed | BOOLEAN | No | TRUE | | Is permission granted? |

## 7. User_Permission
**Purpose:** Overrides permissions for specific users (grant/deny).
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| UserID | INT | No | | PK, FK (Users) | |
| PermissionID | INT | No | | PK, FK (Permission_Master)| |
| IsAllowed | BOOLEAN | No | TRUE | | Explicit allow/deny override |

## 8. PIC_Mapping
**Purpose:** Maps a PIC (User) to specific Kawasan for Issue tracking.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| MappingID | INT | No | Auto Increment | PK | |
| UserID | INT | No | | FK (Users) | The PIC |
| KawasanID | INT | No | | FK (Kawasan_Master) | Assigned location |

## 9. Area_Master
**Purpose:** High level physical or logical Area.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| AreaID | INT | No | Auto Increment | PK | |
| AreaName | VARCHAR(100) | No | | | |

## 10. Kawasan_Master
**Purpose:** Sub-area within an Area.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| KawasanID | INT | No | Auto Increment | PK | |
| AreaID | INT | No | | FK (Area_Master) | |
| KawasanName| VARCHAR(100) | No | | | |

## 11. DetailKawasan_Master
**Purpose:** Specific details of a Kawasan.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| DetailKawasanID| INT | No | Auto Increment | PK | |
| KawasanID | INT | No | | FK (Kawasan_Master) | |
| Name | VARCHAR(100) | No | | | |

## 12. Aspek_Master
**Purpose:** Evaluation aspect category.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| AspekID | INT | No | Auto Increment | PK | |
| AreaID | INT | No | | FK (Area_Master) | |
| AspekName | VARCHAR(100) | No | | | |

## 13. Detail_Master
**Purpose:** Granular detail under an Aspek.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| DetailID | INT | No | Auto Increment | PK | |
| AspekID | INT | No | | FK (Aspek_Master) | |
| DetailName | VARCHAR(100) | No | | | |

## 14. Uraian_Master
**Purpose:** Checklist items to evaluate, holding standard score values.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| UraianID | INT | No | Auto Increment | PK | |
| DetailID | INT | No | | FK (Detail_Master) | |
| UraianText | TEXT | No | | | What to check |
| StandardScore| INT | No | 100 | | Default score if OK |

## 15. Inspection_Header
**Purpose:** Main record for a single audit inspection session.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| InspectionID| INT | No | Auto Increment | PK | |
| AuditorID | INT | No | | FK (Users) | User performing audit |
| AreaID | INT | No | | FK (Area_Master) | |
| KawasanID | INT | No | | FK (Kawasan_Master) | |
| Status | VARCHAR(20) | No | 'Draft' | IN (Draft,Completed,Approved) | |
| Date | TIMESTAMP | No | CURRENT_TIMESTAMP| | |

## 16. Inspection_Result
**Purpose:** Individual checklist evaluation result per Inspection.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| ResultID | INT | No | Auto Increment | PK | |
| InspectionID| INT | No | | FK (Inspection_Header)| |
| UraianID | INT | No | | FK (Uraian_Master) | |
| Checking | VARCHAR(5) | Yes | | IN (OK, NG, NA) | |
| Nilai | INT | Yes | | | Actual score |
| Keterangan | TEXT | Yes | | | Notes |

## 17. uploads
**Purpose:** Generic file storage metadata.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| UploadID | INT | No | Auto Increment | PK | |
| FilePath | VARCHAR(255) | No | | | MinIO path |
| FileType | VARCHAR(50) | No | | | |

## 18. Issue
**Purpose:** Finding/Problem tracked after an Inspection NG result.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| IssueID | INT | No | Auto Increment | PK | |
| ResultID | INT | No | | FK (Inspection_Result)| Originating result |
| PICUserID | INT | No | | FK (Users) | Responsible PIC |
| Status | VARCHAR(20) | No | 'Open' | | Open, InProgress, Resolved, etc |
| DueDate | DATE | No | | | |
| Label | VARCHAR(50) | Yes | | | Priority/Category |
| WOWRStatus | VARCHAR(20) | Yes | | | Works Order/Request status |

## 19. Issue_Photo
**Purpose:** Images proving the issue or showing the resolution.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| PhotoID | INT | No | Auto Increment | PK | |
| IssueID | INT | No | | FK (Issue) | |
| PhotoType | VARCHAR(20) | No | | IN (Initial, FollowUp, WOWR) | |
| FilePath | VARCHAR(255) | No | | | MinIO path |
| UploadedBy | INT | No | | FK (Users) | |

## 20. Issue_Delegate
**Purpose:** Tracks delegation of issues from one user to another.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| DelegateID | INT | No | Auto Increment | PK | |
| IssueID | INT | No | | FK (Issue) | |
| FromUserID | INT | No | | FK (Users) | |
| ToUserID | INT | No | | FK (Users) | |
| DelegateDate| TIMESTAMP | No | CURRENT_TIMESTAMP| | |

## 21. Login_Log
**Purpose:** Tracks user authentication events.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| LogID | INT | No | Auto Increment | PK | |
| UserID | INT | No | | FK (Users) | |
| LoginTime | TIMESTAMP | No | CURRENT_TIMESTAMP| | |
| IPAddress | VARCHAR(50) | Yes | | | |
| UserAgent | VARCHAR(255) | Yes | | | |

## 22. Activity_Log
**Purpose:** Global audit trail. Synced to OpenSearch.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| ActivityID | INT | No | Auto Increment | PK | |
| UserID | INT | Yes | | FK (Users) | Null if system action |
| Action | VARCHAR(100) | No | | | |
| TargetType | VARCHAR(50) | No | | | e.g. 'Issue', 'Inspection' |
| TargetID | INT | Yes | | | |

## 23. System_Setting
**Purpose:** Global app configurations.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| SettingKey | VARCHAR(100) | No | | PK | |
| SettingValue| TEXT | Yes | | | Encrypted for sensitive keys |

## 24. API_Key
**Purpose:** Integration keys for tools like Power BI.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| KeyID | UUID | No | uuid_generate_v4()| PK | |
| KeyHash | VARCHAR(255) | No | | | Hashed key |
| Name | VARCHAR(100) | No | | | |
| IsSingleUse | BOOLEAN | No | FALSE | | |

## 25. Notification
**Purpose:** In-app user notifications.
| Column | Type | Nullable | Default | Constraints | Description |
|---|---|---|---|---|---|
| NotifID | INT | No | Auto Increment | PK | |
| UserID | INT | No | | FK (Users) | |
| Title | VARCHAR(100) | No | | | |
| Message | TEXT | No | | | |
| IsRead | BOOLEAN | No | FALSE | | |
| CreatedAt | TIMESTAMP | No | CURRENT_TIMESTAMP| | |
