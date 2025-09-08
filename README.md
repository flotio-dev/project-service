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

Podman detected on this machine: `podman --version` should return your installed version.
