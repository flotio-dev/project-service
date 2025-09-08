# project-service

Project scaffold.

## Podman Compose & PostgreSQL

This repository includes a Compose setup to run a PostgreSQL database for local development using Podman.

Files added:

- `docker-compose.yml` - Compose stack with `db` (Postgres) and `app` (built from `Dockerfile`).
- `docker/init.sql` - Database initialization SQL (creates `projects` table).
- `.env.example` - Example environment variables for the database.

Quick start (Podman):

1. Copy the example env and edit if needed:

```bash
cp .env.example .env
# edit .env to your needs
```

2. Start the stack with Podman Compose:

```bash
podman compose up -d
```

3. Check the database health / logs:

```bash
podman compose ps
podman compose logs -f db
```

4. To stop and remove volumes (data will be lost):

```bash
podman compose down -v
```

Notes:

- Podman provides a Compose-compatible plugin; commands mirror Docker Compose but run rootless by default. If you prefer the Docker-compatible CLI, use `podman-docker` / `docker` shim if installed on your system.
- The `app` service in `docker-compose.yml` uses the repository `Dockerfile` to build the application image and expects the service to listen on port `8080`. Update if your app uses a different port.
- The DB initialization script is mounted into `/docker-entrypoint-initdb.d/` so it runs only on first container start when the volume is empty.

Environment variables:

- `PORT`: HTTP port for the service (default 8080).
- `DATABASE_URL`: Postgres connection string for the app.
- `KEYCLOAK_BASE_URL` and `KEYCLOAK_REALM`: used to fetch JWKS and validate JWTs.

Podman detected on this machine: `podman --version` should return your installed version.

## Devenv (Nix) 🍀

This repo ships a `devenv` development environment for a local Postgres and app env vars without Docker.

What you get:
- A local Postgres 15 service on 127.0.0.1:5433
- Env vars automatically exported when you `direnv allow`:
  - `POSTGRES_USER=postgres`, `POSTGRES_PASSWORD=postgres`, `POSTGRES_DB=projectdb`, `DB_PORT=5433`
  - `DATABASE_URL=postgres://postgres:postgres@127.0.0.1:5433/projectdb?sslmode=disable`

How to use:
1. Install direnv and devenv (see https://devenv.sh/)
2. In the repo:
	- `direnv allow`
	- `devenv up` (optional; services auto-start on first use)
3. Verify:
	- `psql "$DATABASE_URL" -c 'select 1'`

Notes:
- Devenv Postgres listens on 5433 to avoid clashing with Docker/Podman 5432 mapping.
- `.env` remains used by docker-compose; devenv variables are separate and injected by direnv.
- You can override values by editing `devenv.nix` or exporting before `direnv allow`.
