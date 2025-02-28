#!/bin/bash
set -e

# テスト実行スクリプト
# Go標準のテストフレームワークでテストを実行します

# デフォルト設定
VERBOSE=false
COVERAGE=false
BENCH=false
RACE=false

# コマンドライン引数の解析
for arg in "$@"; do
    case $arg in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --coverage|-c)
            COVERAGE=true
            shift
            ;;
        --bench|-b)
            BENCH=true
            shift
            ;;
        --race|-r)
            RACE=true
            shift
            ;;
        --help|-h)
            echo "使用方法: $0 [オプション]"
            echo "オプション:"
            echo "  --verbose, -v     詳細な出力を表示"
            echo "  --coverage, -c    カバレッジレポートを生成"
            echo "  --bench, -b       ベンチマークを実行"
            echo "  --race, -r        レース検出を有効化"
            echo "  --help, -h        このヘルプメッセージを表示"
            exit 0
            ;;
        *)
            # 不明なオプション
            echo "警告: 不明なオプション: $arg"
            ;;
    esac
done

# テストコマンドの構築
TEST_CMD="go test"

if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ "$COVERAGE" = true ]; then
    TEST_CMD="$TEST_CMD -coverprofile=coverage.out"
fi

if [ "$RACE" = true ]; then
    TEST_CMD="$TEST_CMD -race"
fi

# テスト実行
echo "テストを実行中..."
eval "$TEST_CMD ./..."

# カバレッジレポートの生成
if [ "$COVERAGE" = true ]; then
    echo "カバレッジレポートを生成中..."
    go tool cover -html=coverage.out -o coverage.html
    echo "カバレッジレポートが作成されました: coverage.html"
fi

# ベンチマークの実行
if [ "$BENCH" = true ]; then
    echo "ベンチマークを実行中..."
    go test -bench=. -benchmem ./...
fi

echo "テスト完了"
