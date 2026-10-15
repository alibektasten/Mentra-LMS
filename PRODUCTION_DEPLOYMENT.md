# Production deployment

This guide deploys Mentra LMS behind HTTPS at a real domain. It is separate from `DEMO_DEPLOYMENT.md`: do not use the demo compose file for live learner data.

## Before deployment

1. Provision a Linux server with at least 2 vCPU, 4 GB RAM and persistent disk storage.
2. Create an `A`/`AAAA` DNS record such as `lms.example.edu` pointing to the server.
3. In the firewall, allow only TCP 22 (restricted to administrators), 80 and 443. Do not expose 3000, 5432 or 8080.
4. Install Docker Engine and Docker Compose on the server, then clone the repository.
5. Copy the template and set strong unique secrets:

```bash
cp .env.production.example .env.production
chmod 600 .env.production
nano .env.production
```

Generate `POSTGRES_PASSWORD` and `JWT_SECRET` independently with `openssl rand -hex 32`. Keep `.env.production` only on the server.

If Google sign-in is enabled, register exactly this callback URL in Google Cloud:

```text
https://YOUR_DOMAIN/api/v1/auth/google/callback
```

## Start and update

```bash
docker compose --env-file .env.production -f docker-compose.production.yml up -d --build
docker compose --env-file .env.production -f docker-compose.production.yml ps
curl -fsS https://YOUR_DOMAIN/api/v1/health
```

Caddy obtains and renews the TLS certificate automatically once DNS resolves to the server. For an update, pull the committed revision, run the same `up -d --build` command, and then use the smoke check below.

## Backups and restore drill

The `backup` service creates a completed backup **bundle** in `data/backups/` every 24 hours by default and removes completed bundles older than 14 days. A bundle contains:

- `database.dump.gz` — a gzip-compressed PostgreSQL custom-format dump;
- `uploads.tar.gz` — uploaded files at backup time;
- `manifest.json` — bundle version, identifier, timestamp, formats and SHA-256 checksums.

Only a directory containing all three files is a completed backup. The worker writes the database and uploads archive in a hidden temporary directory, creates the manifest last, and atomically renames the complete directory. Do not restore from loose legacy archives or a directory without `manifest.json` unless you have separately verified it.

This protects from application-level data loss, but a backup stored on the same server does **not** protect against loss of the server. Copy completed bundle directories to encrypted off-site storage daily before considering the installation production-ready. Off-site copying is not configured by this repository.

Verify that a backup exists:

```bash
docker compose --env-file .env.production -f docker-compose.production.yml logs --tail=50 backup
ls -lh data/backups/
```

### Restore into an isolated database and directory

Always restore into a **new database** and **new uploads directory** first. Do not run restore commands against the live `lms` database.

Choose only a completed bundle, then verify its checksums before restoring:

```bash
BUNDLE=data/backups/lms-YYYYMMDDTHHMMSSZ-RANDOM
cd "$BUNDLE"
sha256sum database.dump.gz uploads.tar.gz
# Compare both output hashes with manifest.json.
```

Create a disposable destination database and restore into it. The following example assumes the production Compose stack is running, but uses a separate database name:

```bash
docker compose --env-file .env.production -f docker-compose.production.yml exec -T postgres \
  createdb -U lms lms_restore_check

gunzip -c "$BUNDLE/database.dump.gz" | \
  docker compose --env-file .env.production -f docker-compose.production.yml exec -T postgres \
  pg_restore --exit-on-error --no-owner -U lms -d lms_restore_check

mkdir -p /srv/mentra-restore-check/uploads
tar -xzf "$BUNDLE/uploads.tar.gz" -C /srv/mentra-restore-check/uploads
```

Validate row counts, constraints, identity/sequence behaviour and expected uploads in the isolated destination. Confirm ownership and read permissions for the restored uploads match the account that runs the backend before switching an application to them. Only after a documented review should an operator plan a controlled cutover; this guide intentionally contains no command that clears or overwrites the live database.

Run this restore drill at least once before launch and after material database changes.

### Consistency limitation

The database dump and uploads archive are created sequentially, not as one storage snapshot. A complete bundle proves both archives finished and passed checksum generation; it does not guarantee they represent the identical instant if a file upload or deletion occurs between `pg_dump` and `tar`. For a strictly consistent recovery point, stop application writes using an operational maintenance window, wait for active uploads to finish, run one backup cycle, then resume writes. This repository does not enable maintenance mode automatically.

Retention deletes only completed bundle directories with the `lms-*` name and known legacy archive filenames. It does not delete unrelated files in `data/backups/`.

## Smoke checks, monitoring, and alerts

From a machine that can reach the domain:

```bash
cd frontend
E2E_BASE_URL=https://YOUR_DOMAIN \
E2E_API_URL=https://YOUR_DOMAIN/api/v1 \
npm run test:e2e
```

Configure an external uptime monitor (for example Better Uptime, UptimeRobot, or your cloud provider) to request:

```text
https://YOUR_DOMAIN/api/v1/health
```

Alert the operations mailbox when it fails for two consecutive checks. Also monitor disk capacity on the server: uploads and local backups share the server disk until object storage is enabled.

## S3/R2 migration

The current deployment deliberately keeps uploads on a local mounted volume for a simple, recoverable first production release. Before serving significant real usage, move `data/uploads/` and `data/backups/` to an S3-compatible bucket (Cloudflare R2, AWS S3, Backblaze B2, or equivalent), use a private bucket, and grant the app a least-privilege access key. The next implementation step is to make uploads use a storage-provider interface and generate time-limited download URLs; do not expose a bucket as a public file server.
