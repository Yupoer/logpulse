# LogPulse

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)]()

> A high-throughput, distributed log aggregation system engineered with Go, Kafka, Elasticsearch, and Redis.

## Introduction

LogPulse is a cloud-native backend solution designed to handle massive scale log ingestion and real-time analytics. It addresses the challenge of "Peak Shaving" in high-concurrency scenarios by decoupling the ingestion layer from the storage layer.

Built with a Microservices mindset, LogPulse ensures data consistency and system resilience through asynchronous messaging and a robust caching strategy.

### Dashboard Preview

> Real-time log ingestion visualization in Kibana. This demonstrates the successful flow of data from the **Go API** through **Kafka** to **Elasticsearch**.

![Kibana Discover Dashboard](assets/kibana_discover_logs.png)
![Kibana Discover Dashboard](assets/kibana_discover_logs_expand.png)

### Performance Benchmarks

> Load testing results using [k6](https://k6.io/) with up to **600 concurrent VUs** over a 7-minute stress test.

#### Stress Test Visualizations

| Read/Write/Search Performance | Write Stress Test |
|:-----------------------------:|:-----------------:|
| ![Read Write Search](assets/read_write_search.png) | ![Write Stress Test](assets/write_stress_test.png) |

#### k6 Load Test Results

| Metric | Result |
|--------|--------|
| **Total Requests** | 743,453 |
| **Throughput** | 1,768 req/s |
| **Write Success Rate** | 99.93% |
| **Read Success Rate** | 100.00% |
| **Search Success Rate** | 100.00% |
| **P95 Latency** | 292.41ms |
| **Avg Response Time** | 211.63ms |

<details>
<summary>Detailed Test Report</summary>

```
█ THRESHOLDS

  http_req_duration
  ✓ 'p(95)<1000' p(95)=292.41ms

  read_success_rate
  ✓ 'rate>0.90' rate=100.00%

  search_success_rate
  ✓ 'rate>0.90' rate=100.00%

  write_success_rate
  ✓ 'rate>0.90' rate=99.93%

█ TOTAL RESULTS

  checks_total.......: 756080 1798.735725/s
  checks_succeeded...: 99.94% 755651 out of 756080
  checks_failed......: 0.05%  429 out of 756080

  ✗ write status is 201
    ↳  99% — ✓ 702559 / ✗ 429
  ✓ read status is 200 or 404
  ✓ search status is 200
  ✓ search returns array

  CUSTOM
  read_duration..................: avg=256ms   p(95)=314ms
  search_duration................: avg=339ms   p(95)=400ms
  write_duration.................: avg=207ms   p(95)=291ms

  HTTP
  http_req_duration..............: avg=211.63ms p(95)=292.41ms
  http_req_failed................: 0.05% 429 out of 743453
  http_reqs......................: 743453 1768.695735/s

  EXECUTION
  iterations.....................: 743352 1768.455453/s
  vus............................: max=600
  running time...................: 7m00.3s
```

</details>

#### Run Sample Stress Test

We use [k6](https://k6.io/) for load testing. Install k6 and run:

```bash
# Install k6 (macOS)
brew install k6

# Install k6 (Windows via Chocolatey)
choco install k6

# Run the stress test (ensure the app is running first)
k6 run stress_test.js
```

<details>
<summary>Test Scenarios Configuration</summary>

The stress test includes 3 concurrent scenarios:

| Scenario | Peak VUs | Duration | Description |
|----------|----------|----------|-------------|
| **write_load** | 500 | 7 min | High-throughput log ingestion via `POST /logs` |
| **read_load** | 60 | 4 min | Random log retrieval via `GET /logs/:id` |
| **search_load** | 40 | 4 min | Keyword search via `GET /logs/search?q=` |

**Thresholds:**
- P95 response time < 1000ms
- Write/Read/Search success rate > 90%

</details>

<details>
<summary>Test Data Samples</summary>

```javascript
// Services tested
const services = ['auth-service', 'payment-service', 'order-service', 'user-service', 'notification-service'];

// Log levels
const levels = ['INFO', 'WARN', 'ERROR', 'DEBUG'];

// Sample messages
const messages = [
    'User login successful via OAuth',
    'Database connection timeout during transaction',
    'Processing order items',
    'Cache miss, fetching from database',
    'Payment processed successfully',
    // ...
];

// Search keywords
const searchKeywords = ['login', 'timeout', 'order', 'payment', 'ERROR', 'auth-service', 'user'];
```

</details>

## Getting Started

### System Requirements

Since this stack involves heavy infrastructure (Elasticsearch, Kafka), please ensure your environment meets the minimum requirements:

* **RAM:** 4GB minimum free memory (8GB recommended).
* **Disk:** 10GB free space.
* **Note:** If you are running on low memory, consider disabling Kibana in `docker-compose.yml` to save resources.

### Prerequisites

* **Docker** & **Docker Compose** installed.
* **Make** (Optional, for simplified commands).

### Quick Start (Recommended)

We provide a `Makefile` to simplify common operations.

1. **Clone the repository**

   ```bash
   git clone https://github.com/Yupoer/logpulse.git
   cd logpulse
   ```

2. **Run the application**

   ```bash
   make run
   ```

   This command will automatically build the images and start all services (App, MySQL, Redis, Kafka, ES, Kibana) in the background.
   
   > **Note:** By default, `make run` initializes **3 Go Application Replicas** (API + Worker) and **3 Kafka partitions/consumers** behind an **Nginx Load Balancer** to simulate a production-ready distributed environment.

3. **Stop the application**

   ```bash
   make stop
   ```

### Manual Start (Without Make)

If you are on Windows (without WSL) or don't have `make` installed, you can use the raw Docker commands:

```bash
# Start services
docker-compose -f deployments/docker-compose.yml up -d --build

# Stop services
docker-compose -f deployments/docker-compose.yml down
```

### Verify Status

```bash
docker-compose ps
# OR if you configured it in Makefile:
# make ps
```


### Troubleshooting

If you encounter `bind: address already in use` or Windows WinNAT port issues:

1. Open the `.env` file in the root directory.
2. Change the conflicting port (e.g., change `KIBANA_PORT` from `5601` to `5602`).
3. Run `make run` again.

## API Usage Examples

### Quick Test (VS Code)

We provide an `apiTest.http` file for convenient testing directly within VS Code.

1. Install the **[REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)** extension.
2. Open the `apiTest.http` file in this repository.
3. Click the **Send Request** link that appears above each API call to interact with your running services.

### 1. Ingest a Log (Producer)

Send a log entry to the system. The API will respond immediately (Async).

```bash
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{
    "service": "payment-service",
    "level": "error",
    "message": "Transaction failed due to timeout",
    "timestamp": "2023-12-05T10:00:00Z"
  }'
```

### 2. Search Logs (Consumer & Reader)

Search logs via Elasticsearch.

```bash
curl "http://localhost:8080/logs/search?q=timeout&level=error"
```

## Key Features

*   **High Concurrency Ingestion**: Utilizing Kafka as a buffer to handle traffic spikes and prevent database overload (Peak Shaving).
*   **Full-Text Search**: Integrated Elasticsearch for efficient log indexing and fuzzy search capabilities (CQRS Pattern).
*   **Rate Limiting**: Implemented Redis (Token Bucket / Counter) to protect the API from abuse (DDoS protection).
*   **Clean Architecture**: Codebase structured into Controller, Service, and Repository layers with Dependency Injection, ensuring testability and maintainability.
*   **Fully Containerized**: "One-Click Deployment" for the entire stack (App, DB, Broker, Search) using Docker Compose.
*   **CI/CD Pipeline**: Automated linting, testing, and image building via GitHub Actions.
*   **DevOps Ready**: Implemented Graceful Shutdown and Health Checks for zero-downtime deployments.

## Architecture

The system follows an Event-Driven Architecture:

```mermaid
graph LR
    Client[Client] -->|HTTP POST| Nginx[Nginx Load Balancer]
    Nginx -->|Round Robin| API["Go API Cluster (x3)"]
    API -->|Rate Limit| Redis[(Redis)]
    API -->|Async Push - 3 Partitions| Kafka{Kafka Broker}
    Kafka -->|Batch Pull - 3 Consumers| Worker["Go Worker Cluster (x3)"]
    Worker -->|Dual Write| ES[(Elasticsearch)]
    Worker -->|Persist| MySQL[(MySQL)]
```

```mermaid
graph LR
    subgraph "DevOps Pipeline"
        Git[Push Code] --> GitHub[GitHub Actions]
        GitHub --> Lint[GolangCI-Lint]
        GitHub --> Test[Unit Test]
        GitHub --> Docker[Build & Push Image]
    end
    
    subgraph "Runtime System"
        Client[Client] -->|HTTP| API[Go API]
        API -->|Async| Kafka{Kafka}
        Kafka --> Worker[Go Worker]
        Worker --> ES[(Elasticsearch)]
        Worker --> MySQL[(MySQL)]
    end
```

## Design Decisions & Trade-offs

* **Why Kafka over RabbitMQ?**
    * LogPulse requires high-throughput sequential writing. Kafka's log-based storage offers superior performance for peak shaving (100k+ msg/sec) compared to RabbitMQ's complex routing.
* **Why Elasticsearch?**
    * MySQL performs poorly on fuzzy text search (`LIKE %...%`). ES provides Inverted Indexing, enabling O(1) search complexity for log keywords.
* **Hybrid Data Strategy (The "Write-Async, Read-Aside" Pattern)**
    * **Ingestion (Write):** We use **Asynchronous Write** via Kafka. This ensures the API remains low-latency (<10ms) even if the storage layer is under heavy load.
    * **Retrieval (Read):** We employ the **Cache-Aside Pattern** for specific log retrieval. Data is loaded into Redis only upon request (Lazy Loading), optimizing memory usage by not caching the entire log stream.

## Project Layout

The project follows the Standard Go Project Layout:

```plaintext
.
├── cmd/
│   └── api/
│       └── main.go       # Application entry point
├── configs/
│   └── config.yaml       # Configuration file
├── deployments/
│   └── docker-compose.yml # Infrastructure definition
├── internal/
│   ├── config/           # Configuration loading
│   ├── domain/           # Domain models
│   ├── handler/          # HTTP Handlers (Gin)
│   ├── repository/       # Data Access (MySQL, Redis, ES, Kafka)
│   └── service/          # Business Logic
├── pkg/
│   └── utils/            # Shared utilities
├── nginx/
│   └── nginx.conf        # Nginx Load Balancer Configuration
├── .env                  # Environment variables (if don't have one, 'make run' will auto create one)
├── .golangci.yml         # Linting configuration
├── Dockerfile            # Container definition
├── Makefile              # Management commands
└── README.md
```

## License

Distributed under the MIT License. See LICENSE for more information.