# Mentra LMS

**Mentra** — multi-tenant learning management system for universities and training organisations. It combines course delivery, groups, assessments, analytics and role-based AI learning tools in one workspace.

[Production deployment](PRODUCTION_DEPLOYMENT.md) · [Demo deployment](DEMO_DEPLOYMENT.md) · [API health](http://localhost:8080/health)

## What it includes

- Multi-tenant organisations, scoped JWT sessions and role-based access.
- Separate workspaces for system administrators, organisation administrators, teachers and students.
- Courses, modules, lessons, materials, groups, announcements and gradebooks.
- Assignments, exams, draft auto-save, submissions, manual grading and rubrics.
- AI quiz generation, flashcards, lesson summaries and assignment feedback through OpenAI or local Ollama.
- Learning progress, teacher analytics, notifications and automatically issued completion certificates.
- Personal settings: profile image, display name, interface size, theme and language (English, Russian, Kazakh).
- Localised UI, dark theme, file validation, rate limits, request IDs and AI quotas.

## Technology

| Layer | Stack |
| --- | --- |
| Frontend | Next.js 15, React 19, TypeScript, Tailwind CSS, TanStack Query |
| Backend | Go 1.23, Gin, GORM |
| Data | PostgreSQL 16 with pgvector |
| AI | OpenAI Responses API or Ollama |
| Delivery | Docker Compose, Caddy TLS, GitHub Actions, Playwright |

## Project structure

```text
backend/                 Go API, domain logic, migrations and seed data
frontend/                Next.js application and E2E tests
docker-compose.dev.yml   Local PostgreSQL and Ollama for hot-reload development
docker-compose.demo.yml  IP-based presentation deployment
docker-compose.production.yml
                         HTTPS production deployment with backups
caddy/                   Production reverse-proxy and TLS configuration
docker/backup/           PostgreSQL backup worker
```

## Run locally

### Requirements

- Go 1.23+
- Node.js 22+
- Docker Desktop / Docker Engine with Compose

### 1. Start infrastructure

```bash
docker compose -f docker-compose.dev.yml up -d
```

This starts PostgreSQL on `localhost:5432` and Ollama on `localhost:11434`.

### 2. Configure the API

```bash
cp backend/.env.example backend/.env
```

For the default development database, set at minimum:

```dotenv
SERVER_PORT=8080
SERVER_MODE=debug
CORS_ALLOWED_ORIGINS=http://localhost:3000
DB_DSN=postgres://lms:secret@localhost:5432/lms?sslmode=disable
JWT_SECRET=replace-with-a-long-random-development-secret
AI_PROVIDER=ollama
OLLAMA_BASE_URL=http://localhost:11434
```

To use OpenAI instead, set `AI_PROVIDER=openai`, `OPENAI_API_KEY` and optionally `OPENAI_MODEL`. API keys stay server-side; never use a `NEXT_PUBLIC_` variable for them.

### 3. Start the application

Run each command in a separate terminal:

```bash
cd backend && go run ./cmd/server/main.go
cd frontend && npm install && npm run dev
```

Open [http://localhost:3000](http://localhost:3000). Confirm the backend is ready at [http://localhost:8080/health](http://localhost:8080/health).

### 4. Seed demonstration data

On a fresh local database:

```bash
cd backend && go run ./cmd/seed/main.go
```

The seed command resets application data before creating demonstration organisations, courses, groups and accounts. Do not run it against real data.

## Roles and permissions

| Role | Main capabilities |
| --- | --- |
| System administrator | Organisations, global users and platform access |
| Organisation administrator | Members, groups, courses, settings and organisation analytics |
| Teacher | Assigned courses, materials, groups, assignments, exams, grading, analytics and AI authoring |
| Student | Enrolled courses, lesson progress, submissions, exam attempts, study AI and certificates |

Tenant context is selected by organisation slug (for example `/nu/login`). The API receives `X-Org-Slug`; tokens also contain organisation and role claims. Course and assessment access is checked server-side.

Teachers can open an assignment's **Review submissions** link in the course
builder to view work and save scores and feedback. The student dashboard's
**Next step** card uses visible graded work below 70% to suggest the first
unfinished available lesson in the same course. This is a course-level rule,
not a topic diagnosis or an AI call; it considers the courses loaded on the
dashboard (currently up to 20). Loading, unavailable data and missing evidence
have separate states.

## AI configuration

The provider is selected only by the backend:

```dotenv
# Hosted provider
AI_PROVIDER=openai
OPENAI_API_KEY=your_server_side_key
OPENAI_MODEL=gpt-5.4-mini
AI_DAILY_REQUEST_LIMIT=50

# Or local development provider
AI_PROVIDER=ollama
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=llama3.2
```

Teachers and organisation administrators can create course content with AI. Students can use the study tools only for content they are allowed to access. AI requests are rate-limited and quota-limited per account.

## Quality checks

```bash
# Backend unit tests
(cd backend && go test ./...)

# Frontend type check and optimised build
(cd frontend && npm run type-check && npm run build)
```

Run the isolated setup from the repository root in a fresh shell:

```bash
# Prepare an isolated local E2E stack (never a working or production database)
./scripts/e2e/prepare-isolated.sh
# Copy the export commands printed by the script into this shell, then:
cd frontend && npm run test:e2e:smoke
```

The default Playwright suite verifies health and sign-in for the student,
teacher, organization administrator and super-administrator roles. It requires
explicit `E2E_*` credentials and URLs printed by the isolated setup script; it
does not fall back to a developer's active demo, staging or production session.
The script seeds only a new `mentra-e2e` Docker project. It refuses existing
project containers or volumes (including stopped containers and matching names
without Compose labels), remote Docker endpoints, and failed Docker inventory
checks. Do not rerun preparation to reset a test environment. It never resets
the developer database. Only named test volumes are mounted; the default web
and API ports bind to loopback at `13000` and `18080`.

The isolated seed provides the following test-only accounts. These credentials
are fixtures, not deployment credentials. Sign in at
`http://127.0.0.1:13000/nu/login` for NU or `/kbtu/login` for KBTU; the system
administrator uses `/login`.

| Role / fixture | Organization | Email | Password |
| --- | --- | --- | --- |
| Student | nu | sanzhar@nu.edu.kz | password123 |
| Teacher | nu | baurzhan@nu.edu.kz | password123 |
| Administrator | nu | aiguerim@nu.edu.kz | password123 |
| Other student (access checks) | nu | zhanar@nu.edu.kz | password123 |
| Other teacher (access checks) | nu | damir@nu.edu.kz | password123 |
| Other tenant student | kbtu | nurdaulet@kbtu.edu.kz | password123 |
| System administrator | — | e2e-super-admin@mentra.test | e2e-super-admin-password |

The preparation script prints the complete `E2E_*` exports for these accounts,
the API URL `http://127.0.0.1:18080/api/v1`, locale and timezone. Keep those
exports in the shell used for Playwright. The write suite includes the learning
journey and creates uniquely named course fixtures; it does not require another
seed between runs.

Detailed course UI checks remain explicit and use a fixed locale/timezone:

```bash
E2E_RUN_EXTENDED=true npm run test:e2e:extended
```

Data-changing tests additionally require `E2E_ALLOW_WRITES=true`. Before any
write, the guard verifies local URLs, the running Compose project and services,
published ports, and the frontend image's API build label. Setting
`E2E_ISOLATED=true` alone is not sufficient:

```bash
E2E_ALLOW_WRITES=true npm run test:e2e:write
```

See [frontend/tests/e2e/README.md](frontend/tests/e2e/README.md) for the
scenario matrix and cleanup. Do not treat skipped browser tests as a successful
deployment verification.

Local acceptance on 2026-10-15 used the fresh `mentra-e2e` stack and one seed:

**Follow-up acceptance is blocked:** the later course-progress, strict integer
score and re-auth cancellation changes passed targeted code/type/unit checks,
but their rebuilt image could not be extracted. The host data volume reported
100% capacity (about 97 MiB free), and Docker/containerd returned input/output
errors. The browser results below describe the preceding accepted snapshot,
not these later changes. Existing containers and volumes were preserved; no
prune, cleanup or second seed was performed. Restore disk/Docker storage health
before rebuilding and running the new regressions, including
`npx playwright test browser-session-isolation.spec.ts --project=chromium`.

| Check (run in `frontend` after exporting the setup variables) | Result |
| --- | --- |
| `npm run test:e2e:smoke` | 6 browser tests passed |
| `E2E_ALLOW_WRITES=true npm run test:e2e:write` | 4 passed, 1 skipped, 0 failed |
| `npx playwright test security-session.spec.ts` | 4 adapter tests passed; not browser session tests |
| `npm run type-check` | Passed |
| Production frontend Docker build | Passed |

The write suite covers student submission and teacher grading through the UI,
failed-request draft retention and retry, private submission-file access, hidden
unpublished final grades, and the recommendation card using a real zero score.
A separate browser authoring scenario creates and saves a lesson and assignment
at a 390px viewport, checking labels, persisted values and horizontal overflow.
The legacy lesson-edit test skipped because its first selected course had no
lesson; it is not counted as passed. The main journey creates its initial
module, lesson and assignment through API fixtures.

Limits: browser re-authentication/session-switch flows and Google OAuth were
not exercised in this acceptance. This was not a complete accessibility or UX
audit. A known display defect remains: a newly created course can show the
student's overall completed-lesson count. A follow-up code fix is present but
its browser verification is blocked as described above. The test stack was left running;
there was no deployment, cleanup, or second seed.

GitHub Actions runs backend tests plus frontend type checks and production builds for pull requests and updates to `main`.

## Deployment

| Target | Guide |
| --- | --- |
| Short IP-based presentation | [DEMO_DEPLOYMENT.md](DEMO_DEPLOYMENT.md) |
| HTTPS deployment on a real domain | [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md) |

The production stack includes Caddy-managed HTTPS, non-public database networking, health checks and scheduled PostgreSQL dumps. Before accepting real learners, configure off-site encrypted backups and an external monitor for `/api/v1/health`.

## Security notes

- Do not commit `.env`, `.env.demo` or `.env.production` files.
- Rotate any credential that was ever pasted into a chat, terminal transcript or screenshot.
- Keep PostgreSQL, backend and frontend ports private; expose only 80/443 through the reverse proxy.
- Uploads are limited to 10 MiB and a safe allowlist of document and image formats.
- Use unique production passwords and a long random JWT secret.

## License

This project is currently private and unlicensed. Add a licence before distributing or accepting external contributions.
