#!/bin/bash
set -e

# クリーンビルドスクリプト
# キャッシュや古いバイナリを削除してからビルドします

echo "クリーンビルドを実行中..."

# ビルドキャッシュの削除
echo "ビルドキャッシュを削除中..."
go clean -cache

# 既存のバイナリを削除
echo "既存のバイナリを削除中..."
rm -f image-converter

# ビルド実行
echo "ビルド実行中..."
sh scripts/build.sh "$@"

echo "クリーンビルド完了"
