# Production Operations

## Build and start

Use the Git commit as the application image tag so a release can be identified
and rolled back without guessing which `latest` image is running.

```sh
export APP_IMAGE_TAG="$(git rev-parse --short HEAD)"
podman compose --env-file .env -f docker-compose.prod.yml build
podman compose --env-file .env -f docker-compose.prod.yml up -d
```

Port `3000` remains published by the main Compose file. It can be restricted to
localhost later, after the HTTPS proxy has been tested.

## Automated PostgreSQL backup

The `db-backup` service creates a PostgreSQL custom-format dump immediately at
startup and then every `BACKUP_INTERVAL_SECONDS` (24 hours by default). Each dump
has a SHA-256 checksum. Files older than `BACKUP_RETENTION_DAYS` (14 days by
default) are deleted.

```sh
podman compose --env-file .env -f docker-compose.prod.yml up -d db-backup
podman logs audit_prod_db_backup
podman volume inspect system-audit-_postgres_backups
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
podman compose --env-file .env \
  -f docker-compose.prod.yml \
  -f docker-compose.proxy.yml up -d proxy
```

Caddy obtains and renews the TLS certificate automatically. Only enable this
after ports 80 and 443 are reachable and the real production domain is known.
