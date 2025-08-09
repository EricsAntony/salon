# user-service

Golang microservice for salon platform user auth/registration with phone OTP, JWT, PostgreSQL, and Kubernetes-ready deployment.

## Features
- Phone OTP login (request, register, authenticate)
- JWT access/refresh tokens (refresh stored hashed in DB)
- PostgreSQL with migrations (golang-migrate)
- Zerolog structured logs (Loki-friendly)
- Chi HTTP router, request ID, rate limiting
- Dockerfile + K8s manifests

## Project Structure
See directories:
- `cmd/` main entrypoint
- `internal/api/` HTTP handlers and middleware
- `internal/service/` business logic
- `internal/repository/` Postgres access (pgx)
- `internal/model/` entities
- `internal/auth/` JWT + OTP helpers
- `internal/config/` Viper config loader
- `pkg/logger/` Zerolog init
- `pkg/utils/` helpers
- `migrations/` SQL migrations
- `deployments/k8s/` manifests

## Config
- Config is loaded from `configs/config.yaml` and env vars with prefix `USER_SERVICE_`.
- Copy `.env.example` to `.env` for local development. Viper loads env automatically.

## Endpoints
- `POST /otp/request` -> body `{ "phone_number": "+18005551234" }`
- `POST /user/register` -> body `{ phone_number, name, gender: male|female|other, email?, location?, otp }` returns `{ user, access_token, refresh_token }`
- `POST /user/authenticate` -> body `{ phone_number, otp }` returns `{ access_token, refresh_token }`
- `GET /user/{id}` -> protected by Bearer access token

Note: For demo, OTP codes are logged. Integrate an SMS provider in production.

## Run locally
1. Start Postgres and create DB `salon_users`.
2. Export env or create `.env` (see example).
3. Run migrations:
   ```sh
   make migrate-up MIGRATE_DB_URL="postgres://user:password@localhost:5432/salon_users?sslmode=disable"
   ```
4. Build and run:
   ```sh
   make tidy && make run
   ```

## Docker
```sh
make docker
```

## Kubernetes
- Apply `deployments/k8s/configmap.yaml` and create a Secret `user-service-secrets` with keys: `db_url`, `jwt_access_secret`, `jwt_refresh_secret`.
- Apply `deployments/k8s/deployment.yaml` and `deployments/k8s/service.yaml`.

## Testing
```sh
make test
```

## Security Notes
- Use HTTPS (terminate at ingress/controller).
- Keep JWT secrets in environment variables/Secrets.
- Rate limiting applied to OTP requests.
- Validate inputs on all endpoints.
