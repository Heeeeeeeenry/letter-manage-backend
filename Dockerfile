# ============================================================
# letter-manage-backend — 企业级多阶段构建
# 包含 LibreOffice（xlsx→pdf） + 中文字体
# ============================================================

# ── Stage 1: 构建 Go 二进制 ─────────────────────────────────
FROM golang:1.24-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/server .

# ── Stage 2: 运行时镜像 ─────────────────────────────────────
FROM debian:bookworm-slim

# 安装 LibreOffice + 中文字体 + 健康检查依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    libreoffice-core \
    libreoffice-writer \
    libreoffice-calc \
    libreoffice-impress \
    fonts-wqy-zenhei \
    fonts-wqy-microhei \
    gettext-base \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/* \
    && fc-cache -fv

# 创建非 root 用户（uid=1000，与 people-backend 一致）
RUN groupadd -g 1000 appuser && useradd -u 1000 -g appuser -d /app appuser

COPY --from=builder /app/server /app/server
COPY --from=builder /src/config.yaml.template /app/config.yaml.template
COPY --from=builder /src/templates /app/templates
COPY --from=builder /src/scripts /app/scripts
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# media 目录（持久化存储挂载点）
RUN mkdir -p /app/media/letters && chown -R appuser:appuser /app

WORKDIR /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["/app/docker-entrypoint.sh"]
