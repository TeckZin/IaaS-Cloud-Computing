# Server (Go + PostgreSQL) — Local + EC2 + ECR

This repo contains a Go HTTP API server with PostgreSQL persistence, plus Docker workflows for:
- Local development (Docker Postgres)
- Containerizing the API (Dockerfile)
- Pushing/pulling the API image to/from Amazon ECR
- Running the API + Postgres on an EC2 instance

---

## Project Structure

```
server/
├─ api/
├─ main/
│  ├─ main.go
│  └─ routes.go
├─ models/
│  └─ userModel.go
├─ sql/
│  └─ V001_create_users.sql
├─ storage/
│  └─ UserStorage.go
├─ .env
├─ Dockerfile
├─ go.mod
└─ go.sum
```

---

## Environment Variables

Example `.env` (local or EC2):

```env
PORT=8080
POSTGRES_USER=appuser
POSTGRES_PASSWORD=apppass
POSTGRES_DB=appdb
POSTGRES_PORT=5432

# IMPORTANT:
# - If API runs on your laptop and Postgres is Docker on your laptop: POSTGRES_HOST=localhost
# - If API runs in Docker and Postgres runs in Docker: POSTGRES_HOST=<postgres container name>
# - If using RDS: POSTGRES_HOST=<rds endpoint>
```

---

## Conenction to EC2
```bash
chmod 400 your-key.pem
ssh -i your-key.pem USERNAME@PUBLIC_DNS_OR_IP
```

## Local: Run PostgreSQL in Docker

```bash
docker volume create pgdata

docker run -d   --name IaaSDemoPostgres   --env-file .env   -p 5432:5432   -v pgdata:/var/lib/postgresql/data   postgres:16
```

Run migration (if file exists locally in `server/sql/`):

```bash
docker exec -i IaaSDemoPostgres psql -U appuser -d appdb < sql/V001_create_users.sql
```

---

## Local: Run the API (without Docker)

Go does not auto-load `.env`. Load it into your current shell:

```bash
cd server
set -a; source .env; set +a
go run ./main
```

---

## API Endpoints

### Create user
```bash
curl -i -X POST "http://localhost:8080/api/v1/user/create"   -H "Content-Type: application/json"   -d '{"name":"Teck","age":21,"department":"Engineering"}'
```

### Get user by id
```bash
curl -i "http://localhost:8080/api/v1/user/get?id=1"
```

---

## Dockerize the API (build locally)

From `server/` (where `Dockerfile` is):

```bash
docker build -t iaas-demo-server:latest .
```

Run the API container:

```bash
docker run -d   --name iaas-demo-server   --env-file .env   -p 8080:8080   iaas-demo-server:latest
```

---

## Amazon ECR (Push / Pull)

ECR stores your **API image**. `docker-compose.yml` is not pushed to ECR.

Repo used in this setup:

```
202533502060.dkr.ecr.us-east-1.amazonaws.com/iaas_demo_server
```

### Login to ECR (laptop)
```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 202533502060.dkr.ecr.us-east-1.amazonaws.com
```

### Build + push (recommended: multi-arch)
This avoids Apple Silicon (arm64) vs EC2 (amd64) mismatch.

```bash
cd server
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64   -t 202533502060.dkr.ecr.us-east-1.amazonaws.com/iaas_demo_server:latest   --push .
```

### Pull on EC2
On EC2, the instance should have an IAM Role with `AmazonEC2ContainerRegistryReadOnly`.

```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 202533502060.dkr.ecr.us-east-1.amazonaws.com

docker pull 202533502060.dkr.ecr.us-east-1.amazonaws.com/iaas_demo_server:latest
```

---

## EC2: Run API + Postgres (both in Docker)

### 1) Create a Docker network
```bash
docker network create iaas-net
```

### 2) Run Postgres container (on the same network)
```bash
docker volume create pgdata

docker run -d   --name iaas-demo-postgres   --network iaas-net   -e POSTGRES_USER=appuser   -e POSTGRES_PASSWORD=apppass   -e POSTGRES_DB=appdb   -v pgdata:/var/lib/postgresql/data   postgres:16
```

### 3) Run migration (inline, no file needed)
```bash
docker exec -i iaas-demo-postgres psql -U appuser -d appdb <<'SQL'
CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  age INTEGER NOT NULL CHECK (age >= 0),
  department TEXT NOT NULL
);
SQL
```

### 4) Run API container from ECR (same network)
Map EC2 port 80 -> container 8080:

```bash
docker run -d   --name iaas-demo-server   --network iaas-net   -e PORT=8080   -e POSTGRES_HOST=iaas-demo-postgres   -e POSTGRES_PORT=5432   -e POSTGRES_USER=appuser   -e POSTGRES_PASSWORD=apppass   -e POSTGRES_DB=appdb   -p 80:8080   --restart unless-stopped   202533502060.dkr.ecr.us-east-1.amazonaws.com/iaas_demo_server:latest
```

### 5) Test on EC2
```bash
curl -i -X POST "http://localhost/api/v1/user/create"   -H "Content-Type: application/json"   -d '{"name":"Alice","age":22,"department":"HR"}'

curl -i "http://localhost/api/v1/user/get?id=1"
```

### 6) Test from your laptop (public)
Replace with your EC2 public IP:

```bash
curl -i -X POST "http://EC2_PUBLIC_IP/api/v1/user/create"   -H "Content-Type: application/json"   -d '{"name":"Bob","age":25,"department":"Sales"}'

curl -i "http://EC2_PUBLIC_IP/api/v1/user/get?id=1"
```

EC2 Security Group must allow inbound **HTTP (80)** from `0.0.0.0/0`.

---

## Optional: docker-compose for Postgres + migrations

Create `docker-compose.yml` in `server/`:

```yaml
services:
  postgres:
    image: postgres:16
    container_name: iaas-demo-postgres
    env_file:
      - .env
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB"]
      interval: 5s
      timeout: 5s
      retries: 20

  migrate:
    image: postgres:16
    container_name: iaas-demo-migrate
    env_file:
      - .env
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./sql:/migrations:ro
    entrypoint: ["/bin/sh", "-c"]
    command: >
      for f in /migrations/*.sql; do
        echo "running $$f";
        psql "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@postgres:5432/$$POSTGRES_DB?sslmode=disable" -f "$$f";
      done

volumes:
  pgdata:
```

Run:
```bash
docker compose up -d postgres
docker compose run --rm migrate
```

---

## Useful Docker Commands

Check containers:
```bash
docker ps
docker ps -a
```

Logs:
```bash
docker logs -f iaas-demo-server
docker logs -f iaas-demo-postgres
```

Remove containers:
```bash
docker rm -f iaas-demo-server iaas-demo-postgres
```

Remove all containers/images (dangerous):
```bash
docker rm -f $(docker ps -aq)
docker rmi -f $(docker images -aq)
```

Exit EC2 SSH session:
```bash
exit
```

