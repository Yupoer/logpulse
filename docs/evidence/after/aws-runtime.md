# AWS runtime evidence

日期：2026-08-28（AWS lab）；2026-08-31（GitHub Actions CD runtime，Asia/Taipei）

## Infrastructure

- Region：`ap-northeast-1`。
- Terraform `validate`：PASS；apply 後的空變更 `plan`：`No changes`。
- EKS cluster：`logpulse-eks`；managed node group：`t3.small` x 2，兩台 node 均 `Ready`。
- Add-ons：`coredns`、`kube-proxy`、`vpc-cni`、`metrics-server`。
- ECR runtime image：`local-retry-20260828094000`，digest
  `sha256:b43cd99f3cf0aa753800ca21edb8d5735631ac6b7bc11e0abe744909f97e93b3`。
- GitHub deploy role trust policy 已限制到 `Yupoer/logpulse` 的 `master` branch。
- 建立項目只有 Terraform 宣告的 VPC、ECR、EKS、IAM/OIDC、access entry、add-ons
  與 budget；沒有建立 RDS、MSK、ElastiCache、OpenSearch、NAT Gateway、Ingress
  或 service mesh。

## Kubernetes runtime

```text
kubectl apply --dry-run=server ...                 PASS (22 resources)
kubectl get nodes                                     2 Ready
kubectl get pods -n logpulse                         app 2/2 Ready; backends and monitoring Running
kubectl get apiservice v1beta1.metrics.k8s.io         True
kubectl top nodes                                    PASS
kubectl get hpa -n logpulse                           cpu: 2%/60%, replicas 2
```

後端在 t3.small lab 加上 requests/limits，Kafka 關閉不需要的 log cleaner 並限制 heap；
這是為了避免小型節點被重量服務壓垮，不改 API 或資料庫 schema。

## API / monitoring smoke

透過短暫 `kubectl port-forward`：

```text
GET  /ping                         200 {"message":"pong"}
GET  /metrics                      200
POST /logs                         201
GET  /logs/search?q=aws-smoke      200, count=1
Prometheus /-/ready                200, "Prometheus Server is Ready."
Prometheus up{job="logpulse"}     1
Grafana /api/health                200, database=ok
Grafana dashboard search           LogPulse overview
```

app log 同時出現 `Message sent to partition ...` 與 `[Worker] Bulk Indexed 1 logs to ES`。

## Delivery behaviour

- rolling update：PASS。
- rollback：PASS，兩個 app replicas 回到 `Running`。
- graceful shutdown：對單一副本送 `SIGTERM` 後，`--previous` log 出現
  `Shutting down server...`、`Server exiting`，另一副本持續服務。
- k6 smoke：20 iterations、HTTP failure `0%`、p95 `151.68ms`，thresholds PASS。

## GitHub Actions CD runtime

同一個 master 交付鏈的 runtime 證據：

- CI run [`33356543223`](https://github.com/Yupoer/logpulse/actions/runs/33356543223)：quality/build PASS。
- CD run [`33356650844`](https://github.com/Yupoer/logpulse/actions/runs/33356650844)：deploy job PASS；測試 commit/image tag 為
  `2f84e5ae74278ba86e39b1b27c9386037de5420d`。
- Configure AWS credentials 使用 GitHub OIDC，caller check 輸出
  `AWS OIDC assume-role: PASS`。
- ECR push 完成；該 tag digest 為
  `sha256:ca07ecec66d584ab8bb8e0ac1561dc44742d55eae24a85207855cd2b04295aca`。
- EKS `logpulse-app` rollout 輸出 `deployment "logpulse-app" successfully rolled out`；
  deployment `READY=2 UPDATED=2 AVAILABLE=2`。
- cluster smoke test 以 curl pod 呼叫 `/ping`，回應 `{"message":"pong"}`；該 smoke pod
  隨後刪除。
- runtime 收尾的 app pods 為 `2/2 Running/Ready`、restart `0`；MySQL、Redis、
  Zookeeper、Kafka、Elasticsearch backend pods 亦為 Running/Ready、restart `0`。
- 驗證完成後將 repository variable `EKS_DEPLOY_ENABLED` 設為 `false`。

## Cleanup

- `terraform destroy -auto-approve` 完成，輸出 `Destroy complete! Resources: 61 destroyed.`；
  `terraform state list` 為空。
- `aws eks list-clusters`、`logpulse` ECR repository、Project=logpulse 的 VPC、Budget、
  `logpulse-github-deploy` role 與 GitHub Actions OIDC provider 查詢均為空。
- `aws logout --profile logpulse` 已完成，AWS CLI 快取登入憑證已清除。
- EKS module 的 KMS key 依 AWS 的刪除流程仍是 `PendingDeletion`，AWS 回報刪除日期
  `2026-09-07`；刪除由 AWS 非同步處理，因此不能宣稱 AWS 帳號內所有資源已在同一秒消失。

## Claim boundary

這些是本次 AWS lab 與 GitHub Actions CD 的 runtime 證據，不是 production SLA 或零遺失
保證，也不是 production deployment。CD run `33356650844` 已完成 OIDC、ECR、EKS rollout
與 `/ping` smoke；lab 資源已清理，KMS 仍依 AWS pending deletion 流程處理。
