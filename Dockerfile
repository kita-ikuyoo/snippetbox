# ==========================================
# Stage 1: ビルド環境 (Builder)
# ==========================================
FROM golang:1.26-alpine AS builder

WORKDIR /app

# go.mod と go.sum を先にコピーして依存パッケージをダウンロード
# （ソースコードが変更されても、モジュールに変更がなければキャッシュが効いてビルドが高速になります）
COPY go.mod go.sum ./
RUN go mod download

# プロジェクトの全ファイルをコピー
COPY . .
ARG TARGETOS
ARG TARGETARCH
# Goアプリケーションのビルド
# CGO_ENABLED=0 にすることで、実行環境に依存しない完全な静的バイナリを作成します
#
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o snippetbox ./cmd/web

RUN go run /usr/local/go/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost
# ==========================================
# Stage 2: 実行環境 (Runner)
# ==========================================
# ビルド用の重いツールを含まない、超軽量なOSイメージを使用
FROM alpine:latest

WORKDIR /app

# Stage 1（builder）で作成した実行可能ファイル（バイナリ）だけをコピー
COPY --from=builder /app/snippetbox .

# アプリケーションの実行に必要な静的ファイルや証明書をコピー
COPY --from=builder /app/tls ./tls

RUN apk add --no-cache tzdata
ENV TZ=Asia/Tokyo
# コンテナが使用するポートを明示
EXPOSE 443


# コンテナ起動時の実行コマンド
CMD ["./snippetbox"]