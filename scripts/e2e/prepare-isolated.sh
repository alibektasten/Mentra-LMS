#!/bin/sh
# Create and seed a disposable local E2E stack. This script never invokes
# Playwright and refuses any non-local Docker context or project name other
# than mentra-e2e.
set -eu

. "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/common.sh"

project_name=$(e2e_project_name)
assert_local_docker
assert_fresh_e2e_project
create_e2e_env_file
ports=$(e2e_numeric_ports)
web_port=${ports%%:*}
api_port=${ports##*:}

root_dir=$(e2e_root_dir)
env_file=$e2e_env_file
cleanup() { remove_e2e_env_file; }
trap cleanup EXIT HUP INT TERM

cd "$root_dir"
docker compose --project-name "$project_name" --env-file "$env_file" -f docker-compose.e2e.yml up -d --build --wait
docker compose --project-name "$project_name" --env-file "$env_file" -f docker-compose.e2e.yml exec -T app ./seed
# Seed clears users by design, including the bootstrapped super-admin. Restart
# only this isolated app so its normal startup provisioning creates it again.
docker compose --project-name "$project_name" --env-file "$env_file" -f docker-compose.e2e.yml restart app
docker compose --project-name "$project_name" --env-file "$env_file" -f docker-compose.e2e.yml up -d --wait app

cat <<EOF

Isolated stack is ready. It was seeded only inside Docker project $project_name.
Run the following in this shell before a future browser test:

export E2E_ISOLATED=true
export E2E_BASE_URL=http://127.0.0.1:$web_port
export E2E_API_URL=http://127.0.0.1:$api_port/api/v1
export E2E_TEST_ORG=nu
export E2E_LOCALE=en-US
export E2E_TIMEZONE=Asia/Almaty
export E2E_STUDENT_EMAIL=sanzhar@nu.edu.kz
export E2E_STUDENT_PASSWORD=password123
export E2E_OTHER_STUDENT_EMAIL=zhanar@nu.edu.kz
export E2E_OTHER_STUDENT_PASSWORD=password123
export E2E_TEACHER_EMAIL=baurzhan@nu.edu.kz
export E2E_TEACHER_PASSWORD=password123
export E2E_OTHER_TEACHER_EMAIL=damir@nu.edu.kz
export E2E_OTHER_TEACHER_PASSWORD=password123
export E2E_ADMIN_EMAIL=aiguerim@nu.edu.kz
export E2E_ADMIN_PASSWORD=password123
export E2E_OTHER_ORG=kbtu
export E2E_OTHER_ORG_STUDENT_EMAIL=nurdaulet@kbtu.edu.kz
export E2E_OTHER_ORG_STUDENT_PASSWORD=password123
export E2E_SUPER_ADMIN_EMAIL=e2e-super-admin@mentra.test
export E2E_SUPER_ADMIN_PASSWORD=e2e-super-admin-password
EOF
