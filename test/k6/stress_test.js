import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// ==========================================
// Custom Metrics
// ==========================================
const writeErrors = new Counter('write_errors');
const readErrors = new Counter('read_errors');
const searchErrors = new Counter('search_errors');
const writeSuccessRate = new Rate('write_success_rate');
const readSuccessRate = new Rate('read_success_rate');
const searchSuccessRate = new Rate('search_success_rate');
const writeDuration = new Trend('write_duration');
const readDuration = new Trend('read_duration');
const searchDuration = new Trend('search_duration');

// ==========================================
// Configuration
// ==========================================
const BASE_URL = __ENV.BASE_URL || 'http://localhost';

// Test scenarios - high concurrency read + write
export const options = {
    scenarios: {
        // High Write Load 
        write_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 200 },   // Ramp up to 50 VUs
                { duration: '1m', target: 500 },   // Ramp up to 100 VUs  
                { duration: '5m', target: 500 },   // Stay at 100 VUs
                { duration: '30s', target: 0 },    // Ramp down
            ],
            exec: 'writeScenario',
        },
        // High Read Load 
        read_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 30 },   // Ramp up to 30 VUs
                { duration: '1m', target: 60 },    // Ramp up to 60 VUs
                { duration: '2m', target: 60 },    // Stay at 60 VUs
                { duration: '30s', target: 0 },    // Ramp down
            ],
            exec: 'readScenario',
            startTime: '10s', // Start 10s after writes begin
        },
        // Search Load 
        search_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 20 },   // Ramp up to 20 VUs
                { duration: '1m', target: 40 },    // Ramp up to 40 VUs
                { duration: '2m', target: 40 },    // Stay at 40 VUs
                { duration: '30s', target: 0 },    // Ramp down
            ],
            exec: 'searchScenario',
            startTime: '20s', // Start 20s after writes begin
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<1000'],      // 95% requests < 500ms
        write_success_rate: ['rate>0.90'],     // 95% write success
        read_success_rate: ['rate>0.90'],      // 90% read success
        search_success_rate: ['rate>0.90'],    // 90% search success
    },
};

// ==========================================
// Test Data Generators
// ==========================================
const services = ['auth-service', 'payment-service', 'order-service', 'user-service', 'notification-service'];
const levels = ['INFO', 'WARN', 'ERROR', 'DEBUG'];
const messages = [
    'User login successful via OAuth',
    'Database connection timeout during transaction',
    'Processing order items',
    'Cache miss, fetching from database',
    'Request validation failed',
    'Payment processed successfully',
    'Session expired for user',
    'API rate limit exceeded',
    'File upload completed',
    'Background job started',
];

function generateLogEntry() {
    return {
        service_name: services[Math.floor(Math.random() * services.length)],
        level: levels[Math.floor(Math.random() * levels.length)],
        message: `${messages[Math.floor(Math.random() * messages.length)]} - ${Date.now()}`,
        timestamp: new Date().toISOString(),
    };
}

const searchKeywords = ['login', 'timeout', 'order', 'payment', 'ERROR', 'auth-service', 'user'];

// ==========================================
// Scenarios
// ==========================================

// Write Scenario - POST /logs
export function writeScenario() {
    const payload = JSON.stringify(generateLogEntry());
    const params = {
        headers: { 'Content-Type': 'application/json' },
    };

    const startTime = Date.now();
    const res = http.post(`${BASE_URL}/logs`, payload, params);
    writeDuration.add(Date.now() - startTime);

    const success = check(res, {
        'write status is 201': (r) => r.status === 201,
    });

    writeSuccessRate.add(success);
    if (!success) {
        writeErrors.add(1);
    }

    sleep(Math.random() * 0.1); // 0-0.5s random delay
}

// Read Scenario - GET /logs/:id
export function readScenario() {
    // Random ID between 1-1000 (assumes some logs exist)
    const id = Math.floor(Math.random() * 1000) + 1;

    const startTime = Date.now();
    const res = http.get(`${BASE_URL}/logs/${id}`);
    readDuration.add(Date.now() - startTime);

    const success = check(res, {
        'read status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    });

    readSuccessRate.add(success);
    if (!success) {
        readErrors.add(1);
    }

    sleep(Math.random() * 0.3); // 0-0.3s random delay
}

// Search Scenario - GET /logs/search?q=keyword
export function searchScenario() {
    const keyword = searchKeywords[Math.floor(Math.random() * searchKeywords.length)];

    const startTime = Date.now();
    const res = http.get(`${BASE_URL}/logs/search?q=${keyword}`);
    searchDuration.add(Date.now() - startTime);

    const success = check(res, {
        'search status is 200': (r) => r.status === 200,
        'search returns array': (r) => r.json('data') !== undefined,
    });

    searchSuccessRate.add(success);
    if (!success) {
        searchErrors.add(1);
    }

    sleep(Math.random() * 0.5); // 0-0.5s random delay
}

// ==========================================
// Lifecycle Hooks
// ==========================================
export function setup() {
    // Verify server is running
    const res = http.get(`${BASE_URL}/ping`);
    if (res.status !== 200) {
        throw new Error(`Server not responding at ${BASE_URL}`);
    }
    console.log(`Server is ready at ${BASE_URL}`);

    // Seed some initial logs for read tests
    console.log('Seeding initial logs...');
    for (let i = 0; i < 100; i++) {
        const payload = JSON.stringify(generateLogEntry());
        http.post(`${BASE_URL}/logs`, payload, {
            headers: { 'Content-Type': 'application/json' },
        });
    }
    console.log('Seeded 100 initial logs');
}

export function teardown(data) {
    console.log('Stress test completed!');
}
