#!/bin/sh
# Safe regression tests for the E2E harness. Docker is fully stubbed: no
# container, volume, bind mount, seed, browser, or destructive Compose command
# is run against the host.
set -eu

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
workspace=$(mktemp -d "${TMPDIR:-/tmp}/mentra-e2e-harness.XXXXXX")
cleanup() { find "$workspace" -depth -delete; }
trap cleanup EXIT HUP INT TERM
mkdir "$workspace/bin"
log_file="$workspace/docker.log"

# The stack must mount only its project-scoped named volumes. Build contexts are
# allowed; host paths in a `- source:target` volume entry are not.
if grep -E '^[[:space:]]*-[[:space:]]*(\./|/)' "$root_dir/docker-compose.e2e.yml" >/dev/null; then
  echo 'E2E Compose file contains a host bind mount' >&2
  exit 1
fi
for volume in e2e_pgdata e2e_uploads e2e_ollama_models; do
  grep -F -- "$volume:" "$root_dir/docker-compose.e2e.yml" >/dev/null
done

cat > "$workspace/bin/docker" <<'EOF'
#!/bin/sh
set -eu
printf '%s\n' "$*" >> "$E2E_STUB_DOCKER_LOG"
if [ "$1" = context ] && [ "$2" = show ]; then
  printf '%s\n' "${E2E_STUB_CONTEXT:-default}"
  exit 0
fi
if [ "$1" = context ] && [ "$2" = inspect ]; then
  printf '%s\n' "${E2E_STUB_ENDPOINT:-unix:///var/run/docker.sock}"
  exit 0
fi
if [ "$1" = ps ]; then
  case "${E2E_STUB_INVENTORY_ERROR:-}" in
    container-label) case " $* " in *' --filter '*) exit 17 ;; esac ;;
    container-name) case " $* " in *' --filter '*) ;; *) exit 17 ;; esac ;;
  esac
  case "${E2E_STUB_EXISTING_CONTAINER:-}" in
    label) case " $* " in *' --filter '*) printf '%s\n' e2e-container-id ;; esac ;;
    name) case " $* " in *' --filter '*) ;; *) printf '%s\n' mentra-e2e-app-1 ;; esac ;;
  esac
  exit 0
fi
if [ "$1" = volume ] && [ "$2" = ls ]; then
  case "${E2E_STUB_INVENTORY_ERROR:-}" in
    volume-label) case " $* " in *' --filter '*) exit 17 ;; esac ;;
    volume-name) case " $* " in *' --filter '*) ;; *) exit 17 ;; esac ;;
  esac
  case "${E2E_STUB_EXISTING_VOLUME:-}" in
    label) case " $* " in *' --filter '*) printf '%s\n' mentra-e2e_e2e_pgdata ;; esac ;;
    name) case " $* " in *' --filter '*) ;; *) printf '%s\n' mentra-e2e_e2e_pgdata ;; esac ;;
  esac
  exit 0
fi
exit 0
EOF
chmod +x "$workspace/bin/docker"

run_clean_shell() {
  env -i PATH="$workspace/bin:/usr/bin:/bin" TMPDIR="$workspace" E2E_STUB_DOCKER_LOG="$log_file" \
    sh "$root_dir/scripts/e2e/cleanup-isolated.sh"
}

# A new shell has no E2E variables. The generated test-only env file must let
# Compose evaluate the required interpolation before the isolated teardown.
run_clean_shell
grep -F -- '--project-name mentra-e2e --env-file ' "$log_file" >/dev/null
grep -F -- '-f docker-compose.e2e.yml down --volumes --remove-orphans' "$log_file" >/dev/null

: > "$log_file"
if env -i PATH="$workspace/bin:/usr/bin:/bin" TMPDIR="$workspace" E2E_STUB_DOCKER_LOG="$log_file" \
  DOCKER_HOST='ssh://remote.example.test' sh "$root_dir/scripts/e2e/cleanup-isolated.sh"; then
  echo 'cleanup unexpectedly accepted a remote DOCKER_HOST' >&2
  exit 1
