# LogPulse evidence index

本索引把原始 dirty 狀態、這次實作和延後的 runtime 階段分開。任何沒有實際
命令輸出的項目，都維持 `Implemented but not runtime verified`。

## Baseline

| Evidence | 用途 |
| --- | --- |
| [before/git-head.txt](before/git-head.txt) | 實作前 HEAD |
| [before/git-status.txt](before/git-status.txt) | 實作前 dirty 狀態 |
| [before/working-tree.diff](before/working-tree.diff) | 原始未提交 diff |
| [before/k8s-*.yaml](before/) | 原始 Kubernetes 檔案副本 |
| [before/.claude/](before/.claude/) | 原始 `.claude/` 副本 |

這些檔案是 baseline，不算本次功能成果。

## Original plan comparison

| 原始規劃 | 目前狀態 | 說明 |
| --- | --- | --- |
| Baseline、detached worktree `codex/logpulse-eks-lab` | Partial | dirty Kubernetes 與 `.claude/` 已保存；`.git/worktrees` 權限阻擋 worktree，實作暫留目前 checkout。 |
| Go、metrics、consumer shutdown、Compose、Kubernetes、monitoring | Done | 程式、manifest、local Compose 與 AWS lab runtime 均有證據。 |
| Terraform AWS IaC | Done with lab variant | 設定預設 `t3.large` 單一 node；為符合帳號限制，實際 lab 使用 `t3.small` x 2。 |
| GitHub Actions CI/CD 設定 | Config done | OIDC、`workflow_run.head_sha`、immutable tag、rollout、smoke、`EKS_DEPLOY_ENABLED` 已寫入。 |
| GitHub branch push、PR、workflow run | Partial | branch 已推送、PR #1 已建立、CI run `33137078466` PASS；CD 的 ECR→EKS workflow 尚未執行。 |
| AWS deferred runtime | Done and cleaned | credentials、plan/apply、kubeconfig、dependencies/app/monitoring/HPA、runtime evidence、destroy 已完成；沒有建立 bootstrap user，因此沒有可刪除的暫時 user。 |
| Docker BuildKit cold/warm timing | Done | `--no-cache` cold build 115.43s、warm build 3.24s，兩次 exit 0。 |
| HTML、ELI5、面試稿、evidence index | Done | 入口與講稿已同步 AWS lab 的最新 claim boundary。 |

## Implemented

| Claim | Source | Status |
| --- | --- | --- |
| app 同時執行 API 與 consumer，並等待 graceful shutdown | `cmd/api/main.go`, `internal/repository/kafka_consumer.go` | Implemented |
| HTTP/Kafka metrics，低 cardinality labels | `internal/metrics/`, `internal/middleware/ratelimit.go` | Implemented |
| Compose 移除 Nginx，加入 local Prometheus/Grafana | `deployments/docker-compose.yml`, `monitoring/` | Implemented |
| Kubernetes namespace、probes、resources、rollout、HPA | `k8s/` | Implemented |
| AWS VPC/ECR/EKS/OIDC/access entry/budget IaC | `infra/terraform/aws/` | Implemented and lab runtime verified |
| CI/CD SHA image、OIDC、rollout、smoke test | `.github/workflows/` | Implemented, not run |
| ELI5、面試講稿、入口頁 | `docs/*.html`, `docs/learning/`, `docs/interview/` | Implemented |

## Local checks

| Command | Result | Note |
| --- | --- | --- |
| `go test ./...` | PASS | 使用 workspace `.tmp/gocache` |
| `go vet ./...` | PASS | 使用 workspace `.tmp/gocache` |
| `go test ./internal/metrics ./internal/middleware` | PASS | metrics 與 probe bypass focused tests |
| `gofmt` | PASS | Go source formatted |
| `git diff --check` | PASS (current implementation) | current implementation diff 無 whitespace error；`docs/evidence/before/` 保留原始 baseline whitespace，不改寫歷史證據 |
| Kubernetes YAML parse | PASS | Node `yaml` parser讀取 6 manifests、22 documents |
| Grafana dashboard JSON parse | PASS | Node JSON parser |
| `k6 inspect test/k6/eks_benchmark.js` | PASS | scenario/threshold syntax |
| `docker compose ... config/up --build` | PASS | Docker Desktop、BuildKit、Compose app `/ping`/`/metrics`/`/logs` 已驗證；cold/warm timing 見原始規劃對照 |
| HTML browser smoke | PASS | Chrome + Playwright：3 頁導覽、8 章欄位、9 個面試折疊、timer 遞減 |

## Deferred / blocked runtime

| Area | Status | Reason |
| --- | --- | --- |
| Terraform install / validate / plan | PASS | Terraform 1.13.2；validate 與 apply 後空變更 plan 已通過 |
| AWS credentials、EKS、ECR、OIDC | PASS (lab) | runtime evidence 見 `after/aws-runtime.md`；destroy 已完成 |
| AWS cleanup | PASS with KMS pending deletion | Terraform state、EKS、ECR、VPC、Budget 已清空；KMS 依 AWS 流程待刪除 |
| Docker Compose config/build/up | PASS | Docker Desktop、BuildKit 與 ECR image build 已驗證 |
| Kubernetes dry-run/rollout/HPA | PASS | EKS server-side dry-run、2 Ready nodes、rollout/rollback、HPA metrics 已驗證 |
| Prometheus/Grafana scrape | PASS | Prometheus target `up=1`、Grafana health/dashboard API 已驗證 |
| GitHub push/PR/workflow | CI PASS; CD pending | branch/PR 與 CI run 已驗證；CD 仍受 `EKS_DEPLOY_ENABLED` gate 與已清理的 AWS lab 狀態限制 |
| Playwright browser check | PASS | CLI daemon 在受限環境 crash；改用同一 Playwright runtime + installed Chrome 完成檢查 |

## Claim boundary

本階段可以說「程式與設定已實作，且 AWS lab runtime 已驗證；lab 資源已清理」。不能說
production SLA、零遺失、正式環境已部署，或 GitHub CD 的 ECR→EKS rollout 已完成。GitHub
PR 與 CI run 已有證據；KMS key 的刪除仍受 AWS pending deletion 流程管理。
