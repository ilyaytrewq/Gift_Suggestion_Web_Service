import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const designDir = join(here, '..');
const mockupsDir = join(designDir, 'mockups');

mkdirSync(mockupsDir, { recursive: true });

const W = 1440;
const H = 1024;
const X = 120;
const CONTENT = 1200;

const C = {
  bg: '#f6f1e8',
  surface: '#fffdf8',
  surfaceAlt: '#fbf5ea',
  muted: '#efe6d6',
  ink: '#1f1a14',
  inkMuted: '#625746',
  primary: '#c65a1e',
  primarySoft: '#f3e3d8',
  primaryHover: '#a94816',
  secondary: '#245c4a',
  secondarySoft: '#dbe9e2',
  accent: '#cbae63',
  accentSoft: '#f1e8ca',
  danger: '#b13a2f',
  dangerSoft: '#f6ddd8',
  success: '#2f7a4a',
  successSoft: '#dcebdd',
  border: '#d8c8b0',
  white: '#ffffff',
};

const gifts = [
  ['Кофейный сет', '2 790 ₽', 'Для дома', ['Кофе, керамика и открытка', 'для уютного старта.']],
  ['Плед из мериноса', '5 490 ₽', 'Уют', ['Мягкий подарок для дома', 'и спокойных вечеров.']],
  ['Книга-альбом', '3 200 ₽', 'Культура', ['Большое издание с историей', 'дизайна и иллюстрациями.']],
  ['Умная колонка', '7 990 ₽', 'Техника', ['Музыка, напоминания', 'и голосовые сценарии.']],
  ['Набор для керамики', '4 100 ₽', 'Хобби', ['Глина, инструменты', 'и понятный стартовый гид.']],
  ['Сертификат SPA', '6 500 ₽', 'Впечатления', ['Подарок-восстановление', 'без выбора размера.']],
];

function esc(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;');
}

function attrs(values) {
  return Object.entries(values)
    .filter(([, value]) => value !== undefined && value !== null && value !== false)
    .map(([key, value]) => `${key}="${esc(value)}"`)
    .join(' ');
}

function rect(x, y, width, height, options = {}) {
  return `<rect ${attrs({
    x,
    y,
    width,
    height,
    rx: options.rx ?? 18,
    fill: options.fill ?? C.surface,
    stroke: options.stroke,
    'stroke-width': options.strokeWidth,
    opacity: options.opacity,
    filter: options.filter,
  })}/>`;
}

function line(x1, y1, x2, y2, options = {}) {
  return `<line ${attrs({
    x1,
    y1,
    x2,
    y2,
    stroke: options.stroke ?? C.border,
    'stroke-width': options.width ?? 1,
    opacity: options.opacity,
  })}/>`;
}

function circle(cx, cy, r, options = {}) {
  return `<circle ${attrs({
    cx,
    cy,
    r,
    fill: options.fill ?? C.muted,
    stroke: options.stroke,
    'stroke-width': options.strokeWidth,
    opacity: options.opacity,
  })}/>`;
}

function text(x, y, lines, options = {}) {
  const items = Array.isArray(lines) ? lines : [lines];
  const size = options.size ?? 20;
  const lineHeight = options.lineHeight ?? Math.round(size * 1.35);
  const family = options.family ?? 'Manrope, Segoe UI, sans-serif';
  const fill = options.fill ?? C.ink;

  return `<text ${attrs({
    x,
    y,
    fill,
    'font-family': family,
    'font-size': size,
    'font-weight': options.weight ?? 500,
    'text-anchor': options.anchor,
  })}>${items
    .map((item, index) => `<tspan x="${x}" dy="${index === 0 ? 0 : lineHeight}">${esc(item)}</tspan>`)
    .join('')}</text>`;
}

function heading(x, y, lines, size = 52) {
  return text(x, y, lines, {
    size,
    lineHeight: Math.round(size * 1.06),
    weight: 700,
    family: 'Fraunces, Georgia, serif',
  });
}

function eyebrow(x, y, label) {
  return text(x, y, label, {
    size: 13,
    weight: 800,
    fill: C.primary,
  });
}

function panel(x, y, width, height, body = '', options = {}) {
  return `<g transform="translate(${x} ${y})">
    ${rect(0, 0, width, height, {
      rx: options.rx ?? 24,
      fill: options.fill ?? C.surface,
      stroke: options.stroke ?? C.border,
      strokeWidth: 1,
      filter: options.shadow === false ? undefined : 'url(#shadow)',
    })}
    ${body}
  </g>`;
}

function button(x, y, width, label, variant = 'primary', options = {}) {
  const height = options.height ?? 50;
  const fills = {
    primary: C.primary,
    secondary: C.secondary,
    ghost: C.surface,
    disabled: C.muted,
    danger: C.danger,
  };
  const color = variant === 'ghost' || variant === 'disabled' ? C.ink : C.white;
  const stroke = variant === 'ghost' || variant === 'disabled' ? C.border : 'transparent';
  return `<g opacity="${options.opacity ?? 1}">
    ${rect(x, y, width, height, {
      rx: 999,
      fill: fills[variant] ?? C.primary,
      stroke,
      strokeWidth: 1,
    })}
    <text ${attrs({
      x: x + width / 2,
      y: y + height / 2 + 6,
      fill: color,
      'font-family': 'Manrope, Segoe UI, sans-serif',
      'font-size': options.size ?? 16,
      'font-weight': 800,
      'text-anchor': 'middle',
    })}>${esc(label)}</text>
  </g>`;
}

function chip(x, y, label, width, variant = 'primary') {
  const fill = variant === 'green' ? C.secondarySoft : variant === 'accent' ? C.accentSoft : variant === 'muted' ? C.muted : C.primarySoft;
  const stroke = variant === 'muted' ? C.border : undefined;
  const color = variant === 'green' ? C.secondary : variant === 'accent' ? '#6d5721' : C.ink;
  return `<g>
    ${rect(x, y, width, 32, { rx: 999, fill, stroke, strokeWidth: stroke ? 1 : undefined })}
    ${text(x + 15, y + 21, label, { size: 13, weight: 800, fill: color })}
  </g>`;
}

