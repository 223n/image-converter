#!/bin/bash
set -e

# カスタムビルドスクリプト
# go-webpライブラリに依存せずにビルド

echo "カスタムビルドを実行中..."

# 環境変数を設定して特定のライブラリを無視
export GOPROXY=direct
export GOSUMDB=off

# ビルド実行
echo "ビルド実行中..."
go build -tags 'nocgo' -o image-converter .

echo "ビルド完了"
