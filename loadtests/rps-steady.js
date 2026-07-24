import http from 'k6/http';
import { check } from 'k6';

import { apiPrefix, baseUrl, jsonHeaders } from './lib/config.js';
import { buildRecommendPayload } from './lib/payloads.js';

const TARGET = __ENV.K6_PROBE_TARGET || 'catalog';
const RATE = Number(__ENV.K6_RATE || 10);
const DURATION = __ENV.K6_DURATION || '60s';

export const options = {
  scenarios: {
    steady: {
      executor: 'constant-arrival-rate',
      rate: RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: Number(__ENV.K6_PREALLOCATED_VUS || 50),
      maxVUs: Number(__ENV.K6_MAX_VUS || 100),
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.02'],
    http_req_duration: [`p(95)<${__ENV.K6_P95_MS || 3000}`],
  },
};

export default function () {
  if (TARGET === 'recommendations') {
    const res = http.post(`${apiPrefix}/recommendations`, buildRecommendPayload({ top_n: 12 }), {
      headers: jsonHeaders(),
      tags: { endpoint: 'recommendations' },
    });
    check(res, { ok: (r) => r.status >= 200 && r.status < 300 });
    return;
  }

  if (TARGET === 'health') {
    const res = http.get(`${baseUrl}/health/live`, { tags: { endpoint: 'health' } });
    check(res, { ok: (r) => r.status === 200 });
    return;
  }

  const res = http.get(`${apiPrefix}/catalog/gifts?limit=12&has_image=true`, {
    headers: jsonHeaders(),
    tags: { endpoint: 'catalog_list' },
  });
  check(res, { ok: (r) => r.status >= 200 && r.status < 300 });
}
