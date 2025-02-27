#!/bin/bash
set -e

echo "post-create スクリプトを実行中..."

# Goのツールをインストール
echo "Go ツールをインストール中..."
go install golang.org/x/tools/gopls@latest || echo "警告: goplsのインストールに失敗しました"
go install github.com/go-delve/delve/cmd/dlv@latest || echo "警告: delveのインストールに失敗しました"
go install honnef.co/go/tools/cmd/staticcheck@latest || echo "警告: staticcheckのインストールに失敗しました"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || echo "警告: golangci-lintのインストールに失敗しました"

# プロジェクトの依存関係をインストール
echo "プロジェクトの依存関係をインストール中..."
go mod tidy
go get github.com/chai2010/webp
go get github.com/jdeng/goheif
go get github.com/kolesa-team/go-webp/encoder
go get github.com/kolesa-team/go-webp/webp
go get github.com/Kagami/go-avif
go get gopkg.in/yaml.v3

# textlintの設定ファイルを作成
echo "textlint の設定ファイルを作成中..."
cat > .textlintrc <<EOF
{
  "filters": {},
  "rules": {
    "preset-ja-technical-writing": {
      "no-exclamation-question-mark": false,
      "max-comma": false
    },
    "preset-japanese": true,
    "spellcheck-tech-word": true
  }
}
EOF

# init.vimの設定
echo "vim の設定を作成中..."
mkdir -p ~/.vim
cat > ~/.vim/init.vim <<EOF
set encoding=utf-8
set fileencodings=utf-8,sjis,euc-jp,latin
set fileformats=unix,dos,mac
set ambiwidth=double
set expandtab
set tabstop=4
set softtabstop=4
set shiftwidth=4
set autoindent
set smartindent
set number
set ruler
syntax enable
EOF

echo "post-create スクリプトが完了しました！"