function input(x, y, width, label, value, options = {}) {
  const height = options.height ?? 54;
  return `<g>
    ${text(x, y, label, { size: 14, weight: 800 })}
    ${rect(x, y + 14, width, height, { rx: 12, fill: C.surface, stroke: C.border, strokeWidth: 1 })}
    ${text(x + 17, y + 48, value, { size: 15, weight: 500, fill: options.muted ? C.inkMuted : C.ink })}
  </g>`;
}

function banner(x, y, width, lines, variant = 'info') {
  const theme = {
    info: [C.secondarySoft, '#bdd5ca', C.secondary],
    danger: [C.dangerSoft, '#e1b7af', C.danger],
    success: [C.successSoft, '#b9d3be', C.success],
    warm: [C.primarySoft, '#e1c3ad', C.primary],
  }[variant];
  return panel(
    x,
    y,
    width,
    Array.isArray(lines) && lines.length > 1 ? 82 : 62,
    text(24, 38, lines, { size: 16, weight: 800, fill: theme[2], lineHeight: 22 }),
    { rx: 16, fill: theme[0], stroke: theme[1], shadow: false },
  );
}

function photo(x, y, width, height, variant = 0) {
  const colors = [
    [C.accent, C.secondary],
    [C.primary, '#f0c987'],
    ['#8f9e85', C.muted],
    [C.secondary, C.accent],
    ['#b76e45', C.secondarySoft],
    [C.danger, '#f3d5c7'],
  ][variant % 6];
  return `<g>
    ${rect(x, y, width, height, { rx: 20, fill: colors[1] })}
    <path d="M${x} ${y + height * 0.82} C${x + width * 0.25} ${y + height * 0.58}, ${x + width * 0.52} ${y + height * 0.93}, ${x + width} ${y + height * 0.60} L${x + width} ${y + height} L${x} ${y + height} Z" fill="${colors[0]}" opacity="0.78"/>
    ${circle(x + width * 0.78, y + height * 0.24, Math.min(width, height) * 0.12, { fill: C.surface, opacity: 0.78 })}
    ${rect(x + width * 0.18, y + height * 0.30, width * 0.32, height * 0.32, { rx: 14, fill: C.surface, opacity: 0.74 })}
    ${line(x + width * 0.34, y + height * 0.30, x + width * 0.34, y + height * 0.62, { stroke: colors[0], width: 5, opacity: 0.85 })}
    ${line(x + width * 0.18, y + height * 0.43, x + width * 0.50, y + height * 0.43, { stroke: colors[0], width: 5, opacity: 0.85 })}
  </g>`;
}

function giftCard(x, y, width, gift, index, options = {}) {
  const [name, price, category, desc] = gift;
  const height = options.height ?? 270;
  const imageH = options.imageH ?? 104;
  const showButtons = options.buttons !== false;
  const showDescription = options.description !== false && height >= 300;
  return panel(
    x,
    y,
    width,
    height,
    `${photo(14, 14, width - 28, imageH, index)}
    ${chip(20, imageH + 28, category, Math.max(96, category.length * 9 + 34), 'muted')}
    ${text(20, imageH + 80, name, { size: 22, weight: 700, family: 'Fraunces, Georgia, serif' })}
    ${text(width - 20, imageH + 80, price, { size: 17, weight: 900, anchor: 'end' })}
    ${showDescription ? text(20, imageH + 112, desc, { size: 14, weight: 500, fill: C.inkMuted, lineHeight: 20 }) : ''}
    ${showButtons ? button(20, height - 58, 132, 'Подробнее', 'ghost', { height: 40, size: 13 }) : ''}
    ${showButtons ? button(width - 144, height - 58, 124, 'Купить', 'primary', { height: 40, size: 13 }) : ''}`,
    { rx: 22 },
  );
}

function miniGift(x, y, width, gift, index) {
  const [name, price, category] = gift;
  return panel(
    x,
    y,
    width,
    132,
    `${photo(12, 12, 84, 108, index)}
    ${chip(112, 20, category, Math.max(84, category.length * 8 + 32), 'muted')}
    ${text(112, 72, name, { size: 19, weight: 700, family: 'Fraunces, Georgia, serif' })}
    ${text(112, 104, price, { size: 16, weight: 900 })}`,
    { rx: 18 },
  );
}

function header(signedIn = false, active = 'Каталог') {
  const nav = [
    ['Каталог', 756],
    ['Как это работает', 842],
  ];
  return `<g>
    ${rect(0, 0, W, 80, { rx: 0, fill: C.bg, stroke: C.border, strokeWidth: 1, opacity: 0.96 })}
    ${text(X, 51, 'Gift Suggestion', { size: 25, weight: 800, family: 'Fraunces, Georgia, serif' })}
    ${nav.map(([label, x]) => text(x, 49, label, { size: 16, weight: active === label ? 900 : 700, fill: active === label ? C.ink : C.inkMuted })).join('')}
    ${signedIn
      ? `${chip(1040, 24, 'angelina@example.com', 180, 'muted')}${button(1240, 18, 80, 'Профиль', 'ghost', { height: 44, size: 13 })}`
      : `${text(1068, 49, 'Войти', { size: 16, weight: 700, fill: C.inkMuted })}${button(1140, 18, 172, 'Регистрация', 'secondary', { height: 44, size: 14 })}`}
  </g>`;
}

function footer(y = 918) {
  return `<g>
    ${line(X, y, X + CONTENT, y, { opacity: 0.74 })}
    ${text(X, y + 42, 'Gift Suggestion', { size: 18, weight: 900 })}
    ${text(X, y + 68, 'Подбор идей подарков, который не сводится к безликим фильтрам маркетплейса.', { size: 15, fill: C.inkMuted })}
    ${text(956, y + 52, 'Каталог идей     Войти     Создать аккаунт', { size: 15, fill: C.inkMuted })}
  </g>`;
}

function shell(titleText, body, options = {}) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}" role="img" aria-label="${esc(titleText)}">
  <defs>
    <radialGradient id="warmGlow" cx="0.08" cy="0.04" r="0.68">
      <stop offset="0" stop-color="#cbae63" stop-opacity="0.30"/>
      <stop offset="0.46" stop-color="#f6f1e8" stop-opacity="0.82"/>
      <stop offset="1" stop-color="#f6f1e8"/>
    </radialGradient>
    <filter id="shadow" x="-12%" y="-12%" width="124%" height="136%">
      <feDropShadow dx="0" dy="14" stdDeviation="14" flood-color="#4d3319" flood-opacity="0.10"/>
    </filter>
  </defs>
  ${rect(0, 0, W, H, { rx: 0, fill: 'url(#warmGlow)' })}
  ${options.noHeader ? '' : header(options.signedIn, options.active)}
  ${body}
