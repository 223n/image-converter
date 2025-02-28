#!/bin/bash
set -e

# 実行スクリプト
# プログラムを様々なモードで実行するためのラッパースクリプト

# デフォルト設定
CONFIG_FILE="configs/config.yml"
DRY_RUN=false
REMOTE=false

# コマンドライン引数の解析
for arg in "$@"; do
    case $arg in
        --config=*)
            CONFIG_FILE="${arg#*=}"
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --remote)
            REMOTE=true
            shift
            ;;
        --help)
            echo "使用方法: $0 [オプション]"
            echo "オプション:"
            echo "  --config=FILE     使用する設定ファイルを指定 (デフォルト: configs/config.yml)"
            echo "  --dry-run         ドライランモードで実行（実際の変換を行わない）"
            echo "  --remote          リモートモードで実行（SSHで接続して変換）"
            echo "  --help            このヘルプメッセージを表示"
            exit 0
            ;;
        *)
            # 不明なオプション
            echo "警告: 不明なオプション: $arg"
            ;;
    esac
done

# バイナリが存在するか確認
if [ ! -f "image-converter" ]; then
    echo "バイナリが見つかりません。ビルドします..."
    sh scripts/build.sh
fi

# コマンドライン引数の構築
CMD_ARGS=("-config=$CONFIG_FILE")

if [ "$DRY_RUN" = true ]; then
    CMD_ARGS+=("-dry-run")
fi

if [ "$REMOTE" = true ]; then
    CMD_ARGS+=("-remote")
fi

# 実行
echo "実行中: image-converter ${CMD_ARGS[*]}"
./image-converter "${CMD_ARGS[@]}"