fi
if grep -E '^compose ' "$log_file" >/dev/null; then
  echo 'cleanup attempted Docker Compose after rejecting a remote host' >&2
  exit 1
fi

: > "$log_file"
if env -i PATH="$workspace/bin:/usr/bin:/bin" TMPDIR="$workspace" E2E_STUB_DOCKER_LOG="$log_file" \
  DOCKER_CONTEXT=remote E2E_STUB_ENDPOINT='ssh://remote.example.test' sh "$root_dir/scripts/e2e/prepare-isolated.sh"; then
  echo 'prepare unexpectedly accepted an overridden remote Docker context' >&2
  exit 1
fi
if grep -E '^compose ' "$log_file" >/dev/null; then
  echo 'prepare attempted Docker Compose after rejecting an overridden context' >&2
  exit 1
fi

: > "$log_file"
if env -i PATH="$workspace/bin:/usr/bin:/bin" TMPDIR="$workspace" E2E_STUB_DOCKER_LOG="$log_file" \
  E2E_STUB_ENDPOINT='ssh://remote.example.test' sh "$root_dir/scripts/e2e/prepare-isolated.sh"; then
  echo 'prepare unexpectedly accepted a remote Docker context endpoint' >&2
  exit 1
fi
if grep -E '^compose ' "$log_file" >/dev/null; then
  echo 'prepare attempted Docker Compose after rejecting a remote context' >&2
  exit 1
fi

: > "$log_file"
if env -i PATH="$workspace/bin:/usr/bin:/bin" TMPDIR="$workspace" E2E_STUB_DOCKER_LOG="$log_file" \
  E2E_PROJECT_NAME=existing-production sh "$root_dir/scripts/e2e/cleanup-isolated.sh"; then
  echo 'cleanup unexpectedly accepted a non-isolated project name' >&2
  exit 1
fi
if [ -s "$log_file" ]; then
  echo 'cleanup invoked Docker for a rejected project name' >&2
  exit 1
fi

# `prepare` must never reseed an existing project, even if Compose has no
# labels on an old container or volume. Every refusal must precede `up`.
for existing in container-label container-name volume-label volume-name; do
  : > "$log_file"
  container_case=
  volume_case=
  case "$existing" in
    container-label) container_case=label ;;
    container-name) container_case=name ;;
    volume-label) volume_case=label ;;
    volume-name) volume_case=name ;;
  esac
  if env -i PATH="$workspace/bin:/usr/bin:/bin" TMPDIR="$workspace" E2E_STUB_DOCKER_LOG="$log_file" \
    E2E_STUB_EXISTING_CONTAINER="$container_case" E2E_STUB_EXISTING_VOLUME="$volume_case" \
    sh "$root_dir/scripts/e2e/prepare-isolated.sh"; then
    echo "prepare unexpectedly accepted an existing $existing" >&2
    exit 1
  fi
  if grep -E '^compose ' "$log_file" >/dev/null; then
    echo "prepare called Docker Compose despite existing $existing" >&2
    exit 1
  fi
done

for failing_inventory in container-label container-name volume-label volume-name; do
  : > "$log_file"
  if env -i PATH="$workspace/bin:/usr/bin:/bin" TMPDIR="$workspace" E2E_STUB_DOCKER_LOG="$log_file" \
    E2E_STUB_INVENTORY_ERROR="$failing_inventory" sh "$root_dir/scripts/e2e/prepare-isolated.sh"; then
    echo "prepare unexpectedly accepted a failed $failing_inventory inventory" >&2
    exit 1
  fi
  if grep -E '^compose ' "$log_file" >/dev/null; then
    echo "prepare called Docker Compose despite failed $failing_inventory inventory" >&2
    exit 1
  fi
done

echo 'e2e harness tests passed'
