# First-stage implementation status

日期：2026-08-28（Asia/Taipei）

## Done

- Go app 以 signal context 管理 HTTP server 與 Kafka consumer；consumer 在取消時
  flush batch、關閉 client，main 等待 consumer 完成。
- `/metrics` 提供 HTTP request counter、status、duration summary count/sum，及
  Kafka processing result；`/ping`、`/metrics` 跳過 rate limiter。
- Compose 直接暴露 app 8080，移除 Nginx，加入可選的 local Prometheus/Grafana。
- Kubernetes 以 `logpulse` namespace 分隔，app 有 probes、resource limits、rolling
  update、termination grace、HPA；monitoring 使用 `emptyDir`。
- Terraform 描述最小 AWS lab；CI/CD 描述 SHA image、OIDC、ECR/EKS rollout 與 smoke。
- 入口頁、ELI5 差異教學、20 分鐘面試稿和 evidence index 已建立。

## Verified locally

```text
go test ./...                                  PASS
go vet ./...                                   PASS
go test ./internal/metrics ./internal/middleware PASS
gofmt                                         PASS
git diff --check                              PASS
go build -trimpath ./cmd/api                  PASS
docker compose ... config                     PASS
Docker BuildKit cold/warm                    PASS (115.43s / 3.24s)
HTML browser smoke                         PASS (`output/playwright/logpulse-index.png`)
```

Go build cache 使用 `.tmp/gocache`，避免寫入受限制的 user cache。

## AWS runtime verified

AWS lab 已完成 Terraform apply、ECR image push、EKS deploy、Prometheus/Grafana
scrape、rollout/rollback、graceful shutdown 與 k6 smoke；細節見
`docs/evidence/after/aws-runtime.md`。為符合帳號 Free Tier，runtime 使用 `t3.small`
兩台，並在 Kubernetes 後端加入簡單 requests/limits 與小型 heap 設定。

驗證後已執行 `terraform destroy`，Terraform state、EKS、ECR、VPC 與 Budget 均清空；
EKS module 的 KMS key 仍依 AWS 流程處於 `PendingDeletion`，不把它寫成已立即刪除。

GitHub CLI authentication 已恢復，branch 已推送並建立 PR #1；CI run
`33137078466` 的 lint、test、vet、build 全部 PASS。CD 的 ECR→EKS workflow 尚未執行，
不把它寫成已部署。
