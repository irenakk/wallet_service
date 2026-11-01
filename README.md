# wallet_service

A small Go backend from early 2025 for user wallets: you register, log in to get a JWT, top up your balance and send money to another user by username. It also runs a gRPC `AuthService` with one method, `Validate`, so other services can check tokens issued here.

Stack: Go 1.23, gorilla/mux, sqlx + lib/pq, PostgreSQL 15 (Docker Compose), dgrijalva/jwt-go, bcrypt, gRPC/protobuf.

## How to run

Start Postgres:

```bash
docker-compose up -d
```

Then run the service from the repo root:

```bash
go run .
```

The service connects to Postgres on `localhost:5432` with the user and database from `docker-compose.yml`. The `users` and `wallets` tables are created on startup with `CREATE TABLE IF NOT EXISTS`, there are no migration files.

Ports:

- `8088` HTTP API
- `50059` gRPC

## HTTP API

| Method | Path | Auth | Body |
|---|---|---|---|
| POST | `/register` | no | `{"username": "alice", "password": "..."}` |
| POST | `/login` | no | same as register, returns `{"token": "..."}` |
| POST | `/api/deposit` | Bearer token | `{"amount": 100}` |
| POST | `/api/transfer` | Bearer token | `{"to_username": "bob", "amount": 25.5}` |

Registration hashes the password with bcrypt and creates an empty wallet for the new user. Tokens are HS256 JWTs with the username in `sub` and expire after 5 minutes. A transfer takes the amount from the sender only if `balance >= amount` and credits the recipient in the same transaction.

Example:

```bash
curl -X POST localhost:8088/register -d '{"username":"alice","password":"pass123"}'
curl -X POST localhost:8088/login -d '{"username":"alice","password":"pass123"}'
# {"token":"<jwt>"}
curl -X POST localhost:8088/api/deposit -H 'Authorization: Bearer <jwt>' -d '{"amount":100}'
```

## gRPC

The proto is `docs/grpc/proto/sso.proto` (package `auth`), generated Go code is in `docs/grpc/gen`.

```proto
service AuthService {
  rpc Validate(ValidateRequest) returns (ValidateResponse);
}
```

`Validate` takes a `token` and returns the `login` (username) if the token is valid.

## Project structure

```
main.go               DB connection, schema init, starts the app
domain/controller/    HTTP handlers, JWT creation, gRPC server
domain/service/       user and wallet logic, bcrypt
domain/repository/    SQL via sqlx
domain/entity/        User, Wallet
middleware/           Bearer token check for /api routes
docs/grpc/            proto and generated code
```
