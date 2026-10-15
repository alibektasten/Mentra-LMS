#!/usr/bin/env node
// Regression tests for verify-write-target.mjs. The fake Docker binary only
// records read-only context/config/ps/inspect calls; no host Docker command is
// allowed to run.
import { execFileSync } from 'node:child_process'
import { chmodSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'

const rootDir = path.resolve(path.dirname(new URL(import.meta.url).pathname), '../..')
const tempDir = mkdtempSync(path.join(tmpdir(), 'mentra-write-guard-'))
const binDir = path.join(tempDir, 'bin')
const logFile = path.join(tempDir, 'docker.log')
const fakeDocker = path.join(binDir, 'docker')

try {
  writeFileSync(path.join(tempDir, 'create-bin.sh'), `mkdir -p '${binDir}'\n`)
  execFileSync('sh', [path.join(tempDir, 'create-bin.sh')])
  writeFileSync(fakeDocker, `#!/bin/sh
set -eu
printf '%s\\n' "$*" >> "$E2E_GUARD_LOG"
case "$1:$2" in
  context:show) printf '%s\\n' default ;;
  context:inspect) printf '%s\\n' "\${E2E_GUARD_ENDPOINT:-unix:///var/run/docker.sock}" ;;
  compose:--project-name)
    case " $* " in
      *" config "*) printf '%s\\n' '{"services":{"frontend":{"build":{"args":{"NEXT_PUBLIC_API_URL":"http://127.0.0.1:18080/api/v1"}}}}}' ;;
      *" ps "*)
        if [ "\${E2E_GUARD_CASE:-valid}" = missing-proof ]; then printf '%s\\n' '{"ID":"app-id","Project":"mentra-e2e","Service":"app","State":"running","Publishers":[{"TargetPort":8080,"PublishedPort":18080}]}'
        else printf '%s\\n' '[{"ID":"app-id","Project":"mentra-e2e","Service":"app","State":"running","Publishers":[{"TargetPort":8080,"PublishedPort":18080}]},{"ID":"frontend-id","Project":"mentra-e2e","Service":"frontend","State":"running","Publishers":[{"TargetPort":3000,"PublishedPort":13000}]}]' ; fi ;;
    esac ;;
  inspect:--format)
    case "$*" in *'{{.Image}}'*) printf '%s\\n' frontend-image-id ;;
      *) if [ "\${E2E_GUARD_CASE:-valid}" = wrong-label ]; then printf '%s\\n' other:app
         else case "$*" in *frontend-id*) printf '%s\\n' mentra-e2e:frontend ;; *) printf '%s\\n' mentra-e2e:app ;; esac; fi ;;
    esac ;;
  image:inspect)
    case "\${E2E_GUARD_CASE:-valid}" in
      stale-image-api) printf '%s\\n' http://127.0.0.1:19999/api/v1 ;;
      missing-image-label) printf '%s\\n' '' ;;
      *) printf '%s\\n' http://127.0.0.1:18080/api/v1 ;;
    esac ;;
esac
`)
  chmodSync(fakeDocker, 0o755)

  const base = {
    PATH: `${binDir}:${process.env.PATH}`,
    E2E_GUARD_LOG: logFile,
    E2E_ISOLATED: 'true',
    E2E_BASE_URL: 'http://127.0.0.1:13000',
    E2E_API_URL: 'http://127.0.0.1:18080/api/v1',
  }
  const verify = path.join(rootDir, 'scripts/e2e/verify-write-target.mjs')
  const run = (extra = {}) => execFileSync(process.execPath, [verify], { env: { ...base, ...extra }, stdio: 'pipe' })
  const mustReject = (name, extra) => {
    try { run(extra) } catch { return }
    throw new Error(`${name} unexpectedly passed`)
  }

  run()
  mustReject('remote API', { E2E_API_URL: 'https://staging.example.test/api/v1' })
  mustReject('localhost frontend with another local API', { E2E_API_URL: 'http://127.0.0.1:19999/api/v1' })
  mustReject('other local frontend stack', { E2E_BASE_URL: 'http://127.0.0.1:13001' })
  mustReject('missing frontend runtime proof', { E2E_GUARD_CASE: 'missing-proof' })
  mustReject('wrong Compose labels', { E2E_GUARD_CASE: 'wrong-label' })
  mustReject('old frontend image with another embedded API', { E2E_GUARD_CASE: 'stale-image-api' })
  mustReject('frontend image without build metadata', { E2E_GUARD_CASE: 'missing-image-label' })

  const calls = readFileSync(logFile, 'utf8')
  if (/\b(up|down|exec|run|rm)\b/.test(calls)) throw new Error(`unexpected Docker write call: ${calls}`)
  console.log('write target guard tests passed')
} finally {
  rmSync(tempDir, { recursive: true, force: true })
}
