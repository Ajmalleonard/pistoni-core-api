import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    burst_1k_rps: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s',
      duration: '15s',
      preAllocatedVUs: 120,
      maxVUs: 1200,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<600'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:4000';

export default function () {
  const pick = Math.random();
  let res;

  if (pick < 0.7) {
    res = http.get(`${BASE_URL}/health`);
  } else if (pick < 0.9) {
    res = http.get(`${BASE_URL}/providers/drivers`);
  } else {
    res = http.get(`${BASE_URL}/providers/body-guards`);
  }

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(0.001);
}
