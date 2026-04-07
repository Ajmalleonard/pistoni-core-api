import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    sustained_800_rps_60s: {
      executor: 'constant-arrival-rate',
      rate: 800,
      timeUnit: '1s',
      duration: '60s',
      preAllocatedVUs: 120,
      maxVUs: 1200,
    },
    spike_2000_rps_20s: {
      executor: 'constant-arrival-rate',
      startTime: '65s',
      rate: 2000,
      timeUnit: '1s',
      duration: '20s',
      preAllocatedVUs: 200,
      maxVUs: 2000,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.02'],
    http_req_duration: ['p(95)<1200', 'p(99)<2500'],
    checks: ['rate>0.99'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:1010';

const routes = [
  { path: '/health', weight: 20 },
  { path: '/providers/drivers', weight: 30 },
  { path: '/providers/body-guards', weight: 25 },
  { path: '/cars', weight: 15 },
  { path: '/users', weight: 10 },
];

function weightedPick() {
  const total = routes.reduce((sum, r) => sum + r.weight, 0);
  let roll = Math.random() * total;
  for (const r of routes) {
    roll -= r.weight;
    if (roll <= 0) return r.path;
  }
  return routes[0].path;
}

export default function () {
  const path = weightedPick();
  const res = http.get(`${BASE_URL}${path}`, {
    tags: { endpoint: path },
    timeout: '5s',
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(0.001);
}
