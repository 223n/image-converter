#!/bin/bash
set -e

# 依存関係インストールスクリプト

echo "依存関係のインストールを開始します..."

# OS種別を検出
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS_TYPE=$ID
elif [ -f /etc/debian_version ]; then
    OS_TYPE="debian"
elif [ -f /etc/redhat-release ]; then
    OS_TYPE="rhel"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS_TYPE="macos"
else
    OS_TYPE="unknown"
fi

# Goの依存関係をインストール
echo "Goの依存パッケージをインストール中..."
go mod download
go mod tidy

# システム依存パッケージをインストール
echo "システム依存パッケージをインストール中..."
case $OS_TYPE in
    "debian"|"ubuntu")
        echo "Debian/Ubuntu系OSを検出しました"
        sudo apt-get update
        sudo apt-get install -y webp libaom-dev build-essential libwebp-dev
        ;;
    "rhel"|"fedora"|"centos")
        echo "RHEL/CentOS/Fedora系OSを検出しました"
        sudo dnf install -y libwebp-tools libaom-devel gcc make
        ;;
    "macos")
        echo "macOSを検出しました"
        brew install webp aom
        ;;
    *)
        echo "不明なOS: $OS_TYPE"
        echo "手動でWebPとlibaom開発パッケージをインストールしてください"
        ;;
esac

# 依存コマンドの確認
echo "必要なコマンドの存在を確認中..."
MISSING_COMMANDS=()

# WebP変換ツールの確認
if ! command -v cwebp &> /dev/null; then
    MISSING_COMMANDS+=("cwebp")
fi

if [ ${#MISSING_COMMANDS[@]} -ne 0 ]; then
    echo "警告: 以下のコマンドがインストールされていません:"
    for cmd in "${MISSING_COMMANDS[@]}"; do
        echo "  - $cmd"
    done
    echo "一部の機能が正常に動作しない可能性があります"
else
    echo "すべての必要なコマンドがインストールされています"
fi

echo "依存関係のインストールが完了しました"
