# Frontend Planning

Этот документ фиксирует уже утвержденный implementation plan для frontend-а.

## Branch

- Рабочая ветка: `feature/frontend-typescript`
- База: `f34ba7f`

## Approved architecture

```text
services/frontend/src/
  app/
  pages/
  features/
  entities/
  shared/
```

## Vertical slices

### Slice 1. Public MVP

- bootstrap `Vite + React + TypeScript`
- shared layout and design tokens
- OpenAPI-generated types
- routing
- public pages:
  - `Home`
  - `Catalog`
  - `Gift Card`
  - `Login`
  - `Register`
  - `Password Reset Request`

### Slice 2. Recommendation

- recommendation wizard
- recommendation results
- explanations
- alternatives

### Slice 3. User area

- profile
- wishlist overview/detail

### Slice 4. Admin

- catalog import upload
- import status
- import errors

### Slice 5. Integrations

- VK integration screen and flows when backend exposes endpoints

## UX principles

- сценарий-first, а не page-first
- тёплая light-тема
- editorial commerce visual direction
- hero + catalog discovery на главной
- утилитарный, но не безликий catalog

## Known backend gaps

- recommendation flow пока `auth-only`
- VK HTTP endpoints отсутствуют в текущем OpenAPI
- recommendation contract не содержит `gender`
- gift detail содержит один `store_link`, а не список магазинов
- refine/filter recommendation results сервером пока не поддержан
