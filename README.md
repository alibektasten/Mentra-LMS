# Mentra LMS

## Проверка перед развёртыванием

После запуска frontend и backend выполните smoke-проверку основных пользовательских сценариев:

```bash
cd frontend
npm run test:e2e
```

По умолчанию она проверяет `GET /health`, форму входа и вход в кабинеты студента и преподавателя на демо-данных NU. Для staging/production адреса и тестовые аккаунты передаются через переменные окружения: `E2E_BASE_URL`, `E2E_API_URL`, `TEST_ORG`, `TEST_STUDENT_EMAIL`, `TEST_STUDENT_PASSWORD`, `TEST_TEACHER_EMAIL`, `TEST_TEACHER_PASSWORD`.

Расширенные проверки страниц курсов запускаются отдельно: `E2E_RUN_EXTENDED=true npm run test:e2e`. Тесты, создающие или редактирующие учебный контент, требуют дополнительного явного флага `E2E_ALLOW_WRITES=true`.

A multi-tenant Learning Management System with AI-powered features.

**Backend:** Go 1.23 · Gin · GORM · PostgreSQL + pgvector
**Frontend:** Next.js 15 · React 19 · TypeScript · Tailwind CSS
**AI:** OpenAI Responses API or Ollama (local fallback)

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

---

## Quick Start

1. **Clone the repo**

   ```bash
   git clone https://github.com/your-username/mentra-lms.git
   cd mentra-lms
   ```

2. **Configure environment**

   ```bash
   cp backend/.env.example backend/.env
   ```

   Edit `backend/.env` and fill in your values (DB credentials, JWT secret, etc.).

3. **Start all services**

   ```bash
   docker-compose up --build
   ```

   | Service  | URL                      |
   |----------|--------------------------|
   | Frontend | http://localhost         |
   | API      | http://localhost/api/v1  |
   | Ollama   | http://localhost:11434   |

4. **Seed the database** (first run)

   ```bash
   docker-compose exec app ./seed
   ```

---

## Project Structure

```
backend/    Go REST API (Clean Architecture)
frontend/   Next.js app (App Router)
```

---

## Multi-Tenancy

Each organization gets its own subdomain/slug. All API requests require an `X-Org-Slug` header. JWT tokens embed `org_id` and `role` — all data is scoped per tenant.

---

## AI provider

The backend selects the provider, so the OpenAI key is never exposed to the browser. For OpenAI, add these values to `backend/.env` and restart the API:

```bash
AI_PROVIDER=openai
OPENAI_API_KEY=your_key_here
OPENAI_MODEL=gpt-5.4-mini
```

Use `AI_PROVIDER=ollama` to keep the local model. Quiz generation is restricted to teachers and organization admins; students retain summaries, flashcards, progress insights, and feedback for their own text submissions.

---

## VPS demo with OpenAI

For an IP-address-based demonstration VPS that uses OpenAI and does not run
Ollama, follow [DEMO_DEPLOYMENT.md](DEMO_DEPLOYMENT.md). It includes the single
environment file to fill in, Hetzner settings, deployment commands, and sample
login accounts.
