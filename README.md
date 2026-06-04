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
| GET  | `/healthz`, `/readyz` | – | Health/readiness |

## Pokretanje

```sh
docker compose up -d db          # PostgreSQL na :5432
cp .env.example .env             # podesi JWT_SECRET
go mod tidy                      # povuče zavisnosti, generiše go.sum
go run ./cmd/shophub             # AutoMigrate napravi tabele
```

## Testiranje

```sh
go vet ./...
go test ./... -count=1           # jwt, lozinke, modeli (bez baze)
```

Primer toka:

```sh
curl -s -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"a@b.com","password":"supersecret","displayName":"Vesna"}'

curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"a@b.com","password":"supersecret"}'

curl -s localhost:8080/api/v1/me -H "Authorization: Bearer <ACCESS_TOKEN>"

curl -s -X POST localhost:8080/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refreshToken":"<REFRESH_TOKEN>"}'
```
