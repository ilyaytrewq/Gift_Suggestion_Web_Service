const DEFAULT_BASE_URL = 'http://localhost:8080';

export const baseUrl = (__ENV.BASE_URL || DEFAULT_BASE_URL).replace(/\/$/, '');
export const apiPrefix = `${baseUrl}/api/v1`;

export const accessToken = __ENV.K6_ACCESS_TOKEN || '';
export const testEmail = __ENV.K6_TEST_EMAIL || '';
export const testPassword = __ENV.K6_TEST_PASSWORD || '';

export const defaultHeaders = {
  Accept: 'application/json',
};

export function jsonHeaders(extra = {}) {
  return {
    ...defaultHeaders,
    'Content-Type': 'application/json',
    ...extra,
  };
}

export function authHeaders(token) {
  if (!token) {
    return jsonHeaders();
  }
  return jsonHeaders({ Authorization: `Bearer ${token}` });
}

export function defaultThresholds() {
  return {
    http_req_failed: ['rate<0.02'],
    http_req_duration: ['p(95)<3000'],
    'http_req_duration{endpoint:health}': ['p(95)<500'],
    'http_req_duration{endpoint:catalog_list}': ['p(95)<2000'],
    'http_req_duration{endpoint:catalog_gift}': ['p(95)<1500'],
    'http_req_duration{endpoint:recommendations}': ['p(95)<8000'],
  };
}

export function smokeThresholds() {
  return {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<5000'],
  };
}
