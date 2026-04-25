# Frontend Design Mockups

Static, Figma-ready final-site mockups for the Gift Suggestion frontend. These artifacts are design review materials only: they do not connect to backend, do not add runtime dependencies, and do not change production frontend code.

## Design Concept

The mockups follow the current Slice 1 frontend visual language:

- Warm light background: `#f6f1e8`
- Surface cards: `#fffdf8`
- Primary action: `#c65a1e`
- Secondary action: `#245c4a`
- Borders: `#d8c8b0`
- Editorial commerce tone with serif headings, rounded cards, catalog grids, gift cards, auth form blocks, banners and skeleton states.

Typography mirrors the frontend CSS intent:

- Headings: `Fraunces, Georgia, serif`
- Body/UI: `Manrope, Segoe UI, sans-serif`
- Numeric emphasis follows the current compact UI treatment.

## Covered Screens

- `01-home.svg`
- `02-catalog.svg`
- `03-gift-detail.svg`
- `04-login.svg`
- `05-register.svg`
- `06-password-reset.svg`
- `07-recommendation-wizard.svg`
- `08-recommendation-results.svg`
- `09-profile.svg`
- `10-wishlist.svg`
- `11-integrations-vk.svg`
- `12-admin-import.svg`
- `13-import-status-errors.svg`
- `14-system-states.svg`

Open the preview page:

```text
services/frontend/design/index.html
```

## Figma Import

1. Create or open the Figma page named `Frontend Final Mockups`.
2. Create the sections from `figma-layout-spec.md`.
3. Drag SVG files from `services/frontend/design/mockups/` onto the Figma canvas.
4. Place frames in the specified order and use each SVG file name as the frame caption.
5. For mobile review, create `390x844` companion frames using the layout notes in the spec.

SVG files are pure static artifacts, so they can be opened in a browser and imported into Figma without running the app.

## Backend Gaps Reflected In UI

- Recommendation submit is auth-only, so the wizard includes an auth limitation note.
- VK connect/sync is shown as a disabled planned state because there are no VK HTTP endpoints in the current OpenAPI/backend code.
- Gift detail uses one purchase CTA because backend exposes one `store_link`, not a store list.
- Recommendation request does not contain `gender`, so the wizard does not render that field.
- Recommendation refine/filtering is not drawn as a working server feature.
- Wishlist saved flags are not assumed inside catalog/recommendation payloads; save actions are presented as explicit wishlist operations.

## Regeneration

The SVG and HTML artifacts can be regenerated with:

```bash
node services/frontend/design/tools/render-static-mockups.mjs
```