</svg>
`;
}

function home() {
  return shell('01 Home final mockup', `
    ${panel(X, 118, 720, 352, `
      ${eyebrow(42, 58, 'GIFT SUGGESTION WEB SERVICE')}
      ${heading(42, 122, ['Подобрать подарок', 'без бесконечного', 'скролла маркетплейсов.'], 54)}
      ${text(42, 268, ['Каталог идей открыт без регистрации. Мастер подбора ведёт к входу,', 'потому что recommendation endpoint сейчас требует пользователя.'], { size: 17, fill: C.inkMuted, lineHeight: 23 })}
      ${button(42, 306, 190, 'Подобрать подарок', 'primary')}
      ${button(248, 306, 162, 'Каталог идей', 'secondary')}
    `)}
    ${panel(872, 118, 448, 352, `
      ${panel(24, 24, 400, 86, `${text(26, 36, 'Каталог', { size: 22, weight: 900 })}${text(26, 63, 'Поиск, категории и карточки подарков.', { size: 15, fill: C.inkMuted })}`, { rx: 16, shadow: false })}
      ${panel(24, 132, 400, 86, `${text(26, 36, 'Auth foundation', { size: 22, weight: 900 })}${text(26, 63, 'Login, register и reset request.', { size: 15, fill: C.inkMuted })}`, { rx: 16, shadow: false })}
      ${panel(24, 240, 400, 86, `${text(26, 36, 'Без выдуманного API', { size: 22, weight: 900 })}${text(26, 63, 'UI следует OpenAPI контракту.', { size: 15, fill: C.inkMuted })}`, { rx: 16, shadow: false })}
    `)}
    <g transform="translate(${X} 532)">
      ${eyebrow(0, 0, 'ПОЧЕМУ СЕРВИС ПОЛЕЗЕН')}
      ${heading(0, 48, 'Выбор подарка становится понятным сценарием.', 34)}
      ${panel(0, 82, 376, 116, `${text(24, 44, 'Объяснимые рекомендации', { size: 22, weight: 700, family: 'Fraunces, Georgia, serif' })}${text(24, 76, 'Причины выбора рядом с карточкой.', { size: 15, fill: C.inkMuted })}`, { rx: 20 })}
      ${panel(412, 82, 376, 116, `${text(24, 44, 'Каталог из магазинов', { size: 22, weight: 700, family: 'Fraunces, Georgia, serif' })}${text(24, 76, 'Цена, категория и переход к покупке.', { size: 15, fill: C.inkMuted })}`, { rx: 20 })}
      ${panel(824, 82, 376, 116, `${text(24, 44, 'Готово под wishlist', { size: 22, weight: 700, family: 'Fraunces, Georgia, serif' })}${text(24, 76, 'Паттерны сохранения уже заложены.', { size: 15, fill: C.inkMuted })}`, { rx: 20 })}
    </g>
    <g transform="translate(${X} 782)">
      ${eyebrow(0, 0, 'ПРЕДПРОСМОТР КАТАЛОГА')}
      ${heading(0, 44, 'Несколько свежих идей.', 32)}
      ${miniGift(0, 72, 276, gifts[0], 0)}
      ${miniGift(308, 72, 276, gifts[1], 1)}
      ${miniGift(616, 72, 276, gifts[2], 2)}
      ${miniGift(924, 72, 276, gifts[3], 3)}
    </g>`);
}

function catalog() {
  return shell('02 Catalog final mockup', `
    <g transform="translate(${X} 112)">
      ${eyebrow(0, 0, 'КАТАЛОГ ИДЕЙ')}
      ${heading(0, 52, ['Подберите стартовый список', 'подарков по категории и поиску.'], 42)}
      ${text(0, 154, 'Используются только catalog endpoints: q, category_id и sort.', { size: 16, fill: C.inkMuted })}
      ${panel(0, 190, CONTENT, 108, `
        ${input(24, 28, 520, 'Поиск', 'кофе, хобби, техника', { muted: true })}
        ${input(568, 28, 236, 'Сортировка', 'Сначала новые')}
        ${button(834, 42, 140, 'Искать', 'primary')}
        ${button(992, 42, 138, 'Сбросить', 'ghost')}
      `)}
      ${chip(0, 326, 'Все', 70)}
      ${chip(84, 326, 'Для дома', 112, 'muted')}
      ${chip(210, 326, 'Техника', 104, 'muted')}
      ${chip(328, 326, 'Впечатления', 136, 'green')}
      ${chip(478, 326, 'Хобби', 90, 'muted')}
      ${chip(582, 326, 'Культура', 108, 'muted')}
      ${text(0, 388, 'Найдено: 126', { size: 16, fill: C.inkMuted })}
      ${text(930, 388, 'Показаны первые 24 результата', { size: 16, fill: C.inkMuted })}
      ${giftCard(0, 414, 374, gifts[0], 0, { height: 226, imageH: 82 })}
      ${giftCard(413, 414, 374, gifts[1], 1, { height: 226, imageH: 82 })}
      ${giftCard(826, 414, 374, gifts[2], 2, { height: 226, imageH: 82 })}
      ${giftCard(0, 666, 374, gifts[3], 3, { height: 226, imageH: 82 })}
      ${giftCard(413, 666, 374, gifts[4], 4, { height: 226, imageH: 82 })}
      ${giftCard(826, 666, 374, gifts[5], 5, { height: 226, imageH: 82 })}
    </g>`);
}

function giftDetail() {
  return shell('03 Gift detail final mockup', `
    <g transform="translate(${X} 122)">
      ${text(0, 0, '← Вернуться в каталог', { size: 16, fill: C.inkMuted })}
      ${panel(0, 42, CONTENT, 682, `
        ${photo(28, 28, 548, 626, 1)}
        ${chip(624, 54, 'Уют', 82)}
        ${chip(722, 54, '12+', 64, 'muted')}
        ${heading(624, 142, ['Плед из мериноса', 'для спокойных вечеров'], 50)}
        ${text(624, 264, '5 490 ₽', { size: 30, weight: 900 })}
        ${text(624, 324, ['Тёплый подарок с понятной ценностью: уют, качество и отсутствие', 'риска ошибиться с размером. Подходит близкому человеку, которому', 'важен спокойный отдых дома.'], { size: 17, fill: C.inkMuted, lineHeight: 24 })}
        ${button(624, 446, 206, 'Перейти к покупке', 'primary')}
        ${button(850, 446, 216, 'Смотреть другие идеи', 'secondary')}
        ${banner(624, 548, 506, ['Backend отдаёт один store_link,', 'поэтому здесь один purchase CTA.'], 'warm')}
      `)}
      ${footer(790)}
    </g>`);
}

function authPage(kind) {
  const isLogin = kind === 'login';
  const isRegister = kind === 'register';
  const isReset = kind === 'reset';
  const fileTitle = isLogin ? '04 Login final mockup' : isRegister ? '05 Register final mockup' : '06 Password reset final mockup';
  const titleLines = isLogin
    ? ['Войти, чтобы сохранить', 'подборки и wishlist.']
    : isRegister
      ? ['Создать аккаунт', 'для будущих подборок.']
      : ['Восстановить доступ', 'через email.'];
  const formTitle = isLogin ? 'Вход' : isRegister ? 'Регистрация' : 'Сброс пароля';
  const mainButton = isLogin ? 'Войти' : isRegister ? 'Создать аккаунт' : 'Отправить письмо';
  const buttonY = isRegister ? 438 : isReset ? 322 : 408;

  return shell(fileTitle, `
    <g transform="translate(${X} 136)">
      ${eyebrow(0, 0, isReset ? 'PASSWORD RESET' : 'AUTH FOUNDATION')}
      ${heading(0, 68, titleLines, 50)}
      ${text(0, 200, ['Формы используют текущие auth endpoints и показывают', 'серверные ошибки без localStorage для access token.'], { size: 17, fill: C.inkMuted, lineHeight: 24 })}
      ${panel(0, 318, 480, 142, `${text(28, 48, 'Состояния формы', { size: 24, family: 'Fraunces, Georgia, serif', weight: 700 })}${text(28, 82, ['default, validation error, server error,', 'loading и success.'], { size: 15, fill: C.inkMuted, lineHeight: 21 })}`, { rx: 20 })}
      ${panel(636, 0, 564, 560, `
        ${text(40, 64, formTitle, { size: 34, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${isRegister ? input(40, 108, 484, 'Имя', 'Англеина Федяева') : ''}
        ${input(40, isRegister ? 188 : 130, 484, 'Email', 'angelina@example.com')}
        ${isReset ? '' : input(40, isRegister ? 268 : 212, 484, 'Пароль', '••••••••••')}
        ${isRegister ? input(40, 348, 484, 'Повтор пароля', '••••••••••') : ''}
        ${isLogin ? banner(40, 306, 484, 'Неверный email или пароль', 'danger') : ''}
        ${isReset ? banner(40, 220, 484, ['Если email существует,', 'письмо будет отправлено.'], 'success') : ''}
        ${button(40, buttonY, 484, mainButton, 'primary')}
        ${text(40, buttonY + 88, isLogin ? 'Нет аккаунта? Регистрация      Забыли пароль?' : isRegister ? 'Уже есть аккаунт? Войти' : 'Вернуться ко входу', { size: 15, fill: C.inkMuted })}
      `)}
      ${footer(734)}
    </g>`);
}

function wizard() {
  return shell('07 Recommendation wizard final mockup', `
    <g transform="translate(${X} 118)">
      ${eyebrow(0, 0, 'МАСТЕР ПОДБОРА')}
      ${heading(0, 58, ['Ответьте на несколько вопросов,', 'чтобы получить объяснимые идеи.'], 46)}
      ${panel(0, 178, CONTENT, 606, `
        ${chip(42, 38, '1 Повод', 98)}
        ${chip(156, 38, '2 Получатель', 140, 'green')}
        ${chip(312, 38, '3 Бюджет', 112, 'muted')}
        ${chip(440, 38, '4 Интересы', 126, 'muted')}
        ${chip(582, 38, '5 Проверка', 128, 'muted')}
        ${line(42, 94, 1158, 94)}
        ${text(48, 154, 'Кому выбираем подарок?', { size: 34, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(48, 196, ['Поля соответствуют текущему RecommendRequest.', 'Пол отсутствует в контракте и не выводится.'], { size: 16, fill: C.inkMuted, lineHeight: 22 })}
        ${panel(48, 254, 312, 116, `${text(28, 56, 'Партнёру', { size: 22, weight: 900 })}${text(28, 84, 'Близкий сценарий', { size: 15, fill: C.inkMuted })}`, { rx: 18, fill: C.secondarySoft, stroke: C.secondary, shadow: false })}
        ${panel(392, 254, 312, 116, `${text(28, 56, 'Коллеге', { size: 22, weight: 900 })}${text(28, 84, 'Нейтральный тон', { size: 15, fill: C.inkMuted })}`, { rx: 18, shadow: false })}
        ${panel(736, 254, 312, 116, `${text(28, 56, 'Друзьям', { size: 22, weight: 900 })}${text(28, 84, 'Можно ярче', { size: 15, fill: C.inkMuted })}`, { rx: 18, shadow: false })}
        ${input(48, 424, 312, 'Возраст', '28')}
        ${input(392, 424, 312, 'Бюджет', 'до 7 000 ₽')}
        ${input(736, 424, 312, 'Повод', 'день рождения')}
        ${button(48, 544, 128, 'Назад', 'ghost')}
        ${button(920, 544, 128, 'Дальше', 'primary')}
      `)}
      ${banner(0, 824, CONTENT, 'Recommendation submit доступен после входа пользователя.', 'info')}
    </g>`, { signedIn: true, active: 'Подбор' });
}

function recommendationResults() {
  return shell('08 Recommendation results final mockup', `
    <g transform="translate(${X} 118)">
      ${eyebrow(0, 0, 'РЕЗУЛЬТАТЫ ПОДБОРА')}
      ${heading(0, 58, 'Лучшие идеи с объяснениями.', 46)}
      ${button(966, 4, 234, 'Сохранить подборку', 'secondary')}
      ${giftCard(0, 132, 374, gifts[1], 1, { height: 254, imageH: 100 })}
      ${giftCard(413, 132, 374, gifts[4], 4, { height: 254, imageH: 100 })}
      ${giftCard(826, 132, 374, gifts[0], 0, { height: 254, imageH: 100 })}
      ${panel(0, 428, 770, 238, `
        ${text(34, 58, 'Почему эти варианты', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(34, 112, ['Подбор отдаёт score, source и explanation. UI показывает причины', 'рядом с карточками и не скрывает fallback или пустой результат.'], { size: 17, fill: C.inkMuted, lineHeight: 24 })}
        ${chip(34, 174, 'budget match', 124, 'green')}
        ${chip(174, 174, 'interest overlap', 150)}
        ${chip(340, 174, 'category fit', 126, 'muted')}
      `)}
      ${panel(810, 428, 390, 238, `
        ${text(28, 58, 'Альтернативы', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(28, 112, ['Если первый вариант не подходит,', 'показываем соседние категории.'], { size: 16, fill: C.inkMuted, lineHeight: 22 })}
        ${button(28, 168, 150, 'Открыть', 'ghost', { height: 42 })}
      `)}
      ${banner(0, 706, CONTENT, ['Wishlist state не приходит в recommendation payload.', 'Сохранение требует отдельного wishlist action.'], 'warm')}
    </g>`, { signedIn: true, active: 'Подбор' });
}

function profile() {
  return shell('09 Profile final mockup', `
    <g transform="translate(${X} 122)">
      ${eyebrow(0, 0, 'ПРОФИЛЬ')}
      ${heading(0, 58, 'Личный кабинет пользователя.', 46)}
      ${panel(0, 142, 575, 424, `
        ${circle(50, 72, 38, { fill: C.secondarySoft, stroke: C.secondary, strokeWidth: 1 })}
        ${text(108, 66, 'Англеина Федяева', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(108, 98, 'angelina@example.com', { size: 16, fill: C.inkMuted })}
        ${input(48, 156, 480, 'Отображаемое имя', 'Англеина Федяева')}
        ${input(48, 248, 480, 'Email', 'angelina@example.com')}
        ${button(48, 346, 176, 'Сохранить', 'primary')}
      `)}
      ${panel(625, 142, 575, 424, `
        ${text(40, 62, 'Быстрые переходы', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${panel(40, 102, 495, 82, `${text(26, 50, 'Wishlist', { size: 22, weight: 900 })}${text(396, 50, '12 идей', { size: 16, fill: C.inkMuted })}`, { rx: 16, shadow: false, fill: C.secondarySoft, stroke: C.secondary })}
        ${panel(40, 204, 495, 82, `${text(26, 50, 'Интеграции', { size: 22, weight: 900 })}${text(374, 50, 'VK planned', { size: 16, fill: C.inkMuted })}`, { rx: 16, shadow: false })}
        ${text(40, 340, ['getCurrentUser и updateCurrentUser покрывают', 'этот экран без дополнительных API.'], { size: 16, fill: C.inkMuted, lineHeight: 22 })}
      `)}
    </g>`, { signedIn: true, active: 'Профиль' });
}

function wishlist() {
  return shell('10 Wishlist final mockup', `
    <g transform="translate(${X} 118)">
      ${eyebrow(0, 0, 'WISHLIST')}
      ${heading(0, 58, 'Сохранённые идеи подарков.', 46)}
      ${button(1002, 8, 198, 'Новый список', 'primary')}
      ${panel(0, 140, 318, 574, `
        ${text(28, 58, 'Списки', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${panel(28, 96, 262, 74, `${text(24, 46, 'День рождения', { size: 19, weight: 900 })}${text(224, 46, '5', { size: 18, weight: 900, fill: C.primary })}`, { rx: 14, fill: C.secondarySoft, stroke: C.secondary, shadow: false })}
        ${panel(28, 188, 262, 74, `${text(24, 46, 'Новый год', { size: 19, weight: 900 })}${text(224, 46, '7', { size: 18, fill: C.inkMuted })}`, { rx: 14, shadow: false })}
        ${panel(28, 280, 262, 74, `${text(24, 46, 'Коллегам', { size: 19, weight: 900 })}${text(224, 46, '2', { size: 18, fill: C.inkMuted })}`, { rx: 14, shadow: false })}
      `)}
      ${giftCard(356, 140, 392, gifts[0], 0, { height: 252, imageH: 100 })}
      ${giftCard(808, 140, 392, gifts[1], 1, { height: 252, imageH: 100 })}
      ${giftCard(356, 430, 392, gifts[5], 5, { height: 252, imageH: 100 })}
      ${panel(808, 430, 392, 252, `
        ${text(32, 58, 'Пустой список', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(32, 108, ['Когда список пустой, показываем CTA', 'в каталог без фиктивных рекомендаций.'], { size: 16, fill: C.inkMuted, lineHeight: 22 })}
        ${button(32, 174, 178, 'Открыть каталог', 'secondary')}
      `)}
    </g>`, { signedIn: true, active: 'Wishlist' });
}

function integrations() {
  return shell('11 Integrations VK final mockup', `
    <g transform="translate(${X} 122)">
      ${eyebrow(0, 0, 'ИНТЕГРАЦИИ')}
      ${heading(0, 58, 'VK интересы как planned state.', 46)}
      ${panel(0, 150, CONTENT, 430, `
        ${circle(60, 76, 40, { fill: '#dbe7ff', stroke: '#8aa0d8', strokeWidth: 1 })}
        ${text(124, 82, 'VK Connect', { size: 34, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${chip(1010, 58, 'planned', 104, 'muted')}
        ${text(124, 132, ['Экран показывает будущий entry point. Connect и sync actions заблокированы,', 'потому что в OpenAPI и backend-коде нет VK HTTP endpoints.'], { size: 17, fill: C.inkMuted, lineHeight: 24 })}
        ${button(124, 224, 180, 'Подключить VK', 'disabled', { opacity: 0.72 })}
        ${button(324, 224, 220, 'Синхронизировать', 'disabled', { opacity: 0.72 })}
        ${banner(124, 322, 836, ['Gap: нет endpoint для OAuth connect, status', 'и sync interests.'], 'danger')}
      `)}
    </g>`, { signedIn: true, active: 'Интеграции' });
}

function adminImport() {
  return shell('12 Admin import final mockup', `
    <g transform="translate(${X} 118)">
      ${eyebrow(0, 0, 'ADMIN')}
      ${heading(0, 58, 'Импорт каталога подарков.', 46)}
      ${panel(0, 142, 720, 520, `
        ${rect(42, 48, 636, 236, { rx: 22, fill: C.muted, stroke: C.border, strokeWidth: 2 })}
        ${text(360, 146, 'Перетащите CSV или XLSX', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700, anchor: 'middle' })}
        ${text(360, 188, 'Файл не отправляется до подтверждения', { size: 16, fill: C.inkMuted, anchor: 'middle' })}
        ${button(255, 222, 210, 'Выбрать файл', 'secondary')}
        ${input(42, 334, 320, 'Source label', 'marketplace_april.csv')}
        ${button(392, 348, 214, 'Запустить импорт', 'primary')}
      `)}
      ${panel(780, 142, 420, 520, `
        ${text(32, 62, 'Последний job', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${chip(32, 94, 'processing', 120, 'green')}
        ${text(32, 168, 'Записей обработано', { size: 16, fill: C.inkMuted })}
        ${text(32, 216, '1 248 / 1 540', { size: 42, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${rect(32, 260, 320, 12, { rx: 999, fill: C.muted })}
        ${rect(32, 260, 238, 12, { rx: 999, fill: C.primary })}
        ${button(32, 322, 198, 'Открыть статус', 'ghost')}
      `)}
    </g>`, { signedIn: true, active: 'Admin' });
}

function importStatus() {
  return shell('13 Import status errors final mockup', `
    <g transform="translate(${X} 118)">
      ${eyebrow(0, 0, 'IMPORT JOB')}
      ${heading(0, 58, 'Статус импорта и ошибки.', 46)}
      ${panel(0, 134, CONTENT, 172, `
        ${chip(34, 38, 'completed_with_errors', 200)}
        ${text(34, 110, '1 502 обработано', { size: 31, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(386, 110, '38 ошибок', { size: 31, family: 'Fraunces, Georgia, serif', weight: 700, fill: C.danger })}
        ${text(720, 110, 'Файл: marketplace_april.csv', { size: 18, fill: C.inkMuted })}
      `)}
      ${panel(0, 350, CONTENT, 456, `
        ${text(34, 58, 'Ошибки импорта', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${line(34, 104, 1166, 104)}
        ${text(34, 148, 'Строка', { size: 15, weight: 900, fill: C.inkMuted })}
        ${text(170, 148, 'Поле', { size: 15, weight: 900, fill: C.inkMuted })}
        ${text(330, 148, 'Сообщение', { size: 15, weight: 900, fill: C.inkMuted })}
        ${line(34, 172, 1166, 172)}
        ${text(34, 220, '24', { size: 17 })}
        ${text(170, 220, 'price', { size: 17 })}
        ${text(330, 220, 'Цена должна быть положительным числом', { size: 17, fill: C.inkMuted })}
        ${line(34, 250, 1166, 250)}
        ${text(34, 298, '41', { size: 17 })}
        ${text(170, 298, 'store_link', { size: 17 })}
        ${text(330, 298, 'Некорректный URL магазина', { size: 17, fill: C.inkMuted })}
        ${line(34, 328, 1166, 328)}
        ${text(34, 376, '77', { size: 17 })}
        ${text(170, 376, 'category', { size: 17 })}
        ${text(330, 376, 'Категория не найдена', { size: 17, fill: C.inkMuted })}
      `)}
    </g>`, { signedIn: true, active: 'Admin' });
}

function states() {
  return shell('14 System states final mockup', `
    <g transform="translate(${X} 118)">
      ${eyebrow(0, 0, 'SYSTEM STATES')}
      ${heading(0, 58, 'Loading, empty, error и success.', 46)}
      ${panel(0, 136, 374, 252, `
        ${text(28, 58, 'Loading', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${rect(28, 96, 308, 18, { rx: 999, fill: C.muted })}
        ${rect(28, 132, 246, 18, { rx: 999, fill: C.surfaceAlt })}
        ${rect(28, 168, 308, 58, { rx: 16, fill: C.muted })}
      `)}
      ${panel(413, 136, 374, 252, `
        ${text(28, 58, 'Empty', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(28, 108, ['Каталог ничего не вернул.', 'Предлагаем сбросить фильтры.'], { size: 16, fill: C.inkMuted, lineHeight: 22 })}
        ${button(28, 174, 156, 'Сбросить', 'secondary')}
      `)}
      ${panel(826, 136, 374, 252, `
        ${text(28, 58, 'Error', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${banner(28, 94, 304, 'Не удалось загрузить данные', 'danger')}
        ${button(28, 174, 164, 'Повторить', 'primary')}
      `)}
      ${panel(0, 428, 374, 252, `
        ${text(28, 58, 'Success', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${banner(28, 96, 304, 'Письмо отправлено', 'success')}
      `)}
      ${panel(413, 428, 374, 252, `
        ${text(28, 58, 'Unauthorized', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(28, 108, ['Для wishlist и подборок', 'нужно войти в аккаунт.'], { size: 16, fill: C.inkMuted, lineHeight: 22 })}
        ${button(28, 174, 136, 'Войти', 'primary')}
      `)}
      ${panel(826, 428, 374, 252, `
        ${text(28, 58, 'Not found', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(28, 108, ['Страница или подарок', 'не найдены.'], { size: 16, fill: C.inkMuted, lineHeight: 22 })}
        ${button(28, 174, 158, 'В каталог', 'secondary')}
      `)}
      ${banner(0, 730, CONTENT, ['Все состояния используют текущие banner, empty-state, skeleton', 'и button паттерны из frontend Slice 1.'], 'info')}
    </g>`);
}

const mockups = [
  ['01-home.svg', 'Home', 'Public', home],
  ['02-catalog.svg', 'Catalog', 'Public', catalog],
  ['03-gift-detail.svg', 'Gift Detail', 'Public', giftDetail],
  ['04-login.svg', 'Login', 'Auth', () => authPage('login')],
  ['05-register.svg', 'Register', 'Auth', () => authPage('register')],
  ['06-password-reset.svg', 'Password Reset Request', 'Auth', () => authPage('reset')],
  ['07-recommendation-wizard.svg', 'Recommendation Wizard', 'Recommendation', wizard],
  ['08-recommendation-results.svg', 'Recommendation Results', 'Recommendation', recommendationResults],
  ['09-profile.svg', 'Profile', 'User Area', profile],
  ['10-wishlist.svg', 'Wishlist', 'User Area', wishlist],
  ['11-integrations-vk.svg', 'Integrations / VK', 'User Area', integrations],
  ['12-admin-import.svg', 'Admin Import', 'Admin', adminImport],
  ['13-import-status-errors.svg', 'Import Status / Errors', 'Admin', importStatus],
  ['14-system-states.svg', 'System States', 'States', states],
];

for (const [file, , , render] of mockups) {
  writeFileSync(join(mockupsDir, file), render(), 'utf8');
}

const groups = mockups.reduce((acc, [file, label, group]) => {
  acc[group] ??= [];
  acc[group].push({ file, label });
  return acc;
}, {});

const previewSections = Object.entries(groups)
  .map(([group, items]) => `<section class="preview-section" id="${esc(group.toLowerCase().replaceAll(' ', '-'))}">
    <div class="section-heading">
      <span>${esc(group)}</span>
      <small>${items.length} screen${items.length > 1 ? 's' : ''}</small>
    </div>
    <div class="mockup-grid">
      ${items.map((item) => `<article class="mockup-card">
        <div class="mockup-card__header">
          <span>${esc(item.file.replace('.svg', ''))} · ${esc(item.label)}</span>
          <a href="mockups/${esc(item.file)}" target="_blank" rel="noreferrer">Open SVG</a>
        </div>
        <img src="mockups/${esc(item.file)}" alt="${esc(item.label)} mockup" loading="lazy" />
      </article>`).join('\n')}
    </div>
  </section>`)
  .join('\n');

writeFileSync(
  join(designDir, 'index.html'),
  `<!doctype html>
