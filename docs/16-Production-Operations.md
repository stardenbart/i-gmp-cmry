# Production Operations

## Build and start

Use the Git commit as the application image tag so a release can be identified
and rolled back without guessing which `latest` image is running.

```sh
export APP_IMAGE_TAG="$(git rev-parse --short HEAD)"
podman compose -p audit --env-file .env -f docker-compose.prod.yml build
podman compose -p audit --env-file .env -f docker-compose.prod.yml up -d
```

Port `3000` remains published by the main Compose file. It can be restricted to
localhost later, after the HTTPS proxy has been tested.

> **Always pass `-p audit`.** The running stack belongs to the Compose project
> `audit` (network `audit_audit_network`). Without `-p`, Compose names the
> project after the folder (`system-audit`) and puts new containers on
> `system-audit_audit_network`, where the `db`, `redis` and `backend` hostnames
> do not resolve. The backend then fails with
> `lookup db ... no such host` and the site goes down.

## Releasing a new version (backend/frontend)

On the production server `podman compose` delegates to the legacy
`/usr/bin/docker-compose`, which cannot replace a running container: `up -d`
fails with `the container name "audit_prod_backend" is already in use`. Stop
and remove the old container first, then start the new one. Data is safe: the
database, uploads and logs live in named volumes.

Run from `~/System-Audit`:

```sh
# 1. Backup and validate (in addition to the daily automatic dump)
U=$(grep ^DB_USER= .env | cut -d= -f2); D=$(grep ^DB_NAME= .env | cut -d= -f2)
F=backups/pre_deploy_$(date +%Y%m%d_%H%M%S).dump
podman exec audit_prod_db pg_dump -U "$U" -d "$D" -Fc > "$F"
podman exec -i audit_prod_db pg_restore --list < "$F" >/dev/null && echo "backup ok: $F"

# 2. Fetch the code and build images tagged with the commit
git pull --ff-only origin main
export APP_IMAGE_TAG="$(git rev-parse --short HEAD)"
podman compose -p audit --env-file .env -f docker-compose.prod.yml build backend frontend

# 3. Replace the backend (runs SQL migrations on startup), wait for healthy
podman stop audit_prod_backend && podman rm audit_prod_backend
podman compose -p audit --env-file .env -f docker-compose.prod.yml up -d --no-deps backend
until [ "$(podman inspect --format '{{.State.Health.Status}}' audit_prod_backend)" = healthy ]; do sleep 3; done
podman exec audit_prod_frontend wget -qO- http://backend:8080/health

# 4. Replace the frontend
podman stop audit_prod_frontend && podman rm audit_prod_frontend
podman compose -p audit --env-file .env -f docker-compose.prod.yml up -d --no-deps frontend

# 5. Verify
./ops/check-production.sh
podman logs --since 5m audit_prod_backend 2>&1 | grep -E "❌|panic"
```

Expect about a minute of backend downtime between `podman rm` and healthy.
Each step targets a single container. Do not use `podman compose down`, which
also stops the database, Kafka and the other services.

If `podman exec audit_prod_frontend wget ... backend:8080` reports `bad
address` right after the backend restarts, wait a few seconds and retry. The
Podman DNS needs a moment to register the new container.

### Rollback

The previous images stay in local storage under their commit tag
(`podman images | grep system-audit`). To return to a previous release:

```sh
export APP_IMAGE_TAG=<previous commit, e.g. fbe2ed6>
podman stop audit_prod_backend audit_prod_frontend
podman rm audit_prod_backend audit_prod_frontend
podman compose -p audit --env-file .env -f docker-compose.prod.yml up -d --no-deps backend frontend
./ops/check-production.sh
```

The code rollback does not undo migrations that already ran. If a release
includes a schema change that must be reverted, restore it from the
`backups/pre_deploy_*.dump` taken in step 1.

## Automated PostgreSQL backup

The `db-backup` service creates a PostgreSQL custom-format dump immediately at
startup and then every `BACKUP_INTERVAL_SECONDS` (24 hours by default). Each dump
has a SHA-256 checksum. Files older than `BACKUP_RETENTION_DAYS` (14 days by
default) are deleted.

```sh
podman compose -p audit --env-file .env -f docker-compose.prod.yml up -d db-backup
podman logs audit_prod_db_backup
podman volume inspect audit_postgres_backups
```

Validate a backup without restoring it:

```sh
podman exec audit_prod_db_backup sh -c 'latest=$(find /backups -name "*.dump" -type f | sort | tail -1); test -n "$latest"; sha256sum -c "$latest.sha256"; pg_restore --list "$latest" >/dev/null'
```

The named volume protects against an accidental database reset but not total
server or disk loss. Copy encrypted backups to separate storage according to the
disaster-recovery policy.

## Health and disk check

Run this from the repository directory, manually or from a systemd timer:

```sh
./ops/check-production.sh
```

It fails when a required container is stopped, the frontend/backend health check
fails, or repository filesystem usage reaches `DISK_USAGE_THRESHOLD` (85% by
default).

## Optional HTTPS proxy

Set a DNS name that points to the server and a contact email, then start the
optional Caddy overlay. This does not remove the existing port `3000` mapping.

```sh
export APP_DOMAIN=audit.example.com
export TLS_CONTACT_EMAIL=ops@example.com
podman compose -p audit --env-file .env \
  -f docker-compose.prod.yml \
  -f docker-compose.proxy.yml up -d proxy
```

Caddy obtains and renews the TLS certificate automatically. Only enable this
after ports 80 and 443 are reachable and the real production domain is known.
