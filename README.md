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
cd backend && go test ./...

# Frontend type check and optimised build
cd frontend && npm run type-check && npm run build

# E2E deployment smoke tests
cd frontend && npm run test:e2e
```

The default Playwright suite checks API health, the sign-in form and student/teacher workspaces using local NU demo data. To run against another environment, configure `E2E_BASE_URL`, `E2E_API_URL`, `TEST_ORG` and test-account variables.

Detailed course UI tests are opt-in because they depend on a selected interface language:

```bash
E2E_RUN_EXTENDED=true npm run test:e2e
```

Any E2E test that modifies course data additionally requires `E2E_ALLOW_WRITES=true`.

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
