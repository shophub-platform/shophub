# ShopHub Web (Angular 17)

Standalone Angular 17 aplikacija. Eager-loaded `Dashboard`, lazy-loaded `Shops` feature.

## Pokretanje

```bash
cd web
npm install
npm start        # dev server na http://localhost:4200
npm run build    # production build u dist/
npm test         # unit testovi (ChromeHeadless)
```

## Struktura

```
src/app/
├── app.component.ts        # root (eager)
├── app.config.ts           # provideri (router, http)
├── app.routes.ts           # eager Dashboard + lazy Shops
└── pages/
    ├── dashboard/          # eager-loaded
    └── shops/              # lazy-loaded (loadChildren)
```
