import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const designDir = join(here, '..');
const mockupsDir = join(designDir, 'mockups');

mkdirSync(mockupsDir, { recursive: true });

const W = 1440;
const H = 1024;
const C = {
  bg: '#f6f1e8',
  surface: '#fffdf8',
  muted: '#efe6d6',
  ink: '#1f1a14',
  inkMuted: '#625746',
  primary: '#c65a1e',
  primaryHover: '#a94816',
  secondary: '#245c4a',
  secondarySoft: '#dbe9e2',
  accent: '#cbae63',
  danger: '#b13a2f',
  dangerSoft: '#f6ddd8',
  success: '#2f7a4a',
  successSoft: '#dcebdd',
  border: '#d8c8b0',
};

const gifts = [
  ['Кофейный сет', '2 790 ₽', 'Для дома', 'Свежеобжаренный кофе, керамика и открытка.'],
  ['Плед из мериноса', '5 490 ₽', 'Уют', 'Тёплый подарок для вечеров и спокойного отдыха.'],
  ['Книга-альбом', '3 200 ₽', 'Культура', 'Большое издание с иллюстрациями и историей дизайна.'],
  ['Умная колонка', '7 990 ₽', 'Техника', 'Музыка, напоминания и голосовые сценарии дома.'],
  ['Набор для керамики', '4 100 ₽', 'Хобби', 'Глина, инструменты и понятный стартовый гид.'],
  ['Сертификат SPA', '6 500 ₽', 'Впечатления', 'Подарок-восстановление без угадывания размера.'],
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

function title(x, y, lines, size = 54) {
  return text(x, y, lines, {
    size,
    lineHeight: Math.round(size * 1.04),
    weight: 700,
    family: 'Fraunces, Georgia, serif',
  });
}

function eyebrow(x, y, value) {
  return text(x, y, value, {
    size: 14,
    weight: 800,
    fill: C.primary,
  });
}

function button(x, y, width, label, variant = 'primary', options = {}) {
  const fills = {
    primary: C.primary,
    secondary: C.secondary,
    ghost: C.surface,
    disabled: C.muted,
    danger: C.danger,
  };
  const stroke = variant === 'ghost' || variant === 'disabled' ? C.border : 'transparent';
  const color = variant === 'ghost' || variant === 'disabled' ? C.ink : '#ffffff';
  return `<g opacity="${options.opacity ?? 1}">
    ${rect(x, y, width, options.height ?? 52, {
      rx: 999,
      fill: fills[variant] ?? C.primary,
      stroke,
      strokeWidth: 1,
    })}
    <text ${attrs({
      x: x + width / 2,
      y: y + (options.height ?? 52) / 2 + 7,
      fill: color,
      'font-family': 'Manrope, Segoe UI, sans-serif',
      'font-size': options.size ?? 17,
      'font-weight': 800,
      'text-anchor': 'middle',
    })}>${esc(label)}</text>
  </g>`;
}

function chip(x, y, label, width, variant = 'primary') {
  const fill = variant === 'green' ? C.secondarySoft : variant === 'muted' ? C.muted : '#f3e3d8';
  const color = variant === 'green' ? C.secondary : C.ink;
  return `<g>
    ${rect(x, y, width, 34, { rx: 999, fill, stroke: variant === 'muted' ? C.border : undefined })}
    ${text(x + 16, y + 22, label, { size: 14, weight: 800, fill: color })}
  </g>`;
}

function input(x, y, width, label, value, options = {}) {
  return `<g>
    ${text(x, y, label, { size: 15, weight: 800 })}
    ${rect(x, y + 14, width, options.height ?? 54, { rx: 12, fill: C.surface, stroke: C.border })}
    ${text(x + 18, y + 48, value, { size: 16, weight: 500, fill: options.muted ? C.inkMuted : C.ink })}
  </g>`;
}

function card(x, y, width, height, body = '', options = {}) {
  return `<g>
    ${rect(x, y, width, height, {
      rx: options.rx ?? 28,
      fill: options.fill ?? C.surface,
      stroke: options.stroke ?? C.border,
      strokeWidth: 1,
      filter: options.shadow ? 'url(#shadow)' : undefined,
    })}
    ${body}
  </g>`;
}

function photo(x, y, width, height, variant = 0) {
  const colors = [
    ['#d9b66b', '#245c4a'],
    ['#c65a1e', '#f0c987'],
    ['#8f9e85', '#efe6d6'],
    ['#245c4a', '#cbae63'],
    ['#b76e45', '#dbe9e2'],
  ][variant % 5];
  return `<g>
    ${rect(x, y, width, height, { rx: 22, fill: colors[1] })}
    <path d="M${x} ${y + height * 0.82} C${x + width * 0.25} ${y + height * 0.58}, ${x + width * 0.52} ${y + height * 0.92}, ${x + width} ${y + height * 0.60} L${x + width} ${y + height} L${x} ${y + height} Z" fill="${colors[0]}" opacity="0.78"/>
    ${circle(x + width * 0.78, y + height * 0.24, Math.min(width, height) * 0.13, { fill: '#fffdf8', opacity: 0.74 })}
    ${rect(x + width * 0.19, y + height * 0.29, width * 0.31, height * 0.34, { rx: 14, fill: '#fffdf8', opacity: 0.72 })}
    ${line(x + width * 0.34, y + height * 0.29, x + width * 0.34, y + height * 0.63, { stroke: colors[0], width: 5, opacity: 0.85 })}
    ${line(x + width * 0.19, y + height * 0.43, x + width * 0.50, y + height * 0.43, { stroke: colors[0], width: 5, opacity: 0.85 })}
  </g>`;
}

function giftCard(x, y, width, gift, index, compact = false) {
  const [name, price, category, desc] = gift;
  const imageH = compact ? 110 : 138;
  return card(
    x,
    y,
    width,
    compact ? 255 : 310,
    `${photo(x + 14, y + 14, width - 28, imageH, index)}
    ${chip(x + 22, y + imageH + 28, category, 104, 'muted')}
    ${text(x + 22, y + imageH + 82, name, { size: compact ? 21 : 24, weight: 700, family: 'Fraunces, Georgia, serif' })}
    ${text(x + width - 22, y + imageH + 82, price, { size: 18, weight: 900, anchor: 'end' })}
    ${text(x + 22, y + imageH + 114, [desc], { size: 15, weight: 500, fill: C.inkMuted })}
    ${compact ? '' : button(x + 22, y + 244, 140, 'Подробнее', 'ghost', { height: 42, size: 14 })}
    ${compact ? '' : button(x + width - 158, y + 244, 136, 'Купить', 'primary', { height: 42, size: 14 })}`,
    { shadow: true, rx: 22 },
  );
}

function giftMini(x, y, width, gift, index) {
  const [name, price, category] = gift;
  return card(
    x,
    y,
    width,
    146,
    `${photo(x + 14, y + 14, 88, 118, index)}
    ${chip(x + 118, y + 24, category, 104, 'muted')}
    ${text(x + 118, y + 80, name, { size: 20, weight: 700, family: 'Fraunces, Georgia, serif' })}
    ${text(x + 118, y + 114, price, { size: 17, weight: 900 })}`,
    { shadow: true, rx: 20 },
  );
}

function header(signedIn = false, active = 'Каталог') {
  return `<g>
    ${rect(0, 0, W, 76, { rx: 0, fill: C.bg, stroke: C.border, strokeWidth: 1, opacity: 0.94 })}
    ${text(130, 48, 'Gift Suggestion', { size: 25, weight: 800, family: 'Fraunces, Georgia, serif' })}
    ${text(788, 47, 'Каталог', { size: 16, weight: active === 'Каталог' ? 900 : 700, fill: active === 'Каталог' ? C.ink : C.inkMuted })}
    ${text(872, 47, 'Как это работает', { size: 16, weight: 700, fill: C.inkMuted })}
    ${signedIn
      ? chip(1055, 22, 'maria@example.com', 170, 'muted') + button(1240, 16, 70, 'Профиль', 'ghost', { height: 44, size: 14 })
      : text(1072, 47, 'Войти', { size: 16, weight: 700, fill: C.inkMuted }) + button(1142, 16, 160, 'Регистрация', 'secondary', { height: 44, size: 14 })}
  </g>`;
}

function footer(y = 938) {
  return `<g>
    ${line(130, y, 1310, y, { opacity: 0.72 })}
    ${text(130, y + 44, 'Gift Suggestion', { size: 18, weight: 900 })}
    ${text(130, y + 70, 'Подбор идей подарков, который не сводится к безликим фильтрам маркетплейса.', { size: 15, fill: C.inkMuted })}
    ${text(1000, y + 52, 'Каталог идей     Войти     Создать аккаунт', { size: 15, fill: C.inkMuted })}
  </g>`;
}

function shell(titleText, body, options = {}) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}" role="img" aria-label="${esc(titleText)}">
  <defs>
    <radialGradient id="warmGlow" cx="0.08" cy="0.04" r="0.68">
      <stop offset="0" stop-color="#cbae63" stop-opacity="0.34"/>
      <stop offset="0.46" stop-color="#f6f1e8" stop-opacity="0.8"/>
      <stop offset="1" stop-color="#f6f1e8"/>
    </radialGradient>
    <filter id="shadow" x="-20%" y="-20%" width="140%" height="160%">
      <feDropShadow dx="0" dy="18" stdDeviation="18" flood-color="#4d3319" flood-opacity="0.11"/>
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
    <g transform="translate(130 112)">
      ${card(0, 0, 710, 390, `
        ${eyebrow(44, 64, 'GIFT SUGGESTION WEB SERVICE')}
        ${title(44, 124, ['Подобрать подарок', 'без бесконечного', 'скролла маркетплейсов.'], 56)}
        ${text(44, 280, ['Каталог идей уже открыт без регистрации. Персональный мастер ведёт к auth-flow,', 'потому что recommendation endpoint сейчас требует пользователя.'], { size: 18, fill: C.inkMuted })}
        ${button(44, 320, 196, 'Подобрать подарок', 'primary')}
        ${button(256, 320, 166, 'Каталог идей', 'secondary')}
      `, { shadow: true })}
      ${card(734, 0, 446, 390, `
        ${card(758, 24, 398, 92, `${text(784, 78, 'Каталог', { size: 22, weight: 900 })}${text(784, 104, 'Поиск, категории и карточки подарков.', { size: 15, fill: C.inkMuted })}`, { rx: 18 })}
        ${card(758, 144, 398, 92, `${text(784, 198, 'Auth foundation', { size: 22, weight: 900 })}${text(784, 224, 'Login, register и reset request.', { size: 15, fill: C.inkMuted })}`, { rx: 18 })}
        ${card(758, 264, 398, 92, `${text(784, 318, 'Без выдуманного API', { size: 22, weight: 900 })}${text(784, 344, 'UI следует OpenAPI контракту.', { size: 15, fill: C.inkMuted })}`, { rx: 18 })}
      `, { shadow: true, fill: C.surface })}
      ${eyebrow(0, 462, 'ПОЧЕМУ СЕРВИС ПОЛЕЗЕН')}
      ${title(0, 514, 'Выбор подарка становится понятным сценарием.', 36)}
      ${card(0, 548, 372, 122, `${text(24, 596, 'Объяснимые рекомендации', { size: 23, weight: 700, family: 'Fraunces, Georgia, serif' })}${text(24, 628, 'Причины выбора рядом с карточкой.', { size: 16, fill: C.inkMuted })}`, { shadow: true })}
      ${card(404, 548, 372, 122, `${text(428, 596, 'Каталог из магазинов', { size: 23, weight: 700, family: 'Fraunces, Georgia, serif' })}${text(428, 628, 'Цена, категория и переход к покупке.', { size: 16, fill: C.inkMuted })}`, { shadow: true })}
      ${card(808, 548, 372, 122, `${text(832, 596, 'Готово под wishlist', { size: 23, weight: 700, family: 'Fraunces, Georgia, serif' })}${text(832, 628, 'Паттерны сохранения уже заложены.', { size: 16, fill: C.inkMuted })}`, { shadow: true })}
      ${eyebrow(0, 708, 'ПРЕДПРОСМОТР КАТАЛОГА')}
      ${title(0, 754, 'Несколько свежих идей.', 34)}
      ${giftMini(0, 792, 270, gifts[0], 0)}
      ${giftMini(300, 792, 270, gifts[1], 1)}
      ${giftMini(600, 792, 270, gifts[2], 2)}
      ${giftMini(900, 792, 270, gifts[3], 3)}
    </g>`);
}

function catalog() {
  return shell('02 Catalog final mockup', `
    <g transform="translate(130 122)">
      ${eyebrow(0, 0, 'КАТАЛОГ ИДЕЙ')}
      ${title(0, 58, ['Подберите стартовый список', 'подарков по категории и поиску.'], 46)}
      ${text(0, 174, 'Используются только catalog endpoints: q, category_id и sort.', { size: 18, fill: C.inkMuted })}
      ${card(0, 220, 1180, 124, `
        ${input(26, 28, 530, 'Поиск', 'кофе, хобби, техника', { muted: true })}
        ${input(582, 28, 240, 'Сортировка', 'Сначала новые', { muted: false })}
        ${button(846, 42, 142, 'Искать', 'primary')}
        ${button(1004, 42, 134, 'Сбросить', 'ghost')}
      `, { shadow: true })}
      ${chip(0, 372, 'Все', 72)} ${chip(86, 372, 'Для дома', 114, 'muted')} ${chip(214, 372, 'Техника', 108, 'muted')} ${chip(336, 372, 'Впечатления', 142, 'green')} ${chip(492, 372, 'Хобби', 92, 'muted')} ${chip(598, 372, 'Культура', 110, 'muted')}
      ${text(0, 440, 'Найдено: 126', { size: 16, fill: C.inkMuted })}
      ${text(930, 440, 'Показаны первые 24 результата', { size: 16, fill: C.inkMuted })}
      ${giftCard(0, 468, 360, gifts[0], 0)}
      ${giftCard(410, 468, 360, gifts[1], 1)}
      ${giftCard(820, 468, 360, gifts[2], 2)}
      ${giftCard(0, 808, 360, gifts[3], 3)}
      ${giftCard(410, 808, 360, gifts[4], 4)}
      ${giftCard(820, 808, 360, gifts[5], 5)}
    </g>`);
}

function giftDetail() {
  return shell('03 Gift detail final mockup', `
    <g transform="translate(130 118)">
      ${text(0, 0, '← Вернуться в каталог', { size: 16, fill: C.inkMuted })}
      ${card(0, 38, 1180, 704, `
        ${photo(26, 26, 544, 650, 1)}
        ${chip(616, 58, 'Уют', 82)} ${chip(712, 58, '12+', 64, 'muted')}
        ${title(616, 146, ['Плед из мериноса', 'для спокойных вечеров'], 52)}
        ${text(616, 268, '5 490 ₽', { size: 30, weight: 900 })}
        ${text(616, 330, ['Тёплый подарок с понятной ценностью: уют, качество и отсутствие', 'риска ошибиться с размером. Подходит для близкого человека, которому', 'важен спокойный отдых дома.'], { size: 18, fill: C.inkMuted })}
        ${button(616, 454, 206, 'Перейти к покупке', 'primary')}
        ${button(842, 454, 216, 'Смотреть другие идеи', 'secondary')}
        ${card(616, 550, 500, 92, `${text(642, 590, 'Учтённый gap', { size: 18, weight: 900 })}${text(642, 618, 'Backend отдаёт один store_link, поэтому здесь один purchase CTA.', { size: 15, fill: C.inkMuted })}`, { rx: 16, fill: C.muted })}
      `, { shadow: true })}
      ${footer(802)}
    </g>`);
}

function authPage(kind) {
  const isLogin = kind === 'login';
  const isRegister = kind === 'register';
  const isReset = kind === 'reset';
  const titleLines = isLogin ? ['Войти, чтобы сохранить', 'подборки и wishlist.'] : isRegister ? ['Создать аккаунт', 'для будущих подборок.'] : ['Восстановить доступ', 'через email.'];
  const fileTitle = isLogin ? '04 Login final mockup' : isRegister ? '05 Register final mockup' : '06 Password reset final mockup';
  const formTitle = isLogin ? 'Вход' : isRegister ? 'Регистрация' : 'Сброс пароля';
  const mainButton = isLogin ? 'Войти' : isRegister ? 'Создать аккаунт' : 'Отправить письмо';
  return shell(fileTitle, `
    <g transform="translate(130 136)">
      <g>
        ${eyebrow(0, 0, isReset ? 'PASSWORD RESET' : 'AUTH FOUNDATION')}
        ${title(0, 68, titleLines, 52)}
        ${text(0, 200, ['Формы используют текущие backend auth endpoints и показывают', 'серверные ошибки без локального хранения access token.'], { size: 18, fill: C.inkMuted })}
        ${card(0, 318, 470, 148, `${text(28, 370, 'Состояния формы', { size: 24, family: 'Fraunces, Georgia, serif', weight: 700 })}${text(28, 410, ['default, validation error, server error, loading, success'], { size: 16, fill: C.inkMuted })}`, { shadow: true })}
      </g>
      ${card(628, 0, 552, 560, `
        ${text(672, 64, formTitle, { size: 34, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${isRegister ? input(672, 110, 464, 'Имя', 'Мария', {}) : ''}
        ${input(672, isRegister ? 192 : 130, 464, 'Email', 'maria@example.com', {})}
        ${isReset ? '' : input(672, isRegister ? 274 : 212, 464, 'Пароль', '••••••••••', {})}
        ${isRegister ? input(672, 356, 464, 'Повтор пароля', '••••••••••', {}) : ''}
        ${isLogin ? card(672, 310, 464, 64, `${text(694, 350, 'Неверный email или пароль', { size: 15, fill: C.danger, weight: 800 })}`, { rx: 12, fill: C.dangerSoft, stroke: '#e1b7af' }) : ''}
        ${isReset ? card(672, 220, 464, 70, `${text(694, 260, 'Если email существует, письмо будет отправлено.', { size: 15, fill: C.success, weight: 800 })}`, { rx: 12, fill: C.successSoft, stroke: '#b9d3be' }) : ''}
        ${button(672, isRegister ? 452 : isReset ? 326 : 408, 464, mainButton, 'primary')}
        ${text(672, isRegister ? 538 : isReset ? 414 : 496, isLogin ? 'Нет аккаунта? Регистрация        Забыли пароль?' : isRegister ? 'Уже есть аккаунт? Войти' : 'Вернуться ко входу', { size: 15, fill: C.inkMuted })}
      `, { shadow: true })}
      ${footer(734)}
    </g>`);
}

function wizard() {
  return shell('07 Recommendation wizard final mockup', `
    <g transform="translate(130 116)">
      ${eyebrow(0, 0, 'МАСТЕР ПОДБОРА')}
      ${title(0, 62, ['Ответьте на несколько вопросов,', 'чтобы получить объяснимые идеи.'], 48)}
      ${card(0, 178, 1180, 620, `
        ${chip(42, 38, '1 Повод', 102)} ${chip(158, 38, '2 Получатель', 144, 'green')} ${chip(316, 38, '3 Бюджет', 112, 'muted')} ${chip(442, 38, '4 Интересы', 126, 'muted')} ${chip(582, 38, '5 Проверка', 126, 'muted')}
        ${line(42, 96, 1138, 96)}
        ${text(48, 154, 'Кому выбираем подарок?', { size: 34, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(48, 194, 'Поля соответствуют текущему RecommendRequest. Пол отсутствует в контракте и не выводится.', { size: 17, fill: C.inkMuted })}
        ${card(48, 244, 310, 118, `${text(76, 304, 'Партнёру', { size: 22, weight: 900 })}${text(76, 334, 'Близкий сценарий', { size: 15, fill: C.inkMuted })}`, { rx: 18, fill: C.secondarySoft, stroke: C.secondary })}
        ${card(386, 244, 310, 118, `${text(414, 304, 'Коллеге', { size: 22, weight: 900 })}${text(414, 334, 'Нейтральный тон', { size: 15, fill: C.inkMuted })}`, { rx: 18 })}
        ${card(724, 244, 310, 118, `${text(752, 304, 'Друзьям', { size: 22, weight: 900 })}${text(752, 334, 'Можно ярче', { size: 15, fill: C.inkMuted })}`, { rx: 18 })}
        ${input(48, 416, 310, 'Возраст', '28', {})}
        ${input(386, 416, 310, 'Бюджет', 'до 7 000 ₽', {})}
        ${input(724, 416, 310, 'Повод', 'день рождения', {})}
        ${button(48, 548, 128, 'Назад', 'ghost')}
        ${button(894, 548, 140, 'Дальше', 'primary')}
      `, { shadow: true })}
      ${card(0, 834, 1180, 70, `${text(30, 878, 'Auth-only limitation: submit recommendation доступен после входа пользователя.', { size: 17, fill: C.secondary, weight: 900 })}`, { rx: 18, fill: C.secondarySoft, stroke: '#bdd5ca' })}
    </g>`, { signedIn: true, active: 'Подбор' });
}

function recommendationResults() {
  return shell('08 Recommendation results final mockup', `
    <g transform="translate(130 116)">
      ${eyebrow(0, 0, 'РЕЗУЛЬТАТЫ ПОДБОРА')}
      ${title(0, 60, 'Лучшие идеи с объяснениями.', 48)}
      ${button(970, 10, 210, 'Сохранить подборку', 'secondary')}
      ${giftCard(0, 122, 360, gifts[1], 1)}
      ${giftCard(410, 122, 360, gifts[4], 4)}
      ${giftCard(820, 122, 360, gifts[0], 0)}
      ${card(0, 480, 760, 250, `
        ${text(34, 536, 'Почему эти варианты', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(34, 590, ['Подбор отдаёт score/source и explanation. UI показывает причины выбора', 'рядом с карточками, не скрывая fallback или пустой результат.'], { size: 18, fill: C.inkMuted })}
        ${chip(34, 646, 'budget match', 124, 'green')} ${chip(174, 646, 'interest overlap', 150)} ${chip(340, 646, 'category fit', 126, 'muted')}
      `, { shadow: true })}
      ${card(800, 480, 380, 250, `
        ${text(826, 536, 'Альтернативы', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(826, 590, ['Если первый вариант не подходит,', 'показываем соседние категории.'], { size: 17, fill: C.inkMuted })}
        ${button(826, 654, 156, 'Открыть', 'ghost', { height: 44 })}
      `, { shadow: true })}
      ${card(0, 762, 1180, 76, `${text(30, 810, 'Wishlist state не приходит в recommendation payload: сохранение требует отдельного wishlist action.', { size: 17, fill: C.primary, weight: 900 })}`, { rx: 18, fill: '#f3e3d8', stroke: '#e1c3ad' })}
    </g>`, { signedIn: true, active: 'Подбор' });
}

function profile() {
  return shell('09 Profile final mockup', `
    <g transform="translate(130 118)">
      ${eyebrow(0, 0, 'ПРОФИЛЬ')}
      ${title(0, 58, 'Личный кабинет пользователя.', 48)}
      ${card(0, 130, 560, 420, `
        ${circle(48, 70, 38, { fill: C.secondarySoft, stroke: C.secondary })}
        ${text(106, 64, 'Мария Иванова', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(106, 96, 'maria@example.com', { size: 16, fill: C.inkMuted })}
        ${input(48, 154, 464, 'Отображаемое имя', 'Мария Иванова', {})}
        ${input(48, 246, 464, 'Email', 'maria@example.com', {})}
        ${button(48, 344, 180, 'Сохранить', 'primary')}
      `, { shadow: true })}
      ${card(620, 130, 560, 420, `
        ${text(660, 188, 'Быстрые переходы', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${card(660, 230, 460, 82, `${text(688, 278, 'Wishlist', { size: 22, weight: 900 })}${text(1010, 278, '12 идей', { size: 16, fill: C.inkMuted })}`, { rx: 16 })}
        ${card(660, 332, 460, 82, `${text(688, 380, 'Интеграции', { size: 22, weight: 900 })}${text(1000, 380, 'VK planned', { size: 16, fill: C.inkMuted })}`, { rx: 16 })}
        ${text(660, 474, 'getCurrentUser и updateCurrentUser покрывают этот экран.', { size: 16, fill: C.inkMuted })}
      `, { shadow: true })}
    </g>`, { signedIn: true, active: 'Профиль' });
}

function wishlist() {
  return shell('10 Wishlist final mockup', `
    <g transform="translate(130 116)">
      ${eyebrow(0, 0, 'WISHLIST')}
      ${title(0, 58, 'Сохранённые идеи подарков.', 48)}
      ${button(980, 8, 200, 'Новый список', 'primary')}
      ${card(0, 130, 310, 530, `
        ${text(28, 188, 'Списки', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${card(28, 228, 254, 76, `${text(52, 274, 'День рождения', { size: 19, weight: 900 })}${text(220, 274, '5', { size: 18, weight: 900, fill: C.primary })}`, { rx: 14, fill: C.secondarySoft, stroke: C.secondary })}
        ${card(28, 322, 254, 76, `${text(52, 368, 'Новый год', { size: 19, weight: 900 })}${text(228, 368, '7', { size: 18, fill: C.inkMuted })}`, { rx: 14 })}
        ${card(28, 416, 254, 76, `${text(52, 462, 'Коллегам', { size: 19, weight: 900 })}${text(228, 462, '2', { size: 18, fill: C.inkMuted })}`, { rx: 14 })}
      `, { shadow: true })}
      ${giftCard(350, 130, 380, gifts[0], 0)}
      ${giftCard(770, 130, 380, gifts[1], 1)}
      ${giftCard(350, 470, 380, gifts[5], 5)}
      ${card(770, 470, 380, 310, `${text(804, 546, 'Пустой список', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}${text(804, 594, ['Когда список пустой, показываем CTA', 'в каталог без фиктивных рекомендаций.'], { size: 17, fill: C.inkMuted })}${button(804, 674, 180, 'Открыть каталог', 'secondary')}`, { shadow: true })}
    </g>`, { signedIn: true, active: 'Wishlist' });
}

function integrations() {
  return shell('11 Integrations VK final mockup', `
    <g transform="translate(130 118)">
      ${eyebrow(0, 0, 'ИНТЕГРАЦИИ')}
      ${title(0, 58, 'VK интересы как planned state.', 48)}
      ${card(0, 140, 1180, 438, `
        ${circle(58, 76, 40, { fill: '#dbe7ff', stroke: '#8aa0d8' })}
        ${text(124, 84, 'VK Connect', { size: 34, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${chip(990, 58, 'planned', 104, 'muted')}
        ${text(124, 132, ['Экран показывает будущий entry point, но connect/sync actions заблокированы,', 'потому что в OpenAPI и backend-коде нет VK HTTP endpoints.'], { size: 18, fill: C.inkMuted })}
        ${button(124, 226, 180, 'Подключить VK', 'disabled', { opacity: 0.72 })}
        ${button(324, 226, 220, 'Синхронизировать', 'disabled', { opacity: 0.72 })}
        ${card(124, 326, 820, 70, `${text(150, 370, 'Gap зафиксирован: нет endpoint для OAuth connect, status и sync interests.', { size: 17, fill: C.danger, weight: 900 })}`, { rx: 16, fill: C.dangerSoft, stroke: '#e1b7af' })}
      `, { shadow: true })}
    </g>`, { signedIn: true, active: 'Интеграции' });
}

function adminImport() {
  return shell('12 Admin import final mockup', `
    <g transform="translate(130 116)">
      ${eyebrow(0, 0, 'ADMIN')}
      ${title(0, 58, 'Импорт каталога подарков.', 48)}
      ${card(0, 132, 710, 520, `
        ${rect(42, 52, 626, 234, { rx: 24, fill: C.muted, stroke: C.border, strokeWidth: 2 })}
        ${text(264, 154, 'Перетащите CSV или XLSX', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700, anchor: 'middle' })}
        ${text(355, 196, 'Файл не отправляется до подтверждения', { size: 16, fill: C.inkMuted, anchor: 'middle' })}
        ${button(250, 226, 210, 'Выбрать файл', 'secondary')}
        ${input(42, 332, 300, 'Source label', 'marketplace_april.csv', {})}
        ${button(372, 346, 210, 'Запустить импорт', 'primary')}
      `, { shadow: true })}
      ${card(760, 132, 420, 520, `
        ${text(792, 194, 'Последний job', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${chip(792, 226, 'processing', 120, 'green')}
        ${text(792, 294, 'Записей обработано', { size: 16, fill: C.inkMuted })}
        ${text(792, 340, '1 248 / 1 540', { size: 42, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${rect(792, 382, 320, 12, { rx: 999, fill: C.muted })}
        ${rect(792, 382, 238, 12, { rx: 999, fill: C.primary })}
        ${button(792, 442, 200, 'Открыть статус', 'ghost')}
      `, { shadow: true })}
    </g>`, { signedIn: true, active: 'Admin' });
}

function importStatus() {
  return shell('13 Import status errors final mockup', `
    <g transform="translate(130 116)">
      ${eyebrow(0, 0, 'IMPORT JOB')}
      ${title(0, 58, 'Статус импорта и ошибки.', 48)}
      ${card(0, 130, 1180, 178, `
        ${chip(34, 42, 'completed_with_errors', 206)}
        ${text(34, 112, '1 502 обработано', { size: 32, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(390, 112, '38 ошибок', { size: 32, family: 'Fraunces, Georgia, serif', weight: 700, fill: C.danger })}
        ${text(720, 112, 'Файл: marketplace_april.csv', { size: 18, fill: C.inkMuted })}
      `, { shadow: true })}
      ${card(0, 354, 1180, 458, `
        ${text(34, 412, 'Ошибки импорта', { size: 30, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${line(34, 458, 1146, 458)}
        ${text(34, 502, 'Строка', { size: 15, weight: 900, fill: C.inkMuted })}${text(170, 502, 'Поле', { size: 15, weight: 900, fill: C.inkMuted })}${text(330, 502, 'Сообщение', { size: 15, weight: 900, fill: C.inkMuted })}
        ${line(34, 526, 1146, 526)}
        ${text(34, 574, '24', { size: 17 })}${text(170, 574, 'price', { size: 17 })}${text(330, 574, 'Цена должна быть положительным числом', { size: 17, fill: C.inkMuted })}
        ${line(34, 604, 1146, 604)}
        ${text(34, 652, '41', { size: 17 })}${text(170, 652, 'store_link', { size: 17 })}${text(330, 652, 'Некорректный URL магазина', { size: 17, fill: C.inkMuted })}
        ${line(34, 682, 1146, 682)}
        ${text(34, 730, '77', { size: 17 })}${text(170, 730, 'category', { size: 17 })}${text(330, 730, 'Категория не найдена', { size: 17, fill: C.inkMuted })}
      `, { shadow: true })}
    </g>`, { signedIn: true, active: 'Admin' });
}

function states() {
  return shell('14 System states final mockup', `
    <g transform="translate(130 116)">
      ${eyebrow(0, 0, 'SYSTEM STATES')}
      ${title(0, 58, 'Loading, empty, error и success.', 48)}
      ${card(0, 130, 360, 260, `
        ${text(28, 188, 'Loading', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${rect(28, 226, 304, 18, { rx: 999, fill: C.muted })}
        ${rect(28, 264, 240, 18, { rx: 999, fill: '#f8efe2' })}
        ${rect(28, 302, 304, 72, { rx: 18, fill: C.muted })}
      `, { shadow: true })}
      ${card(410, 130, 360, 260, `
        ${text(438, 188, 'Empty', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(438, 238, ['Каталог ничего не вернул.', 'Предлагаем сбросить фильтры.'], { size: 17, fill: C.inkMuted })}
        ${button(438, 312, 170, 'Сбросить', 'secondary')}
      `, { shadow: true })}
      ${card(820, 130, 360, 260, `
        ${text(848, 188, 'Error', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${card(848, 226, 280, 72, `${text(868, 268, 'Не удалось загрузить данные', { size: 16, fill: C.danger, weight: 900 })}`, { rx: 12, fill: C.dangerSoft, stroke: '#e1b7af' })}
        ${button(848, 324, 180, 'Повторить', 'primary')}
      `, { shadow: true })}
      ${card(0, 438, 360, 260, `
        ${text(28, 496, 'Success', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${card(28, 534, 300, 70, `${text(50, 576, 'Письмо отправлено', { size: 16, fill: C.success, weight: 900 })}`, { rx: 12, fill: C.successSoft, stroke: '#b9d3be' })}
      `, { shadow: true })}
      ${card(410, 438, 360, 260, `
        ${text(438, 496, 'Unauthorized', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(438, 546, ['Для wishlist и подборок', 'нужно войти в аккаунт.'], { size: 17, fill: C.inkMuted })}
        ${button(438, 620, 150, 'Войти', 'primary')}
      `, { shadow: true })}
      ${card(820, 438, 360, 260, `
        ${text(848, 496, 'Not found', { size: 28, family: 'Fraunces, Georgia, serif', weight: 700 })}
        ${text(848, 546, ['Страница или подарок', 'не найдены.'], { size: 17, fill: C.inkMuted })}
        ${button(848, 620, 190, 'В каталог', 'secondary')}
      `, { shadow: true })}
      ${card(0, 748, 1180, 84, `${text(30, 802, 'Все состояния используют те же banner, empty-state, skeleton и button паттерны, что текущий frontend Slice 1.', { size: 17, fill: C.inkMuted, weight: 800 })}`, { rx: 18 })}
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
  .map(([group, items]) => `<section class="preview-section">
    <h2>${esc(group)}</h2>
    <div class="mockup-grid">
      ${items.map((item) => `<article class="mockup-card">
        <div class="mockup-card__header">
          <span>${esc(item.label)}</span>
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
      header {
        max-width: 1180px;
        margin: 0 auto;
        padding: 44px 24px 28px;
      }
      h1, h2 {
        margin: 0;
        font-family: Fraunces, Georgia, serif;
        letter-spacing: -0.03em;
      }
      h1 { max-width: 820px; font-size: clamp(42px, 6vw, 82px); line-height: 0.95; }
      h2 { font-size: 34px; }
      p { max-width: 720px; color: var(--ink-muted); font-size: 18px; line-height: 1.55; }
      .toolbar {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        margin-top: 24px;
      }
      .pill {
        border: 1px solid var(--border);
        border-radius: 999px;
        background: rgba(255,253,248,0.86);
        padding: 10px 14px;
        color: var(--ink-muted);
        font-weight: 700;
      }
      main {
        display: grid;
        gap: 48px;
        max-width: 1180px;
        margin: 0 auto;
        padding: 0 24px 72px;
      }
      .preview-section {
        display: grid;
        gap: 18px;
      }
      .mockup-grid {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 22px;
      }
      .mockup-card {
        overflow: hidden;
        border: 1px solid rgba(216, 200, 176, 0.82);
        border-radius: 24px;
        background: rgba(255,253,248,0.94);
        box-shadow: var(--shadow);
      }
      .mockup-card__header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        border-bottom: 1px solid rgba(216, 200, 176, 0.72);
        padding: 14px 18px;
        font-weight: 800;
      }
      .mockup-card__header a { color: var(--primary); text-decoration: none; }
      img {
        display: block;
        width: 100%;
        background: var(--surface-muted);
      }
      @media (max-width: 860px) {
        .mockup-grid { grid-template-columns: 1fr; }
      }
    </style>
  </head>
  <body>
    <header>
      <h1>Final site mockups for Gift Suggestion</h1>
      <p>Статичные Figma-ready SVG макеты, собранные по текущим frontend токенам и UI-паттернам: тёплый фон, акцент #C65A1E, secondary #245C4A, rounded cards, e-commerce сетка и auth form patterns.</p>
      <div class="toolbar">
        <span class="pill">Desktop SVG 1440x1024</span>
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
- Borders: \`#d8c8b0\`
- Editorial commerce tone with serif headings, rounded cards, catalog grids, gift cards, auth form blocks, banners and skeleton states.

Typography mirrors the frontend CSS intent:

- Headings: \`Fraunces, Georgia, serif\`
- Body/UI: \`Manrope, Segoe UI, sans-serif\`
- Numeric emphasis follows the current compact UI treatment.

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
4. Place frames in the specified order and use each SVG file name as the frame caption.
5. For mobile review, create \`390x844\` companion frames using the layout notes in the spec.

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

console.log(`Rendered ${mockups.length} mockups into ${mockupsDir}`);
