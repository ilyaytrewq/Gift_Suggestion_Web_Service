import { check } from 'k6';

export function checkOk(res, name) {
  return check(res, {
    [`${name} status 2xx`]: (r) => r.status >= 200 && r.status < 300,
    [`${name} has json body`]: (r) => {
      const ct = r.headers['Content-Type'] || r.headers['content-type'] || '';
      return ct.includes('application/json') || (r.body && r.body.length > 0);
    },
  });
}

export function parseJson(res) {
  try {
    return res.json();
  } catch {
    return null;
  }
}

export function extractGiftIds(body) {
  const items = body?.data?.items;
  if (!Array.isArray(items) || items.length === 0) {
    return [];
  }
  return items.map((item) => item.id).filter(Boolean);
}

export function pickRandom(list) {
  if (!list || list.length === 0) {
    return null;
  }
  return list[Math.floor(Math.random() * list.length)];
}

export function weightedPick(entries) {
  const total = entries.reduce((sum, entry) => sum + entry.weight, 0);
  let roll = Math.random() * total;
  for (const entry of entries) {
    roll -= entry.weight;
    if (roll <= 0) {
      return entry.fn;
    }
  }
  return entries[entries.length - 1].fn;
}
