#!/bin/sh
set -eu

backup_interval_hours="${BACKUP_INTERVAL_HOURS:-24}"
retention_days="${BACKUP_RETENTION_DAYS:-14}"
backup_dir="${BACKUP_DIR:-/backups}"
uploads_dir="${UPLOADS_DIR:-/uploads}"
work_dir=""

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
if [ ! -d "$uploads_dir" ]; then
  echo "Uploads directory does not exist: $uploads_dir" >&2
  exit 1
fi

mkdir -p "$backup_dir"

cleanup_work_dir() {
  case "$work_dir" in
    "$backup_dir"/.lms-*) rm -rf "$work_dir" ;;
  esac
}

trap cleanup_work_dir EXIT
trap 'exit 130' INT
trap 'exit 143' HUP TERM

sha256() {
  sha256sum "$1" | awk '{print $1}'
}

write_manifest() {
  database_checksum="$(sha256 "$work_dir/database.dump.gz")"
  uploads_checksum="$(sha256 "$work_dir/uploads.tar.gz")"
  cat > "$work_dir/manifest.json" <<EOF
{
  "version": 1,
  "backup_id": "$backup_id",
  "created_at": "$created_at",
  "database": { "file": "database.dump.gz", "format": "postgresql-custom-gzip", "sha256": "$database_checksum" },
  "uploads": { "file": "uploads.tar.gz", "format": "tar-gzip", "sha256": "$uploads_checksum" }
}
EOF
}

is_complete_bundle() {
  bundle="$1"
  [ -f "$bundle/manifest.json" ] && [ -f "$bundle/database.dump.gz" ] && [ -f "$bundle/uploads.tar.gz" ]
}

remove_expired_bundles() {
  find "$backup_dir" -mindepth 1 -maxdepth 1 -type d -name 'lms-*' -mtime "+$retention_days" -print | while IFS= read -r bundle; do
    if is_complete_bundle "$bundle"; then
      rm -rf "$bundle"
    fi
  done
  # Legacy releases stored archives directly in /backups. Retain those known
  # formats under the same policy without touching unrelated files.
  find "$backup_dir" -mindepth 1 -maxdepth 1 -type f \( -name 'lms-*.sql.gz' -o -name 'lms-*.dump.gz' -o -name 'lms-*-uploads.tar.gz' \) -mtime "+$retention_days" -delete
}

create_backup() {
  created_at="$(date -u +%FT%TZ)"
  stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  work_dir="$(mktemp -d "$backup_dir/.lms-${stamp}.XXXXXX")"
  suffix="${work_dir##*.lms-${stamp}.}"
  backup_id="${stamp}-${suffix}"
  destination="$backup_dir/lms-${backup_id}"

  echo "[$created_at] creating database and uploads backup $backup_id"
  # POSIX /bin/sh does not reliably propagate a pg_dump failure from a pipe.
  # A failed dump therefore cannot produce a publishable compressed archive.
  pg_dump --format=custom --no-owner --no-privileges > "$work_dir/database.dump"
  gzip -c "$work_dir/database.dump" > "$work_dir/database.dump.gz"
  rm -f "$work_dir/database.dump"
  tar -C "$uploads_dir" -czf "$work_dir/uploads.tar.gz" .
  write_manifest

  # The directory rename is atomic on one filesystem. manifest.json is written
  # last, so only a fully prepared database/uploads pair is ever published.
  mv "$work_dir" "$destination"
  work_dir=""
  remove_expired_bundles
  echo "[$(date -u +%FT%TZ)] backup complete: $destination"
}

while true; do
  create_backup
  if [ "${BACKUP_ONCE:-false}" = "true" ]; then
    exit 0
  fi
  sleep "$((backup_interval_hours * 3600))"
done
