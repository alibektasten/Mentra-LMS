#!/bin/sh
set -eu

backup_interval_hours="${BACKUP_INTERVAL_HOURS:-24}"
retention_days="${BACKUP_RETENTION_DAYS:-14}"

case "$backup_interval_hours" in
  ''|*[!0-9]*) echo "BACKUP_INTERVAL_HOURS must be a positive integer" >&2; exit 1 ;;
esac

case "$retention_days" in
  ''|*[!0-9]*) echo "BACKUP_RETENTION_DAYS must be a positive integer" >&2; exit 1 ;;
esac

if [ "$backup_interval_hours" -lt 1 ] || [ "$retention_days" -lt 1 ]; then
  echo "Backup interval and retention must be positive" >&2
  exit 1
fi

mkdir -p /backups

while true; do
  stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  temporary="/backups/lms-${stamp}.sql.gz.partial"
  destination="/backups/lms-${stamp}.sql.gz"

  echo "[$(date -u +%FT%TZ)] creating database backup"
  pg_dump --format=custom --no-owner --no-privileges | gzip -c > "$temporary"
  mv "$temporary" "$destination"
  find /backups -type f -name 'lms-*.sql.gz' -mtime "+$retention_days" -delete
  echo "[$(date -u +%FT%TZ)] backup complete: $destination"

  sleep "$((backup_interval_hours * 3600))"
done
