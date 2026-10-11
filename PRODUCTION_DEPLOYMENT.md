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

The `backup` service creates a compressed PostgreSQL dump in `data/backups/` every 24 hours by default and deletes copies older than 14 days. This protects from an accidental application-level change, but a backup stored on the same server does **not** protect against loss of the server. Copy that directory to an encrypted off-site bucket daily before considering the installation production-ready.

Verify that a backup exists:

```bash
docker compose --env-file .env.production -f docker-compose.production.yml logs --tail=50 backup
ls -lh data/backups/
```

Test restoration on a disposable server or database, never on the live database:

```bash
gunzip -c data/backups/lms-YYYYMMDDTHHMMSSZ.sql.gz | \
  docker compose --env-file .env.production -f docker-compose.production.yml exec -T postgres \
  pg_restore --clean --if-exists --no-owner -U lms -d lms
```

Run this restore drill at least once before launch and after material database changes.

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
