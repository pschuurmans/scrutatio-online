# Scrutatio Online — Bijbel lezen met kruisverwijzingen

Een **statische website / PWA** voor het lezen van de Nederlandse Katholieke Bijbel
(Willibrordvertaling 1975) met kruisverwijzingen. Er is **geen backend nodig**:
alle data is statische JSON die direct door de browser wordt geladen.

De app is bedoeld om **naast een fysieke Bijbel** te gebruiken tijdens de scrutatio:
niet om te lezen, maar om kruisverwijzingen te vinden bij verzen die je interesseert.
De kernloop is **vers → kruisverwijzingen → naar verwante verzen**.

Dit project was oorspronkelijk een Go API met daarnaast een Svelte-frontend. Omdat
alle data statisch is (embedded JSON), is de Go backend volledig geschrapt en in
de frontend geïntegreerd.

## Projectstructuur

```
frontend/
  public/data/            # Alle statische data (lazy geladen)
    books.json            # Boekenlijst (id, naam, volgorde)
    books/*.json          # Volledige tekst per boek
    jb/                   # Kruisverwijzingen (Jeruzalem Bijbel 1966)
      index.json          # Engelse afkorting -> kruisverwijzingsbestand
      book-mapping.json   # Engelse afkorting -> Nederlands boek-id
      *.json              # Kruisverwijzingen per boek
  src/
    App.svelte            # Hoofdapplicatie
    lib/bibleData.js      # Data-aanroepen + tekst-opschoning (cleanVerseText)
```

Kruisverwijzingen komen uitsluitend uit de Jeruzalem Bijbel (1966) extractie onder
`public/data/jb/`. De oudere OpenBible-bron (`public/data/crossref/`) staat nog in
de repo ter referentie maar wordt niet meer geladen.

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

## Up-to-date houden & validatie (CI)

Geautomatiseerde dependency-updates kunnen de app breken; daarom is er een
test- en CI-laag:

- **Tests** (`cd frontend && npm run test`, Vitest + jsdom):
  - `src/lib/bibleData.test.js` — unit-tests voor de kernlogica
    (`cleanVerseText`, hoofdstuk-laden, kruisverwijzings-mapping).
  - `src/lib/dataIntegrity.test.js` — controleert dat de statische JSON onder
    `public/data/` onderling consistent is (elk boek heeft een bestand, alle
    verzen in bereik, elke kruisverwijzing heeft een geldige Engelse→Nederlandse
    mapping). Vangt beschadigde of onvolledige data-updates.
- **CI** (`.github/workflows/ci.yml`) — draait `npm ci`, `npm run test` en
  `npm run build` op elke **PR** en op pushes naar `main`. Een gebroken
  dependency-update wordt hier rood en kan niet mergen.
- **Dependabot** (`.github/dependabot.yml`) — opent wekelijks PR's voor
  npm-pakketten en GitHub Actions. De CI-check bepaalt of ze veilig zijn.

> **Aanbeveling:** stel een branch protection rule in op `main` die deze
> "test-and-build"-check **vereist** vóór mergen. Dan kan een automatische
> dependency-update (of welke PR dan ook) nooit stilletjes een kapotte app
> naar main/deploy brengen.

## Technologie

- **Frontend:** Svelte, Vite, vite-plugin-pwa
- **Data:** statische JSON (Willibrordvertaling 1975 + Jeruzalem Bijbel 1966 kruisverwijzingen)
- **Native wrapper:** Capacitor (iOS) — optioneel
