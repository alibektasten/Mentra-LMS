#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
workspace="$(mktemp -d)"
trap 'rm -rf "$workspace"' EXIT HUP INT TERM

mkdir -p "$workspace/bin" "$workspace/backups" "$workspace/uploads"
printf 'student attachment\n' > "$workspace/uploads/file with spaces.txt"
printf 'материал\n' > "$workspace/uploads/қазақша.txt"

write_command() {
  command_name="$1"
  command_body="$2"
  printf '%s\n' '#!/bin/sh' "$command_body" > "$workspace/bin/$command_name"
  chmod +x "$workspace/bin/$command_name"
}

write_command pg_dump 'printf "%s\\n" "$*" > "$PG_DUMP_ARGS_FILE"; printf "Mentra test database dump\\n"'

PATH="$workspace/bin:$PATH" PG_DUMP_ARGS_FILE="$workspace/pg_dump_args" BACKUP_DIR="$workspace/backups" UPLOADS_DIR="$workspace/uploads" BACKUP_ONCE=true sh "$root/docker/backup/backup.sh"
[ "$(cat "$workspace/pg_dump_args")" = '--format=custom --no-owner --no-privileges' ]

database_backup="$(find "$workspace/backups" -name 'database.dump.gz' -type f)"
uploads_backup="$(find "$workspace/backups" -name 'uploads.tar.gz' -type f)"
[ -n "$database_backup" ] && [ -n "$uploads_backup" ]
bundle="$(dirname "$database_backup")"
[ -f "$bundle/manifest.json" ]
gzip -t "$database_backup"
mkdir "$workspace/restored-uploads"
tar -xzf "$uploads_backup" -C "$workspace/restored-uploads"
cmp "$workspace/uploads/file with spaces.txt" "$workspace/restored-uploads/file with spaces.txt"
cmp "$workspace/uploads/қазақша.txt" "$workspace/restored-uploads/қазақша.txt"
database_hash="$(sha256sum "$database_backup" | awk '{print $1}')"
uploads_hash="$(sha256sum "$uploads_backup" | awk '{print $1}')"
grep -q "$database_hash" "$bundle/manifest.json"
grep -q "$uploads_hash" "$bundle/manifest.json"

if PATH="$workspace/bin:$PATH" BACKUP_DIR="$workspace/backups" UPLOADS_DIR="$workspace/missing" BACKUP_ONCE=true sh "$root/docker/backup/backup.sh"; then
  echo 'backup unexpectedly succeeded without uploads directory' >&2
  exit 1
fi

assert_failed_backup_keeps_previous_bundle() {
  label="$1"
  body="$2"
  write_command "$label" "$body"
  if PATH="$workspace/bin:$PATH" PG_DUMP_ARGS_FILE="$workspace/pg_dump_args" BACKUP_DIR="$workspace/backups" UPLOADS_DIR="$workspace/uploads" BACKUP_ONCE=true sh "$root/docker/backup/backup.sh"; then
    echo "backup unexpectedly succeeded with failing $label" >&2
    exit 1
  fi
  [ -f "$bundle/manifest.json" ]
  gzip -t "$database_backup"
  if find "$workspace/backups" \( -name '.lms-*' -o -name '*.partial' \) | grep -q .; then
    echo "temporary artifact remained after failing $label" >&2
    exit 1
  fi
}

assert_failed_backup_keeps_previous_bundle pg_dump 'exit 23'
write_command pg_dump 'printf "Mentra test database dump\\n"'
assert_failed_backup_keeps_previous_bundle gzip 'exit 24'
rm "$workspace/bin/gzip"
write_command tar 'exit 25'
assert_failed_backup_keeps_previous_bundle tar 'exit 25'
rm "$workspace/bin/tar"

# A controlled TERM while uploads are being archived must remove the hidden
# working directory and must not publish a manifest-less bundle.
write_command tar 'sleep 5'
PATH="$workspace/bin:$PATH" BACKUP_DIR="$workspace/backups" UPLOADS_DIR="$workspace/uploads" BACKUP_ONCE=true sh "$root/docker/backup/backup.sh" >"$workspace/interrupted.log" 2>&1 &
backup_pid=$!
attempts=0
while ! find "$workspace/backups" -name '.lms-*' -type d | grep -q .; do
  attempts=$((attempts + 1))
  if [ "$attempts" -gt 30 ]; then
    echo 'backup did not create a temporary directory before interruption test' >&2
    exit 1
  fi
  sleep 1
done
kill -TERM "$backup_pid"
if wait "$backup_pid"; then
  echo 'interrupted backup unexpectedly succeeded' >&2
  exit 1
fi
rm "$workspace/bin/tar"
if find "$workspace/backups" -name '.lms-*' -type d | grep -q .; then
  echo 'temporary bundle remained after interruption' >&2
  exit 1
fi

mkdir "$workspace/backups/lms-incomplete"
touch -t 200001010000 "$workspace/backups/lms-incomplete"
touch -t 200001010000 "$bundle" "$bundle/manifest.json" "$database_backup" "$uploads_backup"
touch "$workspace/backups/lms-legacy.sql.gz" "$workspace/backups/operator-note.txt"
touch -t 200001010000 "$workspace/backups/lms-legacy.sql.gz" "$workspace/backups/operator-note.txt"
write_command pg_dump 'printf "fresh backup\\n"'
PATH="$workspace/bin:$PATH" BACKUP_DIR="$workspace/backups" UPLOADS_DIR="$workspace/uploads" BACKUP_RETENTION_DAYS=1 BACKUP_ONCE=true sh "$root/docker/backup/backup.sh"
[ ! -e "$bundle" ]
[ -d "$workspace/backups/lms-incomplete" ]
[ ! -e "$workspace/backups/lms-legacy.sql.gz" ]
[ -f "$workspace/backups/operator-note.txt" ]

echo 'backup tests passed'
