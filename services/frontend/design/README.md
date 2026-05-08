# Frontend Design Mockups

Static, Figma-ready final-site mockups for the Gift Suggestion frontend. These artifacts are design review materials only: they do not connect to backend, do not add runtime dependencies, and do not change production frontend code.

## Design Concept

The mockups follow the current frontend visual language:

- Warm light background: `#f6f1e8`
- Surface cards: `#fffdf8`
- Primary action: `#c65a1e`
- Secondary action: `#245c4a`
- Accent: `#cbae63`
- Borders: `#d8c8b0`
- Editorial commerce tone with serif headings, rounded cards, catalog grids, gift cards, auth form blocks, banners and skeleton states.

Typography:

- Headings: `Fraunces, Georgia, serif`
- Body/UI: `Manrope, Segoe UI, sans-serif`

Each screen is a fixed `1440x1024` SVG frame with no external images, fonts, local file links or backend calls.

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
4. Keep each imported SVG at `1440x1024`.
5. Place frames in the specified order and use each SVG file name as the frame caption.

## Notes on Current Implementation

The mockups were created before full VK integration was implemented. The following clarifications apply when comparing mockups to the live app:

- **VK connect/sync** — эндпоинты реализованы, реальный импорт групп через `groups.get` работает при `VK_ENABLED=true`. Mockup `11-integrations-vk.svg` отражает UI-концепцию; живая панель в профиле функциональна.
- **Gift detail offers** — backend возвращает несколько offers (`gift_offers`) на подарок; mockup показывал одну кнопку.
- **Recommendation** — поле `gender` реализовано в backend и frontend.
- **Wishlist flags** — сохранение в wishlist доступно из каталога и карточки подарка через кнопку.

## Regeneration

The SVG and HTML artifacts can be regenerated with:

```bash
node services/frontend/design/tools/render-static-mockups.mjs
```