<html lang="ru">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Gift Suggestion - Final Site Mockups</title>
    <style>
      :root {
        --bg: ${C.bg};
        --surface: ${C.surface};
        --surface-muted: ${C.muted};
        --ink: ${C.ink};
        --ink-muted: ${C.inkMuted};
        --primary: ${C.primary};
        --secondary: ${C.secondary};
        --border: ${C.border};
        --shadow: 0 18px 50px rgba(77, 51, 25, 0.10);
      }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        background:
          radial-gradient(circle at top left, rgba(203, 174, 99, 0.30), transparent 26%),
          radial-gradient(circle at top right, rgba(36, 92, 74, 0.13), transparent 22%),
          var(--bg);
        color: var(--ink);
        font-family: Manrope, "Segoe UI", sans-serif;
      }
      a { color: inherit; }
      header {
        max-width: 1220px;
        margin: 0 auto;
        padding: 46px 24px 30px;
      }
      h1, h2 {
        margin: 0;
        font-family: Fraunces, Georgia, serif;
      }
      h1 { max-width: 920px; font-size: clamp(42px, 6vw, 78px); line-height: 0.96; }
      p { max-width: 800px; color: var(--ink-muted); font-size: 18px; line-height: 1.55; }
      .toolbar {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        margin-top: 24px;
      }
      .pill {
        border: 1px solid var(--border);
        border-radius: 999px;
        background: rgba(255,253,248,0.88);
        padding: 10px 14px;
        color: var(--ink-muted);
        font-weight: 800;
      }
      main {
        display: grid;
        gap: 58px;
        max-width: 1220px;
        margin: 0 auto;
        padding: 0 24px 78px;
      }
      .preview-section {
        display: grid;
        gap: 18px;
      }
      .section-heading {
        display: flex;
        align-items: end;
        justify-content: space-between;
        gap: 16px;
        border-bottom: 1px solid rgba(216, 200, 176, 0.75);
        padding-bottom: 12px;
      }
      .section-heading span {
        font-family: Fraunces, Georgia, serif;
        font-size: 34px;
        font-weight: 700;
      }
      .section-heading small {
        color: var(--ink-muted);
        font-weight: 800;
      }
      .mockup-grid {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 24px;
      }
      .mockup-card {
        overflow: hidden;
        border: 1px solid rgba(216, 200, 176, 0.82);
        border-radius: 24px;
        background: rgba(255,253,248,0.96);
        box-shadow: var(--shadow);
      }
      .mockup-card__header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        border-bottom: 1px solid rgba(216, 200, 176, 0.72);
        padding: 14px 18px;
        font-weight: 900;
      }
      .mockup-card__header a {
        color: var(--primary);
        text-decoration: none;
        white-space: nowrap;
      }
      img {
        display: block;
        width: 100%;
        background: var(--surface-muted);
      }
      @media (max-width: 900px) {
        .mockup-grid { grid-template-columns: 1fr; }
        .section-heading { align-items: start; flex-direction: column; }
      }
    </style>
  </head>
  <body>
    <header>
      <h1>Gift Suggestion final site mockups</h1>
      <p>Статичные Figma-ready SVG макеты, собранные по текущим frontend токенам и UI-паттернам: тёплый фон, акцент #C65A1E, secondary #245C4A, rounded cards, e-commerce сетка и auth form patterns.</p>
      <div class="toolbar">
        <span class="pill">14 desktop screens</span>
        <span class="pill">1440x1024 SVG</span>
        <span class="pill">Figma import ready</span>
        <span class="pill">No backend runtime</span>
      </div>
    </header>
    <main>
      ${previewSections}
    </main>
  </body>
