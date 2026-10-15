#!/bin/sh
# Shared guards for the disposable local E2E Compose project.

e2e_project_name() {
  name=${E2E_PROJECT_NAME:-mentra-e2e}
  if [ "$name" != "mentra-e2e" ]; then
    echo "E2E_PROJECT_NAME must be mentra-e2e; refusing a non-isolated project." >&2
    return 1
  fi
  printf '%s\n' "$name"
}

e2e_numeric_ports() {
  web_port=${E2E_WEB_PORT:-13000}
  api_port=${E2E_API_PORT:-18080}
  case "$web_port:$api_port" in
    *[!0-9:]* | :* | *:)
      echo "E2E_WEB_PORT and E2E_API_PORT must be numeric." >&2
      return 1
      ;;
  esac
  printf '%s:%s\n' "$web_port" "$api_port"
}

assert_local_docker() {
  if [ -n "${DOCKER_HOST:-}" ]; then
    case "$DOCKER_HOST" in
      unix://* | npipe://*) ;;
      *)
        echo "Refusing Docker host $DOCKER_HOST; E2E data is allowed only on a local Unix/npipe socket." >&2
        return 1
        ;;
    esac
  fi

  # DOCKER_CONTEXT overrides the selected context for Docker Compose. Inspect
  # that exact context instead of trusting `docker context show` alone.
  context=${DOCKER_CONTEXT:-$(docker context show)}
  endpoint=$(docker context inspect "$context" --format '{{ .Endpoints.docker.Host }}')
  case "$endpoint" in
    unix://* | npipe://*) return 0 ;;
    *)
      echo "Refusing Docker context $context ($endpoint); E2E data is allowed only on a local Unix/npipe socket." >&2
      return 1
      ;;
  esac
}

# `prepare` runs a destructive seed, so it may only ever create a brand-new
# test project. Compose labels and names are both checked: old volumes may
# predate Compose's project labels, and stopped containers still count.
assert_fresh_e2e_project() {
  project_name=$(e2e_project_name) || return 1
  labelled_containers=$(docker ps -a --filter "label=com.docker.compose.project=$project_name" --format '{{.ID}}') || return 1
  all_container_names=$(docker ps -a --format '{{.Names}}') || return 1
  named_containers=$(printf '%s\n' "$all_container_names" | grep "^${project_name}[-_]" || true)
  labelled_volumes=$(docker volume ls --filter "label=com.docker.compose.project=$project_name" --format '{{.Name}}') || return 1
  all_volume_names=$(docker volume ls --format '{{.Name}}') || return 1
  named_volumes=$(printf '%s\n' "$all_volume_names" | grep "^${project_name}_" || true)
  if [ -n "$labelled_containers" ] || [ -n "$named_containers" ] || [ -n "$labelled_volumes" ] || [ -n "$named_volumes" ]; then
    echo "Refusing prepare: existing $project_name containers or volumes detected; seed would reset their data." >&2
    return 1
  fi
}

create_e2e_env_file() {
  ports=$(e2e_numeric_ports) || return 1
  web_port=${ports%%:*}
  api_port=${ports##*:}
  e2e_env_file=$(mktemp "${TMPDIR:-/tmp}/mentra-e2e.XXXXXX") || return 1

  umask 077
  {
    printf 'E2E_POSTGRES_PASSWORD=%s\n' 'e2e-postgres-local-only'
    printf 'E2E_JWT_SECRET=%s\n' 'e2e-jwt-local-only-not-for-production'
    printf 'E2E_WEB_PORT=%s\n' "$web_port"
    printf 'E2E_API_PORT=%s\n' "$api_port"
    printf 'E2E_SUPER_ADMIN_EMAIL=%s\n' 'e2e-super-admin@mentra.test'
    printf 'E2E_SUPER_ADMIN_PASSWORD=%s\n' 'e2e-super-admin-password'
  } > "$e2e_env_file"
}

remove_e2e_env_file() {
  if [ -n "${e2e_env_file:-}" ]; then
    rm -f "$e2e_env_file"
  fi
}

e2e_root_dir() {
  CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd
}
