# shophub

ShopHub is the platform's central web application, made up of a front end and a back end. It is the administrative panel through which a user registers, logs in, and manages their shop sites: creating new ones, changing their configuration, and deleting them. The ShopHub back end does not run shops itself. Instead, for every shop it creates a `Shop` (and an associated `Wallet`) custom resource in the Kubernetes cluster. The actual deployment (Deployment, Service, Ingress, database) is handled by `shop-operator`, which reacts to those custom resources. ShopHub tracks a shop's status by reading `.status.phase` from the custom resource and maps it to a status shown to the user, such as pending, deploying, running, or error.

## Role in the architecture

The user creates a shop in the ShopHub UI by setting its name, availability (standard maps to 2 replicas, high maps to 3 replicas), the wallet address that receives payments, and the database type (postgres or redis). The back end writes the shop's metadata into its own PostgreSQL database and, at the same time, creates the `Shop` and `Wallet` custom resources in the cluster through the Kubernetes API using a client-go dynamic client. The user can open their shop's site in a new tab, edit its configuration (availability, wallet address, database), or delete it. Deleting a shop also removes the custom resources, so `shop-operator` cleans up all related Kubernetes resources. If the cluster is not reachable, the back end falls back to a no-op orchestrator: the REST API still works, but no custom resources are created. This is useful for local development without a cluster.

## Repository structure

The `cmd/shophub/main.go` file is the entry point of the back end service. `internal/auth/` handles JWT access and refresh tokens and password hashing with bcrypt. `internal/httpapi/` contains HTTP handlers and middleware for auth, shops, and health checks, built on the standard `net/http` router without an external web framework. `internal/shops/` is the service layer and GORM repository for shops, including ownership checks, with `orchestrator_noop.go` as the clusterless fallback. `internal/k8s/` is the client-go orchestrator that creates, updates, and deletes `Shop` and `Wallet` custom resources in the `shop.shophub.io/v1alpha1` group and reads their status. `internal/models/` holds the GORM models for `User` and `Shop`. `internal/database/` manages the connection and migrations using golang-migrate, embedded into the binary through `embed.FS`. `internal/health/` implements liveness and readiness checks. `api/openapi.yaml` is the OpenAPI 3.0 specification of the REST API. `web/` is the front end application, built with Angular 17 and standalone components.

## Main features

The back end REST API, under `/api/v1/...`, provides registration, login, refresh, and logout with JWT and versioned refresh tokens for invalidation, full CRUD over a user's shops at `/api/v1/shops` kept in sync with the `Shop` and `Wallet` custom resources in the cluster, retrieval of a shop's public URL, and health and readiness endpoints.

The front end, under `web/`, provides user registration and login with JWT tokens stored in local storage through `AuthService`, `auth.interceptor.ts`, and `auth.guard.ts`, a dashboard listing all of a user's shops, a wizard for creating a new shop (name, availability, wallet address, database type), a detail and edit view for a single shop that supports changing configuration, deleting the shop, opening its site in a new tab, and showing provisioning status, and state management through NgRx stores for auth, shops, and UI concerns with their associated effects.

## Technical stack

The back end is written in Go using the standard `net/http` package without an external framework, GORM with PostgreSQL, golang-jwt, client-go and controller-runtime for talking to the Kubernetes API, golang-migrate for schema migrations, and Testcontainers together with envtest for integration testing. The front end is built with Angular 17 using standalone components and lazy loaded routes, Angular Material, NgRx (store and effects), RxJS, and TypeScript, tested with Playwright for end to end tests and Karma/Jasmine for unit tests. Continuous integration runs on GitHub Actions and includes commitlint for Conventional Commits, Go vet, test, and build, integration tests with Testcontainers and envtest under a coverage gate of at least 60 percent, and building and pushing a Docker image to GHCR when a version tag is created, following Semantic Versioning. In deployment, the application is installed into the cluster through the `shophub` Helm chart from the `helm-charts` repository, and the state of that installation is tracked through the `kube-state` repository via an ArgoCD Application.

## Running locally

```sh
docker compose up -d db          # PostgreSQL on :5432
cp .env.example .env             # set JWT_SECRET
go mod tidy
go run ./cmd/shophub              # AutoMigrate creates the tables
```

For testing, run `go test ./...` for unit tests, `go test -tags=integration ./internal/httpapi/ -run Integration -v` for integration tests (requires Docker), and `go test -tags=integration ./internal/k8s/ -run Envtest -v` for the Kubernetes envtest layer. Detailed commands for coverage and an end to end check (register, log in, create a shop, verify the Shop custom resource in the cluster) can be found in the CI pipeline definition at `.github/workflows/ci.yml`.
