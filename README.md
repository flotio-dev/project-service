# project-service

Project scaffold.

## Docker Compose & PostgreSQL

This repository includes a Docker Compose setup to run a PostgreSQL database for local development.

Files added:

- `docker-compose.yml` - Compose stack with `db` (Postgres) and `app` (built from `Dockerfile`).
- `docker/init.sql` - Database initialization SQL (creates `projects` table).
- `.env.example` - Example environment variables for the database.

Quick start:

1. Copy the example env and edit if needed:

```bash
cp .env.example .env
# edit .env to your needs
```

2. Start the stack:

```bash
docker compose up -d
```

3. Check the database health:

```bash
docker compose ps
docker compose logs db --follow
```

4. To stop and remove volumes (data will be lost):

```bash
docker compose down -v
```

Notes:

- The `app` service in `docker-compose.yml` uses the repository `Dockerfile` to build the application image and expects the service to listen on port `8080`. Update if your app uses a different port.
- The DB initialization script is mounted into `/docker-entrypoint-initdb.d/` so it runs only on first container start when the volume is empty.