</html>
`,
  'utf8',
);

writeFileSync(
  join(designDir, 'README.md'),
  `# Frontend Design Mockups

Static, Figma-ready final-site mockups for the Gift Suggestion frontend. These artifacts are design review materials only: they do not connect to backend, do not add runtime dependencies, and do not change production frontend code.

## Design Concept

The mockups follow the current Slice 1 frontend visual language:

- Warm light background: \`#f6f1e8\`
- Surface cards: \`#fffdf8\`
- Primary action: \`#c65a1e\`
- Secondary action: \`#245c4a\`
- Accent: \`#cbae63\`
- Borders: \`#d8c8b0\`
- Editorial commerce tone with serif headings, rounded cards, catalog grids, gift cards, auth form blocks, banners and skeleton states.

Typography mirrors the frontend CSS intent:

- Headings: \`Fraunces, Georgia, serif\`
- Body/UI: \`Manrope, Segoe UI, sans-serif\`
- Numeric emphasis uses compact, high-contrast UI text.

The quality pass keeps each screen inside a fixed \`1440x1024\` SVG frame, with no external images, fonts, local file links or backend calls.

## Covered Screens

- \`01-home.svg\`
- \`02-catalog.svg\`
- \`03-gift-detail.svg\`
- \`04-login.svg\`
- \`05-register.svg\`
- \`06-password-reset.svg\`
- \`07-recommendation-wizard.svg\`
- \`08-recommendation-results.svg\`
- \`09-profile.svg\`
- \`10-wishlist.svg\`
- \`11-integrations-vk.svg\`
- \`12-admin-import.svg\`
- \`13-import-status-errors.svg\`
- \`14-system-states.svg\`

Open the preview page:

\`\`\`text
services/frontend/design/index.html
\`\`\`

## Figma Import

1. Create or open the Figma page named \`Frontend Final Mockups\`.
2. Create the sections from \`figma-layout-spec.md\`.
3. Drag SVG files from \`services/frontend/design/mockups/\` onto the Figma canvas.
4. Keep each imported SVG at \`1440x1024\`.
5. Place frames in the specified order and use each SVG file name as the frame caption.

SVG files are pure static artifacts, so they can be opened in a browser and imported into Figma without running the app.

## Backend Gaps Reflected In UI

- Recommendation submit is auth-only, so the wizard includes an auth limitation note.
- VK connect/sync is shown as a disabled planned state because there are no VK HTTP endpoints in the current OpenAPI/backend code.
- Gift detail uses one purchase CTA because backend exposes one \`store_link\`, not a store list.
- Recommendation request does not contain \`gender\`, so the wizard does not render that field.
- Recommendation refine/filtering is not drawn as a working server feature.
- Wishlist saved flags are not assumed inside catalog/recommendation payloads; save actions are presented as explicit wishlist operations.

## Regeneration

The SVG and HTML artifacts can be regenerated with:

\`\`\`bash
node services/frontend/design/tools/render-static-mockups.mjs
\`\`\`
`,
  'utf8',
);

