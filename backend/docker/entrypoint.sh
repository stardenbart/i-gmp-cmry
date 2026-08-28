#!/bin/sh
set -eu

# Existing named volumes may have been created by an older root container.
# Repair ownership once per volume, then skip the recursive scan on restarts.
for writable_dir in /app/uploads /app/logs; do
  marker="$writable_dir/.owner-10001"
  if [ ! -f "$marker" ]; then
    chown -R app:app "$writable_dir"
    su-exec app:app touch "$marker"
  fi
done

su-exec app:app ./migrate
exec su-exec app:app "$@"
