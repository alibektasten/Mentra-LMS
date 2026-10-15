#!/bin/sh
# Runs an isolated, disposable backup and restore drill. It never connects to
# the application's configured database or uses project upload directories.
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
workspace="$(mktemp -d)"
container_name="mentra-restore-drill-$$"
image_name="mentra-backup-restore-drill"

cleanup() {
  if docker inspect "$container_name" >/dev/null 2>&1; then
    docker stop "$container_name" >/dev/null 2>&1 || true
    docker rm "$container_name" >/dev/null 2>&1 || true
  fi
  find "$workspace" -depth -delete 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM

mkdir -p "$workspace/uploads" "$workspace/backups" "$workspace/restored-uploads"
cp "$root/README.md" "$workspace/uploads/guide with spaces.txt"
cp "$root/docker/backup/backup.sh" "$workspace/uploads/материал.txt"
source_one="$(sha256sum "$workspace/uploads/guide with spaces.txt" | awk '{print $1}')"
source_two="$(sha256sum "$workspace/uploads/материал.txt" | awk '{print $1}')"

docker build -t "$image_name" "$root/docker/backup" >/dev/null
docker run -d --name "$container_name" \
  -e POSTGRES_USER=lms \
  -e POSTGRES_PASSWORD=mentra_restore_test \
  -e POSTGRES_DB=mentra_source \
  pgvector/pgvector:pg16 >/dev/null

attempts=0
until docker exec "$container_name" pg_isready -U lms -d mentra_source >/dev/null 2>&1; do
  attempts=$((attempts + 1))
  [ "$attempts" -lt 30 ] || { echo "Temporary PostgreSQL did not become ready" >&2; exit 1; }
  sleep 1
done

docker exec "$container_name" psql -v ON_ERROR_STOP=1 -U lms -d mentra_source -c "
  CREATE TABLE courses (id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, title text NOT NULL);
  CREATE TABLE materials (id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, course_id bigint NOT NULL REFERENCES courses(id), title text NOT NULL, body text, optional_note text);
  INSERT INTO courses (title) VALUES ('Алгоритмы'), ('Data systems');
  INSERT INTO materials (course_id, title, body, optional_note) VALUES (1, 'Конспект', 'Unicode: қазақша мәтін', NULL), (2, 'Reference', 'linked row', 'optional');
" >/dev/null

docker run --rm --network "container:$container_name" \
  -e PGHOST=127.0.0.1 -e PGPORT=5432 -e PGDATABASE=mentra_source -e PGUSER=lms -e PGPASSWORD=mentra_restore_test \
  -e BACKUP_ONCE=true \
  -v "$workspace/backups:/backups" \
  -v "$workspace/uploads:/uploads:ro" \
  "$image_name" >/dev/null

bundle="$(find "$workspace/backups" -mindepth 1 -maxdepth 1 -type d -name 'lms-*' | head -n 1)"
[ -n "$bundle" ] && [ -f "$bundle/manifest.json" ]
database_hash="$(sha256sum "$bundle/database.dump.gz" | awk '{print $1}')"
uploads_hash="$(sha256sum "$bundle/uploads.tar.gz" | awk '{print $1}')"
grep -q "$database_hash" "$bundle/manifest.json"
grep -q "$uploads_hash" "$bundle/manifest.json"

docker exec "$container_name" createdb -U lms mentra_restore_target
gunzip -c "$bundle/database.dump.gz" | docker exec -i "$container_name" pg_restore --exit-on-error --no-owner -U lms -d mentra_restore_target
tar -xzf "$bundle/uploads.tar.gz" -C "$workspace/restored-uploads"

[ "$source_one" = "$(sha256sum "$workspace/restored-uploads/guide with spaces.txt" | awk '{print $1}')" ]
[ "$source_two" = "$(sha256sum "$workspace/restored-uploads/материал.txt" | awk '{print $1}')" ]
rows="$(docker exec "$container_name" psql -At -U lms -d mentra_restore_target -c "SELECT count(*) FROM materials")"
nulls="$(docker exec "$container_name" psql -At -U lms -d mentra_restore_target -c "SELECT count(*) FROM materials WHERE optional_note IS NULL")"
foreign_keys="$(docker exec "$container_name" psql -At -U lms -d mentra_restore_target -c "SELECT count(*) FROM pg_constraint WHERE conrelid='materials'::regclass AND contype='f'")"
next_identity="$(docker exec "$container_name" psql -qAt -U lms -d mentra_restore_target -c "INSERT INTO courses (title) VALUES ('After restore') RETURNING id")"
[ "$rows" = "2" ] && [ "$nulls" = "1" ] && [ "$foreign_keys" = "1" ] && [ "$next_identity" = "3" ]

printf 'restore drill passed: rows=%s nulls=%s foreign_keys=%s next_identity=%s\n' "$rows" "$nulls" "$foreign_keys" "$next_identity"
