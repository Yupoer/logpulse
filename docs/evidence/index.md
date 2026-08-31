# LogPulse evidence index

本索引把原始 dirty 狀態、這次實作和 runtime 證據分開。任何沒有實際命令輸出的項目，
仍維持「已實作、尚無 runtime output」；本輪 GitHub Actions CD 已有同一 run
的 runtime output。

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
| Baseline、detached worktree `codex/logpulse-eks-lab` | Done with boundary | dirty Kubernetes 與 `.claude/` 已保存；隔離 clone 已 review 並合併到 `master`，原始 checkout 保留。 |
| Go、metrics、consumer shutdown、Compose、Kubernetes、monitoring | Done | 程式、manifest、local Compose 與 AWS lab runtime 均有證據。 |
| Terraform AWS IaC | Done with lab variant | 設定預設 `t3.large` 單一 node；為符合帳號限制，實際 lab 使用 `t3.small` x 2。 |
| GitHub Actions CI/CD 設定 | Runtime verified in lab | OIDC、`workflow_run.head_sha`、immutable tag、rollout、smoke、`EKS_DEPLOY_ENABLED` 已寫入並由 master workflow 執行。 |
| GitHub branch push、PR、workflow run | Done with boundary | PR #1–#4 已合併；master CI run `33356543223` 與 CD run `33356650844` PASS，細節見 after evidence。 |
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
| CI/CD SHA image、OIDC、rollout、smoke test | `.github/workflows/` | CI and CD runtime verified in lab |
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
| GitHub push/PR/workflow | CI and CD PASS (lab) | PR #1–#4 已合併；master CI `33356543223`、CD `33356650844` 的 OIDC、ECR、rollout、smoke output 見 after evidence |
| Playwright browser check | PASS | CLI daemon 在受限環境 crash；改用同一 Playwright runtime + installed Chrome 完成檢查 |

## Claim boundary

本階段可以說「程式與設定已實作，AWS lab runtime 已驗證，且 GitHub Actions CD 在該 lab
完成 OIDC、ECR、EKS rollout 與 `/ping` smoke test；lab 資源已清理」。不能說 production
SLA、零遺失或正式環境已部署。KMS key 的刪除仍受 AWS pending deletion 流程管理。

## Claim Audit（2026-08-31）

下表是 current-facing claim 的固定判定；`before/` 僅保存歷史 baseline，不作為目前成果。

| Claim | Status | Evidence path | Environment / limit |
| --- | --- | --- | --- |
| AWS EKS | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md` | `ap-northeast-1` lab；已 destroy，不是 production cluster |
| Terraform | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md`；LogPulse `infra/terraform/aws/` | validate／apply／destroy 僅在 lab；無持續運行環境 |
| Kubernetes | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md`；LogPulse `k8s/` | EKS lab manifests／runtime；不代表 production platform |
| HPA | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md`；LogPulse `k8s/hpa.yaml` | 配置 2–6；只觀察到 metrics 與當時 2 replicas，未宣稱到達 6 |
| Readiness／liveness probes | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md`；LogPulse `k8s/app.yaml` | lab rollout／pod health；不是 production SLO |
| Rollout／rollback | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md` | EKS lab 兩副本觀察；不宣稱 zero downtime |
| Graceful shutdown | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md`；`docs/evidence/after/implementation-status.md` | SIGTERM／`--previous` log 來自 lab pod；不等於 zero data loss |
| Prometheus | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md`；LogPulse `monitoring/prometheus.yml` | lab scrape `up=1`；storage 為 ephemeral |
| Grafana | `LAB_ONLY` | `docs/evidence/after/aws-runtime.md`；LogPulse `monitoring/grafana/provisioning/` | lab health／dashboard API；非長期 observability 平台 |
| GitHub Actions CI | `VERIFIED` | PR run `33137078466`；master run `33297971228` | GitHub-hosted `quality`／`build` PASS；只證明 CI gate |
| GitHub Actions CD | `VERIFIED_LAB_ONLY` | `docs/evidence/after/aws-runtime.md`；run `33356650844` | OIDC、ECR immutable SHA image、EKS rollout、`/ping` smoke PASS；驗證後 gate=false，非 production |
| BuildKit | `VERIFIED` | `docs/evidence/after/implementation-status.md`；`docs/evidence/index.md` | Docker Desktop controlled build；cold／warm 結果不代表 production throughput |
| Kafka | `VERIFIED` | `docs/evidence/after/aws-runtime.md`；`docs/evidence/index.md` | controlled local／lab integration smoke；不宣稱 production durability |
| Redis Token Bucket | `VERIFIED` | `docs/evidence/index.md`；LogPulse `internal/middleware/ratelimit.go` | local／lab shared rate-limit evidence；無 production SLA |
| MySQL | `VERIFIED` | `docs/evidence/after/aws-runtime.md`；`docs/evidence/index.md` | lab API／consumer integration；非 managed production database |
| Elasticsearch | `VERIFIED` | `docs/evidence/after/aws-runtime.md`；`docs/evidence/index.md` | lab write／search smoke；不宣稱 production durability |
| k6 | `VERIFIED` | `docs/evidence/after/aws-runtime.md`；LogPulse `test/k6/eks_benchmark.js` | controlled local／AWS lab runs；數字不是 production capacity 或 SLA |
