# ==========================================
# Stage 1: Builder (編譯層)
# 使用包含 Go 編譯器的完整映像檔
# ==========================================
FROM golang:1.23-alpine AS builder

# 設定工作目錄
WORKDIR /app

# 1. 下載依賴 (利用 Docker Layer Caching)
# 先 Copy go.mod 和 go.sum，如果這兩個檔沒變，Docker 會直接用快取，不重新下載
COPY go.mod go.sum ./
RUN go mod download

# 2. 複製程式碼並編譯
COPY . .
# CGO_ENABLED=0: 關閉 CGO，確保編譯出的是靜態連結的二進位檔 (Static Binary)
# GOOS=linux: 強制編譯為 Linux 版本 (因為 Container 是跑 Linux)
RUN CGO_ENABLED=0 GOOS=linux go build -o logpulse cmd/api/main.go

# ==========================================
# Stage 2: Runner (執行層)
# 使用極簡的 Linux (Alpine)
# ==========================================
FROM alpine:latest

WORKDIR /root/

# 安裝必要的基礎套件 (HTTPS 憑證與時區資料)
RUN apk --no-cache add ca-certificates tzdata

# 設定時區 (Taipei) - 選用，方便你看 Log
ENV TZ=Asia/Taipei

# 從 Builder 層把編譯好的 "logpulse" 執行檔複製過來
# 注意：我們只複製了執行檔，原始碼完全沒複製過來 (安全 + 輕量)
COPY --from=builder /app/logpulse .

# 宣告 Port (僅作文件用途，實際要在 docker-compose map)
EXPOSE 8080

# 啟動應用
CMD ["./logpulse"]