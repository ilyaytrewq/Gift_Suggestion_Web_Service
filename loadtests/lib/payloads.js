const OCCASIONS = [
  'birthday',
  'new_year',
  'wedding',
  'graduation',
  'housewarming',
  'anniversary',
];

const RELATIONSHIPS = [
  'friend',
  'partner',
  'parent',
  'colleague',
  'sibling',
  'child',
];

const INTEREST_POOL = [
  'books',
  'sports',
  'music',
  'cooking',
  'travel',
  'games',
  'tech',
  'art',
  'fitness',
  'photography',
];

const BUDGETS = ['1500.00', '3000.00', '5000.00', '10000.00', '25000.00'];

const GENDERS = ['male', 'female', 'other'];

function sampleInterests() {
  const count = 1 + Math.floor(Math.random() * 4);
  const picked = [];
  const pool = [...INTEREST_POOL];
  for (let i = 0; i < count && pool.length > 0; i += 1) {
    const idx = Math.floor(Math.random() * pool.length);
    picked.push(pool.splice(idx, 1)[0]);
  }
  return picked;
}

export function buildRecommendPayload(overrides = {}) {
  const topN = Math.random() < 0.3 ? 12 : 24;
  const payload = {
    occasion: OCCASIONS[Math.floor(Math.random() * OCCASIONS.length)],
    relationship: RELATIONSHIPS[Math.floor(Math.random() * RELATIONSHIPS.length)],
    recipient_age: 18 + Math.floor(Math.random() * 50),
    recipient_gender: GENDERS[Math.floor(Math.random() * GENDERS.length)],
    budget_max: BUDGETS[Math.floor(Math.random() * BUDGETS.length)],
    interests: sampleInterests(),
    top_n: topN,
    use_wishlist_context: false,
    ...overrides,
  };
  return JSON.stringify(payload);
}

export function catalogListQuery() {
  const limit = [12, 20, 24][Math.floor(Math.random() * 3)];
  const offset = Math.floor(Math.random() * 3) * limit;
  const maxPrice = BUDGETS[Math.floor(Math.random() * BUDGETS.length)];
  const parts = [
    `limit=${limit}`,
    `offset=${offset}`,
    `max_price=${maxPrice}`,
    'has_image=true',
    'sort=price_asc',
  ];
  if (Math.random() < 0.5) {
    const q = ['книга', 'игра', 'набор', 'подарок'][Math.floor(Math.random() * 4)];
    parts.push(`q=${encodeURIComponent(q)}`);
  }
  return parts.join('&');
}
