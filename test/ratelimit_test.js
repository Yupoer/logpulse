import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

// Custom metrics to track rate limit behavior
const allowed = new Counter('requests_allowed');
const rateLimited = new Counter('requests_rate_limited');

// Test configuration - designed to trigger rate limiting
export const options = {
    scenarios: {
        // Single scenario: sustained load to test rate limiter
        rate_limit_test: {
            executor: 'constant-vus',
            vus: 20,              // 20 concurrent users
            duration: '30s',      // Run for 30 seconds
        },
    },
    thresholds: {
        'http_req_duration': ['p(95)<500'],
        'requests_allowed': ['count>0'],
        'requests_rate_limited': ['count>0'], // We EXPECT some 429s
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost';

export default function () {
    const res = http.get(`${BASE_URL}/ping`);

    const checks = check(res, {
        'status is 200 (allowed)': (r) => r.status === 200,
        'status is 429 (rate limited)': (r) => r.status === 429,
    });

    // Track metrics
    if (res.status === 200) {
        allowed.add(1);
    } else if (res.status === 429) {
        rateLimited.add(1);
    }

    // Small delay between requests (but still fast enough to hit limit)
    sleep(0.05); // 50ms = ~20 requests/sec per VU
}

export function handleSummary(data) {
    const allowedCount = data.metrics.requests_allowed?.values?.count || 0;
    const limitedCount = data.metrics.requests_rate_limited?.values?.count || 0;
    const total = allowedCount + limitedCount;
    const allowedPercent = total > 0 ? ((allowedCount / total) * 100).toFixed(2) : 0;
    const limitedPercent = total > 0 ? ((limitedCount / total) * 100).toFixed(2) : 0;

    console.log('\n========================================');
    console.log('Rate Limiting Test Results');
    console.log('========================================');
    console.log(`Allowed (200):       ${allowedCount.toString().padStart(6)} requests (${allowedPercent}%)`);
    console.log(`Rate Limited (429):  ${limitedCount.toString().padStart(6)} requests (${limitedPercent}%)`);
    console.log(`Total Requests:      ${total.toString().padStart(6)}`);
    console.log('========================================');
    console.log('With default config (capacity=100, rate=50/sec):');
    console.log('   - First ~100 requests use burst capacity');
    console.log('   - After that, ~50 requests/sec are allowed');
    console.log('   - VUs: 20 x 20 req/sec = 400 req/sec attempted');
    console.log('   - Expected: ~12.5% allowed, ~87.5% rate limited');
    console.log('========================================\n');

    return {
        'stdout': '', // k6 handles standard output
    };
}
