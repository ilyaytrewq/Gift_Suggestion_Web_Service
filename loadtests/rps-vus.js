import http from 'k6/http';
import { check } from 'k6';

import { apiPrefix, jsonHeaders } from './lib/config.js';
import { buildRecommendPayload } from './lib/payloads.js';

const TARGET = __ENV.K6_PROBE_TARGET || 'catalog';
const VUS = Number(__ENV.K6_VUS || 10);
const DURATION = __ENV.K6_DURATION || '45s';

export const options = {
  scenarios: {
    parallel_vus: {
      executor: 'constant-vus',
      vus: VUS,
      duration: DURATION,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.02'],
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

  const res = http.get(`${apiPrefix}/catalog/gifts?limit=12&has_image=true`, {
    headers: jsonHeaders(),
    tags: { endpoint: 'catalog_list' },
  });
  check(res, { ok: (r) => r.status >= 200 && r.status < 300 });
}
