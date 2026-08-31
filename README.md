# LogPulse

LogPulse 是一個 Go log ingestion lab：API 將 log 放進 Kafka，consumer 再寫入
MySQL 與 Elasticsearch；Redis 提供 rate limit 與簡單快取。每個 app instance
同時執行 HTTP API 和 Kafka consumer，不再依賴 Nginx。

## 本機啟動

需求：Go、Docker Desktop（Compose v2）。

```powershell
Copy-Item .env.example .env
docker compose -f deployments/docker-compose.yml config
docker compose -f deployments/docker-compose.yml up -d --build
curl http://localhost:8080/ping
curl http://localhost:8080/metrics
docker compose -f deployments/docker-compose.yml down
```

Compose 只用於本機與 integration lab，包含 MySQL、Redis、Kafka、
Elasticsearch、Kibana、Prometheus 和 Grafana。資料 volume 是本機用途，
不是 AWS production storage 設計。

## API

| Method | Path | 用途 |
| --- | --- | --- |
| GET | `/ping` | process health probe |
| GET | `/metrics` | Prometheus text metrics |
| POST | `/logs` | 將 log 發到 Kafka |
| GET | `/logs/:id` | 先查 Redis，再查 MySQL |
| GET | `/logs/search?q=...` | Elasticsearch search |

`/ping` 與 `/metrics` 不經 Redis rate limiter，避免 Kubernetes probe 或
Prometheus scrape 因流量而得到 429。metrics 只使用 method、route template、
status 和固定 processing result 等低 cardinality labels。

## Kubernetes lab

先確認 cluster 與 kubectl context，再依序套用：

```powershell
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/config-secret.yaml
kubectl apply -f k8s/backends.yaml
kubectl apply -f k8s/app.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/monitoring.yaml
kubectl -n logpulse get pods,svc,hpa
kubectl -n logpulse port-forward svc/logpulse 8080:80
```

`k8s/app.yaml` 使用兩個 app replicas、ClusterIP、readiness/liveness、
resource requests/limits、rolling update、rollback history 和 30 秒
termination grace period。`k8s/backends.yaml` 的 stateful dependencies 是
單副本與 `emptyDir`，只適合 lab；AWS 階段不會把它們誤稱為 managed service。
HPA 需要 Metrics Server。AWS 版本透過 Terraform EKS add-on 設定；本 repo 已
用 AWS lab 驗證節點、metrics API 與 HPA。

Rollback 範例：

```powershell
kubectl -n logpulse rollout undo deployment/logpulse-app
kubectl -n logpulse rollout status deployment/logpulse-app
```

## AWS Terraform lab

`infra/terraform/aws/` 只描述本計畫需要的 VPC、兩個 public subnets、IGW、
ECR、EKS、managed node group、基本 EKS add-ons、GitHub OIDC deploy role、EKS
access entry 與 budget alert。刻意不建立 RDS、MSK、ElastiCache、
OpenSearch、NAT Gateway、Ingress 或 service mesh。

請先複製 `infra/terraform/aws/terraform.tfvars.example`，再依該目錄 README
執行 `terraform init`、`validate`、`plan`、`apply`。Free Tier lab 可在未提交的
`terraform.tfvars` 使用 `t3.small` 與 `node_count = 2`；完成驗證後執行
`terraform destroy`。

## CI/CD

- CI：download dependencies、lint、test、vet、build。
- CD：只在 `vars.EKS_DEPLOY_ENABLED == 'true'` 時執行，使用 GitHub OIDC，
  image tag 固定為 `workflow_run.head_sha`，推送 ECR 後更新 EKS，等待 rollout
  並在 cluster 內執行 `/ping` smoke test。

本次 lab runtime 驗證期間已開啟 gate 並補齊 repository variables/secret：master
CI run `33356543223` PASS，CD run `33356650844` PASS。CD 透過 OIDC assume-role、
以測試 commit `2f84e5ae74278ba86e39b1b27c9386037de5420d` 推送 ECR immutable image，
完成 EKS rollout 與 cluster `/ping` smoke test；image digest 為
`sha256:ca07ecec66d584ab8bb8e0ac1561dc44742d55eae24a85207855cd2b04295aca`。
驗證後 gate 已關閉，之後的文件變更不會再次部署 lab。

## 文件與證據

- [入口頁](docs/index.html)
- [ELI5 差異教學](docs/learning/LogPulse_補強前後差異教學.html)
- [20 分鐘面試講稿](docs/interview/LogPulse_20分鐘面試講稿.html)
- [Evidence index](docs/evidence/index.md)
- [原始 dirty baseline](docs/evidence/before/)

AWS lab runtime、ECR image、EKS、Prometheus/Grafana、rollout/rollback、graceful
shutdown、k6 與 GitHub Actions CD runtime 證據見
[aws-runtime.md](docs/evidence/after/aws-runtime.md)。所有 lab 結果都不等於 production
SLA。
