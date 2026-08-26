# Backend maintenance utilities

Each directory is an independent Go command so the module remains compatible
with `go test ./...`. Run commands from the `backend` directory.

Database utilities read the same `.env`/OS environment as the backend and never
contain embedded credentials:

```bash
go run ./scripts/check-checklist
go run ./scripts/check-results
go run ./scripts/check-uraians
go run ./scripts/test-checklist-api
go run ./scripts/test-db
```

API utilities require `JWT_SECRET` and expect the API at localhost port 8080:

```bash
go run ./scripts/test-perms
go run ./scripts/test-users
```

Spreadsheet generators accept repository-relative output paths:

```bash
go run ./scripts/gen-template -output templates/master_gmp.xlsx
go run ./scripts/generate-template
```