writeFileSync(
  join(designDir, 'figma-layout-spec.md'),
  `# Figma Layout Spec

Use this spec to place the static mockups into a Figma review page.

## Page

Page name: \`Frontend Final Mockups\`

Canvas order:

1. \`00 Cover / Notes\`
2. \`01 Public\`
3. \`02 Auth\`
4. \`03 Recommendation\`
5. \`04 User Area\`
6. \`05 Admin\`
7. \`06 States\`
8. \`07 Mobile Companion Frames\`

## Frame Sizes

- Desktop frames: \`1440x1024\`
- Mobile companion frames: \`390x844\`
- Section spacing: \`160px\` horizontal, \`140px\` vertical
- Caption spacing below imported SVG: \`24px\`

Every SVG has \`width="1440"\`, \`height="1024"\` and \`viewBox="0 0 1440 1024"\`.

## Desktop Frame Order

Place desktop SVGs left to right, then wrap to the next row.

### 01 Public

1. \`01-home.svg\` - Home
2. \`02-catalog.svg\` - Catalog
3. \`03-gift-detail.svg\` - Gift Detail

### 02 Auth

1. \`04-login.svg\` - Login
2. \`05-register.svg\` - Register
3. \`06-password-reset.svg\` - Password Reset Request

### 03 Recommendation

1. \`07-recommendation-wizard.svg\` - Recommendation Wizard
2. \`08-recommendation-results.svg\` - Recommendation Results

### 04 User Area

1. \`09-profile.svg\` - Profile
2. \`10-wishlist.svg\` - Wishlist
3. \`11-integrations-vk.svg\` - Integrations / VK

### 05 Admin

1. \`12-admin-import.svg\` - Admin Import
2. \`13-import-status-errors.svg\` - Import Status / Errors

### 06 States

1. \`14-system-states.svg\` - Loading / Empty / Error / Success / Unauthorized / Not Found

## Suggested Coordinates

Start at \`x=0, y=0\` inside each section.

- Frame 1: \`x=0, y=0\`
- Frame 2: \`x=1600, y=0\`
- Frame 3: \`x=3200, y=0\`
- New row: add \`y=1220\`

For captions, add text below each imported SVG:

- Font: \`Manrope SemiBold\`
- Size: \`24\`
- Color: \`#1f1a14\`

## Mobile Companion Frames

Create mobile frames at \`390x844\` for review of responsive behavior. These are companion frames derived from the desktop mockups, not separate API screens.

Recommended mobile order:

1. \`M01 Home\`
2. \`M02 Catalog\`
3. \`M03 Gift Detail\`
4. \`M04 Login\`
5. \`M05 Recommendation Wizard\`
6. \`M06 Wishlist Empty\`
7. \`M07 Admin Import Status\`

Mobile layout rules:

- Header stacks brand, nav and auth action vertically.
- Page horizontal padding: \`16px\`.
- Catalog and wishlist grids collapse to one column.
- Gift detail image moves above content.
- Auth layout becomes intro text above form.
- Admin status table becomes stacked rows.
- Primary CTA remains visible before secondary actions.

## Import Notes

Drag files from \`services/frontend/design/mockups/\` into the Figma canvas. Keep imported SVG dimensions at \`1440x1024\`. Do not scale mockups non-proportionally.

The mockups intentionally preserve backend limitations:

- VK actions are disabled/planned.
- Recommendation submit requires auth.
- Gift detail has one purchase CTA from \`store_link\`.
- No invented wishlist flag is shown in catalog cards.
`,
  'utf8',
);

console.log(`Rendered ${mockups.length} polished mockups into ${mockupsDir}`);
