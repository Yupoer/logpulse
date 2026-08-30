import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  scenarios: {
    ingest: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },
        { duration: '2m', target: 10 },
        { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1000'],
  },
};

export default function () {
  const response = http.post(
    `${BASE_URL}/logs`,
    JSON.stringify({
      service_name: 'k6-benchmark',
      level: 'INFO',
      message: `benchmark-${Date.now()}`,
      timestamp: new Date().toISOString(),
    }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  check(response, { 'log accepted': (res) => res.status === 201 });
}
