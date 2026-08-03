# Scrutatio Online — Bijbel lezen met kruisverwijzingen

Een **statische website / PWA** voor het lezen van de Nederlandse Katholieke Bijbel
(Willibrordvertaling 1975) met meer dan 340.000 kruisverwijzingen. Er is **geen
backend nodig**: alle data is statische JSON die direct door de browser wordt
geladen.

Dit project was oorspronkelijk een Go API met daarnaast een Svelte-frontend. Omdat
alle data statisch is (embedded JSON), is de Go backend volledig geschrapt en in
de frontend geïntegreerd.

## Projectstructuur

```
frontend/
  public/data/            # Alle statische data (lazy geladen)
    books.json            # Boekenlijst (id, naam, volgorde)
    books/*.json          # Volledige tekst per boek
    crossref/*.json       # Kruisverwijzingen + mapping per boek
  src/
    App.svelte            # Hoofdapplicatie
    lib/bibleData.js      # Data-aanroepen + tekst-opschoning (cleanVerseText)
```

De `cleanVerseText`-logica (HTML-entities decoderen, voetnoot/`<abbr>`-markers en
tags strippen, whitespace normaliseren) die vroeger in Go zat, is nu een exacte
port in `frontend/src/lib/bibleData.js`. Van de 35.541 verzen is de uitvoer
geverifieerd identiek aan de oorspronkelijke Go backend.

## Lokaal draaien

```bash
cd frontend
npm install          # eerste keer alleen
npm run dev          # dev-server op http://localhost:5173
```

## Productie-build

```bash
cd frontend
npm run build        # schrijft naar dist/ (inclusief data)
npm run preview      # lokaal de productie-build bekijken
```

## Hosting

Omdat het een pure statische site is, kan deze overal draaien waar statische
bestanden gehost worden.

### GitHub Pages (standaard setup)

De repo bevat een workflow (`.github/workflows/deploy-pages.yml`) die bij elke
push naar `main` de frontend bouwt en naar GitHub Pages deployt.

Eénmalig instellen (Settings → Pages → Build and deployment):
- **Source:** "GitHub Actions" (niet "Deploy from a branch").

Daarna staat de app op `https://<user>.github.io/bijbel-scrutineren/`. De build
gebruikt een relatief `base`, dus hij werkt ook op een root-domein of met een
custom domain via `Settings → Pages → Custom domain`.

> Let op: de data is ~90 MB. Dat is prima voor bovenstaande hosts, maar houd er
> rekening mee dat de eerste keer laden per boek groot is (de app laadt boeken
> en kruisverwijzingen **lazy** en cachet ze daarna in de PWA/service worker).

## PWA

De app is een installable Progressive Web App (via `vite-plugin-pwa`):
- Statische JSON onder `/data/` wordt na het eerste bezoek gecached (CacheFirst),
  zodat bijbeltekst offline leesbaar is.
- Eventuele data-updates vereisen een SW-versiebump (workbox precache).

## Technologie

- **Frontend:** Svelte, Vite, vite-plugin-pwa
- **Data:** statische JSON (Willibrordvertaling 1975 + OpenBible.info kruisverwijzingen)
- **Native wrapper:** Capacitor (iOS) — optioneel
