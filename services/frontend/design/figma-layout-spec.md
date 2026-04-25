# Figma Layout Spec

Use this spec to place the static mockups into a Figma review page.

## Page

Page name: `Frontend Final Mockups`

Canvas order:

1. `00 Cover / Notes`
2. `01 Public`
3. `02 Auth`
4. `03 Recommendation`
5. `04 User Area`
6. `05 Admin`
7. `06 States`
8. `07 Mobile Companion Frames`

## Frame Sizes

- Desktop frames: `1440x1024`
- Mobile companion frames: `390x844`
- Section spacing: `160px` horizontal, `140px` vertical
- Caption spacing below imported SVG: `24px`

## Desktop Frame Order

Place desktop SVGs left to right, then wrap to the next row.

### 01 Public

1. `01-home.svg` - Home
2. `02-catalog.svg` - Catalog
3. `03-gift-detail.svg` - Gift Detail

### 02 Auth

1. `04-login.svg` - Login
2. `05-register.svg` - Register
3. `06-password-reset.svg` - Password Reset Request

### 03 Recommendation

1. `07-recommendation-wizard.svg` - Recommendation Wizard
2. `08-recommendation-results.svg` - Recommendation Results

### 04 User Area

1. `09-profile.svg` - Profile
2. `10-wishlist.svg` - Wishlist
3. `11-integrations-vk.svg` - Integrations / VK

### 05 Admin

1. `12-admin-import.svg` - Admin Import
2. `13-import-status-errors.svg` - Import Status / Errors

### 06 States

1. `14-system-states.svg` - Loading / Empty / Error / Success / Unauthorized / Not Found

## Suggested Coordinates

Start at `x=0, y=0` inside each section.

- Frame 1: `x=0, y=0`
- Frame 2: `x=1600, y=0`
- Frame 3: `x=3200, y=0`
- New row: add `y=1220`

For captions, add text below each imported SVG:

- Font: `Manrope SemiBold`
- Size: `24`
- Color: `#1f1a14`

## Mobile Companion Frames

Create mobile frames at `390x844` for review of responsive behavior. These are companion frames derived from the desktop mockups, not separate API screens.

Recommended mobile order:

1. `M01 Home`
2. `M02 Catalog`
3. `M03 Gift Detail`
4. `M04 Login`
5. `M05 Recommendation Wizard`
6. `M06 Wishlist Empty`
7. `M07 Admin Import Status`

Mobile layout rules:

- Header stacks brand, nav and auth action vertically.
- Page horizontal padding: `16px`.
- Catalog and wishlist grids collapse to one column.
- Gift detail image moves above content.
- Auth layout becomes intro text above form.
- Admin status table becomes stacked rows.
- Primary CTA remains visible before secondary actions.

## Import Notes

Drag files from `services/frontend/design/mockups/` into the Figma canvas. Keep imported SVG dimensions at `1440x1024`. Do not scale mockups non-proportionally.

The mockups intentionally preserve backend limitations:

- VK actions are disabled/planned.
- Recommendation submit requires auth.
- Gift detail has one purchase CTA from `store_link`.
- No invented wishlist flag is shown in catalog cards.
