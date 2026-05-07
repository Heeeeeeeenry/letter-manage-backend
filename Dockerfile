# ============================================
# letter-manage-backend 完整 Docker 部署
# 包含 LibreOffice 用于 xlsx→pdf 转换
# ============================================

FROM golang:1.24-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# ============================================
FROM debian:bookworm-slim

# 安装 LibreOffice + 中文字体
RUN apt-get update && apt-get install -y \
    libreoffice-core \
    libreoffice-writer \
    libreoffice-calc \
    libreoffice-impress \
    fonts-wqy-zenhei \
    fonts-wqy-microhei \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server /app/server
COPY --from=builder /app/config.yaml /app/config.yaml

WORKDIR /app

EXPOSE 8080

CMD ["/app/server"]
