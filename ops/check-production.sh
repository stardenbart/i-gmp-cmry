#!/bin/sh
set -eu

compose_project="${COMPOSE_PROJECT_NAME:-audit}"
disk_threshold="${DISK_USAGE_THRESHOLD:-85}"

case "$disk_threshold" in
  *[!0-9]*|'') echo "DISK_USAGE_THRESHOLD must be an integer" >&2; exit 1 ;;
esac

disk_usage="$(df -P . | awk 'NR == 2 {gsub(/%/, "", $5); print $5}')"
if [ "$disk_usage" -ge "$disk_threshold" ]; then
  echo "disk usage is ${disk_usage}% (threshold ${disk_threshold}%)" >&2
  exit 1
fi

required_containers="
audit_prod_db
audit_prod_redis
audit_prod_minio
audit_prod_zookeeper
audit_prod_kafka
audit_prod_opensearch
audit_prod_backend
audit_prod_frontend
audit_prod_db_backup
"

for container in $required_containers; do
  status="$(podman inspect --format '{{.State.Status}}' "$container" 2>/dev/null || true)"
  if [ "$status" != "running" ]; then
    echo "$container is not running (project: $compose_project)" >&2
    exit 1
  fi
done

curl -fsS http://127.0.0.1:3000/login >/dev/null
podman exec audit_prod_frontend wget -qO- http://backend:8080/health >/dev/null

echo "production checks passed; disk usage ${disk_usage}%"
