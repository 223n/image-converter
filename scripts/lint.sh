#!/bin/bash
set -e

# Lintチェックスクリプト
# Reviveを使用してコードの品質チェックを行います

# Reviveがインストールされているかチェック
if ! command -v revive &> /dev/null; then
    echo "Reviveがインストールされていません。インストールします..."
    go install github.com/mgechev/revive@latest
fi

echo "Reviveによるlintチェックを実行中..."

# プロジェクト内の全てのGoファイルに対してReviveを実行
REVIVE_RESULT=$(revive -config revive.toml ./... 2>&1)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "Lintチェック成功: 問題は見つかりませんでした"
else
    echo "Lintチェックで問題が見つかりました:"
    echo "$REVIVE_RESULT"
    
    # 修正可能な問題を自動修正するか尋ねる
    read -p "自動修正可能な問題を修正しますか？ (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "gofmtによる自動フォーマットを実行中..."
        go fmt ./...
        echo "自動フォーマット完了"
    fi
fi

exit $EXIT_CODE
