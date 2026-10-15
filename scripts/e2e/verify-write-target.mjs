#!/usr/bin/env node
/**
 * Read-only proof that a future write-E2E run targets the disposable
 * `mentra-e2e` stack. It must finish before Playwright is started.
 */
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..')
const composeFile = path.join(rootDir, 'docker-compose.e2e.yml')
const project = 'mentra-e2e'
const localHosts = new Set(['127.0.0.1', 'localhost', '::1'])

function fail(message) {
  throw new Error(`Refusing write E2E: ${message}`)
}

function required(name) {
  const value = process.env[name]?.trim()
  if (!value) fail(`${name} is required`)
  return value
}

function localURL(name) {
  let value
  try {
    value = new URL(required(name))
  } catch {
    fail(`${name} must be an absolute URL`)
  }
  if (!['http:', 'https:'].includes(value.protocol) || !localHosts.has(value.hostname)) {
    fail(`${name} must use a local http(s) host`)
  }
  return value
}

function docker(args, extraEnv = {}) {
  try {
    return execFileSync('docker', args, {
      cwd: rootDir,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
      env: { ...process.env, ...extraEnv },
    }).trim()
  } catch (error) {
    const stderr = error?.stderr?.toString().trim()
    fail(`cannot inspect the isolated Docker stack${stderr ? `: ${stderr}` : ''}`)
  }
}

function parseJSONOutput(output, label) {
  try {
    return JSON.parse(output)
  } catch {
    try {
      return output.split('\n').filter(Boolean).map((line) => JSON.parse(line))
    } catch {
      fail(`${label} returned invalid JSON`)
    }
  }
}

function portFor(url) {
  return Number(url.port || (url.protocol === 'https:' ? 443 : 80))
}

function composeEnv(baseURL, apiURL) {
  return {
    E2E_POSTGRES_PASSWORD: 'e2e-postgres-local-only',
    E2E_JWT_SECRET: 'e2e-jwt-local-only-not-for-production',
    E2E_WEB_PORT: String(portFor(baseURL)),
    E2E_API_PORT: String(portFor(apiURL)),
    E2E_SUPER_ADMIN_EMAIL: 'e2e-super-admin@mentra.test',
    E2E_SUPER_ADMIN_PASSWORD: 'e2e-super-admin-password',
  }
}

function composeArgs(command) {
  return ['compose', '--project-name', project, '--project-directory', rootDir, '-f', composeFile, ...command]
}

function normalizeServices(value) {
  return Array.isArray(value) ? value : [value]
}

function publishedPort(service, targetPort) {
  const publisher = (service.Publishers ?? []).find((entry) => Number(entry.TargetPort) === targetPort)
  return publisher ? Number(publisher.PublishedPort) : 0
}

function assertContainerIdentity(service, expectedService) {
  if (!service?.ID || service.Service !== expectedService || service.Project !== project || service.State !== 'running') {
    fail(`expected a running ${project}/${expectedService} container`)
  }
  const label = docker([
    'inspect', '--format',
    '{{ index .Config.Labels "com.docker.compose.project" }}:{{ index .Config.Labels "com.docker.compose.service" }}',
    service.ID,
  ])
  if (label !== `${project}:${expectedService}`) {
    fail(`container ${service.ID} does not have the expected Compose identity`)
  }
}

function assertFrontendImageAPI(frontend, expectedAPI) {
  const imageID = docker(['inspect', '--format', '{{.Image}}', frontend.ID])
  if (!imageID) fail('frontend container does not expose an image identity')
  const builtAPI = docker([
    'image', 'inspect', '--format',
    '{{ index .Config.Labels "org.mentra.frontend.api-url" }}',
    imageID,
  ])
  if (!builtAPI) fail('frontend image has no org.mentra.frontend.api-url build metadata')
  if (builtAPI !== expectedAPI) fail('running frontend image was built with a different API URL')
}

function main() {
  if (process.env.E2E_ISOLATED !== 'true') fail('E2E_ISOLATED=true is required')
  if (process.env.E2E_PROJECT_NAME && process.env.E2E_PROJECT_NAME !== project) {
    fail('E2E_PROJECT_NAME must be mentra-e2e')
  }

  const baseURL = localURL('E2E_BASE_URL')
  const apiURL = localURL('E2E_API_URL')
  if (apiURL.hostname !== '127.0.0.1' || apiURL.pathname.replace(/\/$/, '') !== '/api/v1') {
    fail('E2E_API_URL must exactly target the isolated app mapping at http://127.0.0.1:<port>/api/v1')
  }

  if (process.env.DOCKER_HOST && !/^(unix|npipe):\/\//.test(process.env.DOCKER_HOST)) {
    fail('DOCKER_HOST must use a local Unix/npipe socket')
  }
  const context = process.env.DOCKER_CONTEXT || docker(['context', 'show'])
  const endpoint = docker(['context', 'inspect', context, '--format', '{{ .Endpoints.docker.Host }}'])
  if (!/^(unix|npipe):\/\//.test(endpoint)) fail(`Docker context ${context} is not local`)

  const env = composeEnv(baseURL, apiURL)
  const config = parseJSONOutput(docker(composeArgs(['config', '--format', 'json']), env), 'docker compose config')
  const expectedAPI = `http://127.0.0.1:${portFor(apiURL)}/api/v1`
  const configuredAPI = config?.services?.frontend?.build?.args?.NEXT_PUBLIC_API_URL
  if (configuredAPI !== expectedAPI) {
    fail('frontend build API URL does not match E2E_API_URL')
  }

  const services = normalizeServices(parseJSONOutput(docker(composeArgs(['ps', '--format', 'json']), env), 'docker compose ps'))
  const app = services.find((service) => service.Service === 'app')
  const frontend = services.find((service) => service.Service === 'frontend')
  assertContainerIdentity(app, 'app')
  assertContainerIdentity(frontend, 'frontend')
  if (publishedPort(app, 8080) !== portFor(apiURL)) fail('E2E_API_URL is not the app container mapping')
  if (publishedPort(frontend, 3000) !== portFor(baseURL)) fail('E2E_BASE_URL is not the frontend container mapping')
  assertFrontendImageAPI(frontend, expectedAPI)

  process.stdout.write('isolated write target verified\n')
}

main()
