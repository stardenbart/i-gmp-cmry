#!/bin/sh
set -eu

backup_interval="${BACKUP_INTERVAL_SECONDS:-86400}"
retention_days="${BACKUP_RETENTION_DAYS:-14}"

case "$backup_interval" in
  *[!0-9]*|'') echo "BACKUP_INTERVAL_SECONDS must be a positive integer" >&2; exit 1 ;;
esac
case "$retention_days" in
  *[!0-9]*|'') echo "BACKUP_RETENTION_DAYS must be a positive integer" >&2; exit 1 ;;
esac
if [ "$backup_interval" -le 0 ] || [ "$retention_days" -le 0 ]; then
  echo "backup interval and retention must be greater than zero" >&2
  exit 1
fi

mkdir -p /backups

while true; do
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  temporary="/backups/.monitoring_audit_${timestamp}.dump.tmp"
  destination="/backups/monitoring_audit_${timestamp}.dump"

  if pg_dump \
    --host="${DB_HOST:-db}" \
    --port="${DB_PORT:-5432}" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --format=custom \
    --no-password \
    --file="$temporary"; then
    mv "$temporary" "$destination"
    sha256sum "$destination" > "${destination}.sha256"
    find /backups -type f \( -name '*.dump' -o -name '*.dump.sha256' \) \
      -mtime "+$retention_days" -delete
    echo "backup completed: $(basename "$destination")"
    next_run_in="$backup_interval"
  else
    rm -f "$temporary"
    echo "backup failed; retrying in 60 seconds" >&2
    next_run_in=60
  fi

  sleep "$next_run_in"
done
