#!/bin/sh
# Delete only the named disposable E2E Compose project and its named volumes.
# It deliberately cannot be pointed at a development, demo, or production stack.
set -eu

. "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/common.sh"
project_name=$(e2e_project_name)
assert_local_docker
create_e2e_env_file
cleanup() { remove_e2e_env_file; }
trap cleanup EXIT HUP INT TERM

cd "$(e2e_root_dir)"
docker compose --project-name "$project_name" --env-file "$e2e_env_file" -f docker-compose.e2e.yml down --volumes --remove-orphans
