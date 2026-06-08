# shophub-web — ShopHub admin UI (Faza F4, 9.1)

Angular 17 (standalone) + Angular Material + NgRx admin panel za upravljanje
sajtovima prodavnica. Priča sa ShopHub backend-om (`/api/v1`).

## Šta je implementirano (9.1 — Član 1)

- **Login i registracija** — Angular Material kartice, reactive forms, validacija.
- **Dashboard** — kartice prodavnica sa statusom (Provisioning / Ready / Error).
- **Wizard za kreiranje** — 3 koraka (osnovno → wallet → baza) preko `mat-stepper`.
- **Detalj prodavnice** — edit forma + dugme **Open shop** (otvara URL u novom tab-u).
- **Izmena konfiguracije** — availability / wallet / baza → `PATCH /shops/{id}`.
- **Confirm dialog** pre brisanja (`MatDialog`).
- **Toast notifikacije** (`MatSnackBar`), **loading** state-ovi i **error handling**.
- **NgRx store**: `AuthState`, `ShopsState`, `UIState` (actions/reducers/effects/selectors).
- **HTTP interceptor** (9.3): dodaje JWT, retry za GET, centralizovane greške, 401 → /login.
- **Theming** preko CSS varijabli (`src/styles.css`).

## Struktura

```
src/app/
  core/            modeli, servisi (Auth, Shop), interceptor, guard
  store/           NgRx: auth/ shops/ ui/  (+ index.ts root reducers)
  features/        auth/ (login, register), dashboard/, shops/ (wizard, detail)
  shared/          confirm-dialog, status-chip
```

## Pokretanje (development)

Zahteva pokrenut backend na `:8080` (`go run ./cmd/shophub` u root-u repo-a).
`proxy.conf.json` prosleđuje `/api` ka backend-u tokom `ng serve`.

```sh
cd web
npm install            # povlači Angular Material, NgRx, Playwright (treba mreža)
npm start              # ng serve -> http://localhost:4200
```

## Build (produkcija)

```sh
npm run build          # izlaz: dist/shophub-web/browser
```

## Docker / nginx (DoD 9.4)

```sh
docker build -t shophub-web:latest .
docker run --rm -p 8081:80 shophub-web:latest    # http://localhost:8081
```

`nginx.conf` servira SPA (`try_files ... /index.html`) i proksira `/api` ka
backend servisu (`shophub-backend:8080` — prilagodi imenu servisa u deployment-u).

## E2E test (Playwright, DoD 9.4)

Happy-path testovi mokuju API, pa ne zavise od backend-a:

```sh
npx playwright install chromium    # jednom
npm run e2e                         # diže ng serve i pokreće testove
```

Pokriveno: registracija → dashboard, prazan dashboard → poziv na kreiranje,
wizard kreira prodavnicu kroz 3 koraka.

## Lighthouse (DoD 9.4)

Posle `npm run build` i serviranja (nginx ili `npx http-server dist/shophub-web/browser`):

```sh
npx lighthouse http://localhost:8081 --only-categories=performance,accessibility
```

Cilj: Performance i Accessibility > 80.

## Napomene

- Tokeni se čuvaju u `localStorage` (`sh_access`, `sh_refresh`).
- Statusi backend-a (`pending|deploying|running|error`) mapiraju se na prikaz
  (Na čekanju / Provisioning / Ready / Error) u `core/models/shop.model.ts`.
- Stari `pages/` folder iz skeleta se više ne koristi (zamenjen sa `features/`).
