import { defaultThresholds } from './lib/config.js';
import { loadTestData } from './lib/setup.js';
import { recommendFlow, thinkTime } from './lib/scenarios.js';

export const options = {
  scenarios: {
    recommendations_stress: {
      executor: 'ramping-arrival-rate',
      startRate: Number(__ENV.K6_START_RATE || 2),
      timeUnit: '1s',
      preAllocatedVUs: Number(__ENV.K6_PREALLOCATED_VUS || 20),
      maxVUs: Number(__ENV.K6_MAX_VUS || 60),
      stages: [
        { duration: __ENV.K6_RAMP_UP || '1m', target: Number(__ENV.K6_TARGET_RATE || 10) },
        { duration: __ENV.K6_STEADY || '2m', target: Number(__ENV.K6_TARGET_RATE || 10) },
        { duration: __ENV.K6_RAMP_DOWN || '30s', target: 0 },
      },
    },
  },
  thresholds: {
    ...defaultThresholds(),
    'http_req_duration{endpoint:recommendations}': ['p(95)<12000'],
  },
};

export function setup() {
  return loadTestData();
}

export default function (data) {
  recommendFlow(data);
  thinkTime(0.1, 0.5);
}
