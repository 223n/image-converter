#!/bin/bash
set -e

# カスタムビルドスクリプト

echo "ビルドを実行中..."

# 環境変数を設定
export GOPROXY=direct
export GOSUMDB=off
export CGO_ENABLED=1

# バージョン情報
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date +%FT%T%z)

# ビルドフラグ
LDFLAGS="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}"

# ビルドタイプによるフラグの設定
if [ "$1" = "nocgo" ]; then
    echo "CGOを無効にしてビルドします..."
    export CGO_ENABLED=0
    BUILD_FLAGS="-tags 'nocgo'"
else
    BUILD_FLAGS=""
fi

# ビルド実行
echo "ビルド実行中..."
go build ${BUILD_FLAGS} -ldflags="${LDFLAGS}" -o image-converter cmd/image-converter/main.go

echo "ビルド完了: image-converter"
