# shophub

ShopHub backend — web aplikacija za upravljanje sajtovima prodavnica.

Modul: `github.com/shophub-platform/shophub`

## Faza F1 — Član 1: Autentifikacija + GORM modeli

Implementirano u ovoj fazi (spec. 1.1 Prijava i registracija, 1.2 Upravljanje sajtovima):

- **JWT auth servis** (`internal/auth`, `internal/httpapi`): registracija, login, refresh
  (+ logout sa poništavanjem refresh tokena, i `GET /me`). Lozinke su bcrypt-heširane.
  Access token traje 15 min, refresh 7 dana; refresh token nosi `version` koji se poredi
  sa `User.RefreshTokenVersion`, pa logout poništava postojeće refresh tokene.
- **GORM modeli** (`internal/models`): `User` i `Shop` u PostgreSQL bazi
  (`internal/database`, `AutoMigrate` na startu + `migrations/0001_init.sql`).

HTTP sloj koristi standardni `net/http` (Go 1.22 routing), bez eksternog framework-a,
u skladu sa postojećim stilom repozitorijuma.

## Endpoint-i

| Metoda | Putanja | Auth | Opis |
|--------|---------|------|------|
| POST | `/api/v1/auth/register` | – | Kreira nalog, vraća par tokena |
| POST | `/api/v1/auth/login` | – | Autentifikacija, vraća par tokena |
| POST | `/api/v1/auth/refresh` | – | Razmena refresh tokena za novi par |
| POST | `/api/v1/auth/logout` | Bearer | Poništava refresh tokene |
| GET  | `/api/v1/me` | Bearer | Trenutni identitet |
| GET  | `/api/v1/shops` | Bearer | Lista korisnikovih prodavnica |
| POST | `/api/v1/shops` | Bearer | Kreira Shop CR + meta-podatke u DB |
| GET  | `/api/v1/shops/{id}` | Bearer | Detalji + status (čita CR phase) |
| PATCH | `/api/v1/shops/{id}` | Bearer | Menja availability/wallet/database |
| DELETE | `/api/v1/shops/{id}` | Bearer | Briše Shop CR + meta-podatke |
| GET  | `/api/v1/shops/{id}/url` | Bearer | Javni URL prodavnice |
| GET  | `/healthz`, `/readyz` | – | Health/readiness |

## Faza F2 — ShopHub backend + Kubernetes (Shop CR)

- **Shops REST sloj** (`internal/shops`, `internal/httpapi`): servisni sloj
  (`Service`) + GORM repozitorijum (`Repository`) sa proverom vlasništva nad
  svakim resursom. CRUD nad meta-podacima u PostgreSQL-u i sinhronizacija sa
  Kubernetes klasterom.
- **client-go orkestrator** (`internal/k8s`): dinamički klijent kreira/ažurira/
  briše `Shop` CR-ove (grupa `shop.shophub.io/v1alpha1`) i čita `.status.phase`.
  Koristi in-cluster config u podu, a kubeconfig lokalno. Status CRD faza se
  mapira na javni status: `Ready→running`, `Provisioning→deploying`,
  `Failed→error`, ostalo `pending`.
- **Bezbedan fallback**: ako klaster nije dostupan, server radi sa no-op
  orkestratorom (REST API funkcioniše, CR-ovi se ne kreiraju).
- Ime Shop CR-a je determinističko: `shop-<uuid>`; namespace, podrazumevana
  slika i URL šablon se konfigurišu env varijablama (vidi `.env.example`).
- **Migracije**: `golang-migrate` sa SQL fajlovima u `migrations/`
  (`0001_init.up.sql` / `.down.sql`), ugrađenim u binarni fajl preko `embed.FS`
  i pokrenutim na startu (`database.Migrate`). CLI alternativa: `make migrate-up`.
- **API dokumentacija**: `swag` anotacije (`// @...`) u handlerima; OpenAPI se
  generiše sa `make swagger` (`swag init`) pored ručno održavanog `api/openapi.yaml`.

## Pokretanje

```sh
docker compose up -d db          # PostgreSQL na :5432
cp .env.example .env             # podesi JWT_SECRET
go mod tidy                      # povuče zavisnosti, generiše go.sum
go run ./cmd/shophub             # AutoMigrate napravi tabele
```

## Testiranje

### 1. Unit testovi (bez Docker-a i klastera)

```sh
go mod tidy                       # JEDNOM: povlači client-go, testify, testcontainers...
go vet ./...
go test ./... -count=1            # auth, modeli, shops servis (mock), shop handleri (mock)
```

### 2. Integracioni testovi (zahteva Docker za PostgreSQL Testcontainers)

```sh
go test -tags=integration ./internal/httpapi/ -run Integration -v
```

### 3. K8s integracioni test (envtest — in-memory API server)

```sh
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@release-0.18
export KUBEBUILDER_ASSETS=$(setup-envtest use -p path 1.30.0)
go test -tags=integration ./internal/k8s/ -run Envtest -v
```

### 4. Coverage (DoD: min 60%)

```sh
go test -tags=integration -coverpkg=./internal/...,./pkg/... \
  -coverprofile=cover.out ./...
go tool cover -func=cover.out | tail -1     # prikazuje total %
go tool cover -html=cover.out               # HTML izveštaj u browseru
```

## Definition of Done — F2: end-to-end provera

```sh
# (a) Podigni bazu i (opciono) lokalni klaster sa instaliranim Shop CRD-om.
docker compose up -d db
# kind create cluster        # ili minikube/k3d
# kubectl apply -f ../shop-operator/config/crd/bases/   # instaliraj Shop CRD

cp .env.example .env
go mod tidy
go run ./cmd/shophub         # AutoMigrate napravi tabele; loguje da li je K8s aktivan

# (b) register -> login -> validan JWT
curl -s -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"a@b.com","password":"supersecret","displayName":"Vesna"}'

ACCESS=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"a@b.com","password":"supersecret"}' | jq -r .accessToken)

# (c) kreiraj prodavnicu -> Shop CR
curl -s -X POST localhost:8080/api/v1/shops \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"name":"moja-radnja","availability":"high","walletAddress":"0xabc","databaseType":"postgres"}'

# (d) Shop CR mora biti vidljiv u klasteru
kubectl get shops

# lista / detalji / url / brisanje
curl -s localhost:8080/api/v1/shops -H "Authorization: Bearer $ACCESS"
curl -s localhost:8080/api/v1/shops/<ID> -H "Authorization: Bearer $ACCESS"
curl -s localhost:8080/api/v1/shops/<ID>/url -H "Authorization: Bearer $ACCESS"
curl -s -X PATCH localhost:8080/api/v1/shops/<ID> \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"availability":"standard"}'
curl -s -X DELETE localhost:8080/api/v1/shops/<ID> -H "Authorization: Bearer $ACCESS" -i
```

## Docker image (CI tag v0.1.0)

```sh
docker build --build-arg VERSION=v0.1.0 -t ghcr.io/shophub-platform/shophub:v0.1.0 .
```

CI (`.github/workflows/ci.yml`) ima `build-test` (vet + unit + build),
`integration` (Testcontainers + envtest + coverage gate ≥60%) i `docker`
(build + tag `v0.1.0`, push na GHCR pri `v*` tagu).