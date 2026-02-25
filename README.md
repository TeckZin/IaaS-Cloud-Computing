# Server

Go HTTP API server with PostgreSQL persistence.

## Project Structure

```
server/
├─ api/                    # (optional) API-related docs or contracts
├─ go.mod / go.sum         # Go module definition + dependencies
├─ main/
│  ├─ main.go              # entrypoint (boot server, connect DB)
│  └─ routes.go            # route registration + HTTP handlers
├─ models/
│  └─ userModel.go         # request/response + domain models
├─ sql/
│  └─ V001_create_users.sql# DB migration (create users table)
└─ storage/
   └─ UserStorage.go       # DB access layer (CRUD functions)
```

## Requirements

- Go (1.20+ recommended)
- Docker (for PostgreSQL)
- (Optional) `psql` CLI for DB inspection

## Environment Variables

Create a `.env` file in `server/` (same folder as `go.mod`):

locally
```env
PORT=8080
POSTGRES_USER=...
POSTGRES_PASSWORD=...
POSTGRES_DB=...
POSTGRES_PORT=5432
```

## Run PostgreSQL in Docker (local)

From the `server/` folder (or anywhere if your `.env` path is correct):

```bash
docker volume create pgdata

docker run -d \
  --name IaaSDemoPostgres \
  --env-file .env \
  -p 5432:5432 \
  -v pgdata:/var/lib/postgresql/data \
  postgres:16
```

Check logs:

```bash
docker logs -f IaaSDemoPostgres
```

## Run the API Server

Load the `.env` into your current shell (Go does not auto-load `.env`):

```bash
cd server
set -a; source .env; set +a
go run ./main
```

You should see the server listening on `http://localhost:8080` (or whatever `PORT` is set to).

## Database Migration

Apply the migration in `sql/V001_create_users.sql` to your database.

Quick option using `psql` inside the container:

```bash
docker exec -i IaaSDemoPostgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" < sql/V001_create_users.sql
```

## API Endpoints

### Create User

- **POST** `/api/v1/user/create`
- Body (JSON):

```json
{
  "name": "Teck",
  "age": 21,
  "department": "Engineering"
}
```

Example:

```bash
curl -i -X POST "http://localhost:8080/api/v1/user/create" \
  -H "Content-Type: application/json" \
  -d '{"name":"Teck","age":21,"department":"Engineering"}'
```

### Get User

- **GET** `/api/v1/user/get?id=1`

Example:

```bash
curl -i "http://localhost:8080/api/v1/user/get?id=1"
```

## Useful Commands

List running containers:

```bash
docker ps
```

Inspect users table:

```bash
docker exec -it IaaSDemoPostgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT * FROM users ORDER BY id;"
```

Stop/remove Postgres container (keeps data volume):

```bash
docker stop IaaSDemoPostgres
docker rm IaaSDemoPostgres
```

Delete the DB volume (WARNING: deletes all data):

```bash
docker volume rm pgdata
```

