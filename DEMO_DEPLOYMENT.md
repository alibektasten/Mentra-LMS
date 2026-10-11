# Tomorrow's VPS demo

This deployment is designed for a short presentation. It runs the app, PostgreSQL, and Nginx on one VPS, uses the OpenAI API for AI features, and **does not run Ollama**. You can open it via the server's public IPv4 address — no domain or HTTPS setup is needed for the demo.

## In the Hetzner creation form

Choose these options:

| Field | Choose / enter |
| --- | --- |
| Type | **CPX32** (4 vCPU, 8 GB RAM) |
| Location | **HEL1** (or NBG1 if you prefer) |
| Image | **Ubuntu 24.04** |
| Public networking | **IPv4 enabled** |
| SSH key | Add your Mac's public SSH key before creating the server |
| Firewall | Allow TCP **22** from your own IP and TCP **80** from anywhere. Do **not** open 5432, 8080, or 3000. |
| Backups / volumes | Not needed for a one-day demo |

After creation, copy the public IPv4 address from the Hetzner console. In this guide it is called `SERVER_IP`.

## Fill one file

On the VPS, after cloning the repository, create the real environment file:

```bash
cp .env.demo.example .env.demo
nano .env.demo
```

Fill only these entries:

| Variable | Exactly what to enter |
| --- | --- |
| `PUBLIC_ORIGIN` | `http://SERVER_IP` — replace `SERVER_IP` with the actual IPv4, no trailing `/` |
| `POSTGRES_PASSWORD` | Output of `openssl rand -hex 32` |
| `JWT_SECRET` | Output of a **second** `openssl rand -hex 32` command |
| `OPENAI_API_KEY` | Your secret API key from the OpenAI platform |
| `OPENAI_MODEL` | Keep `gpt-5.4-mini` unless you deliberately want another supported model |
| `SUPER_ADMIN_EMAIL` | Your email for the system-level admin account |
| `SUPER_ADMIN_PASSWORD` | A new strong password |

Do not commit, send, or screenshot `.env.demo`. The `.gitignore` file prevents it from being added accidentally.

## First deployment commands

First make sure your current local changes are committed and pushed to GitHub. The VPS can only clone what is in the remote repository.

SSH into the new Ubuntu VPS:

```bash
ssh root@SERVER_IP
```

Install Docker Engine and the Compose plugin from Docker's official Ubuntu repository, then clone the project:

```bash
sudo apt update
sudo apt install -y ca-certificates curl git
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
sudo tee /etc/apt/sources.list.d/docker.sources >/dev/null <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

git clone https://github.com/sailybaev/Mentra-LMS.git
cd Mentra-LMS
```

Create and fill `.env.demo` as described above, then start the demo:

```bash
docker compose --env-file .env.demo -f docker-compose.demo.yml up -d --build
```

Watch until the API has started successfully:

```bash
docker compose --env-file .env.demo -f docker-compose.demo.yml logs -f app
```

Press `Ctrl+C` after you see that startup/migrations have completed; the containers keep running. Then seed the ready-made presentation data:

```bash
docker compose --env-file .env.demo -f docker-compose.demo.yml exec app ./seed
docker compose --env-file .env.demo -f docker-compose.demo.yml restart app
```

`./seed` deliberately deletes the application's current data before creating sample organizations, courses, groups, and accounts. Use it only for this fresh presentation server. The restart recreates the configured super-admin account after the seed reset.

Open these in the browser:

| What to demonstrate | URL | Sign in as |
| --- | --- | --- |
| Student role | `http://SERVER_IP/nu/login` | `sanzhar@nu.edu.kz` / `password123` |
| Teacher role | `http://SERVER_IP/nu/login` | `baurzhan@nu.edu.kz` / `password123` |
| Organization admin | `http://SERVER_IP/nu/login` | `aiguerim@nu.edu.kz` / `password123` |
| System admin | `http://SERVER_IP/login` | The `SUPER_ADMIN_EMAIL` and `SUPER_ADMIN_PASSWORD` you set |

## Quick checks before presenting

```bash
docker compose --env-file .env.demo -f docker-compose.demo.yml ps
curl -I http://SERVER_IP
```

All four services should be `running` and the second command should return an HTTP response. In the browser, log in as the student, open a published course, then use one AI study feature. That confirms the OpenAI key and the student permissions flow.

## When the presentation ends

In the Hetzner console, **delete the server** (not merely power it off) so hourly billing stops. Copy anything you need first; deletion removes the VPS disk and its data.
