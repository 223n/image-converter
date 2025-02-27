# Go 画像変換ツール Makefile

# 変数
BINARY_NAME=image-converter
GO=go
GOFLAGS=-v
CONFIG_FILE=config.yml
GOFMT=gofmt
GOFILES=$(wildcard *.go)

.PHONY: all build clean run run-dry remote remote-dry fmt install-deps help test

# デフォルトターゲット
all: fmt build

# ビルド
build:
	@echo "ビルド中..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# 依存関係の問題を回避した安全なビルド
build-safe:
	@echo "安全なビルドを実行します..."
	@chmod +x ./build.sh
	@./build.sh

# ポータブルなビルド (静的リンク)
build-static:
	@echo "静的リンクでビルド中..."
	CGO_ENABLED=0 $(GO) build -a -ldflags '-extldflags "-static"' -o $(BINARY_NAME) .

# バイナリの削除
clean:
	@echo "クリーン中..."
	$(GO) clean
	rm -f $(BINARY_NAME)

# 実行
run: build-safe
	@echo "実行中..."
	./$(BINARY_NAME) -config=$(CONFIG_FILE)

# ドライランモードで実行
run-dry: build-safe
	@echo "ドライランモードで実行中..."
	./$(BINARY_NAME) -config=$(CONFIG_FILE) -dry-run

# リモートモードで実行
remote: build-safe
	@echo "リモートモードで実行中..."
	./$(BINARY_NAME) -config=$(CONFIG_FILE) -remote

# リモートモードでドライラン実行
remote-dry: build-safe
	@echo "リモートモードでドライラン実行中..."
	./$(BINARY_NAME) -config=$(CONFIG_FILE) -remote -dry-run

# フォーマット
fmt:
	@echo "コードのフォーマット中..."
	$(GOFMT) -w $(GOFILES)

# 依存関係のインストール
install-deps:
	@echo "依存関係のインストール中..."
	$(GO) mod tidy
	sudo apt-get update && sudo apt-get install -y webp libaom-dev libwebp-dev

# システムへのインストール
install: build
	@echo "システムへインストール中..."
	cp $(BINARY_NAME) /usr/local/bin/
	mkdir -p /etc/image-converter
	cp $(CONFIG_FILE) /etc/image-converter/

# Docker イメージのビルド
docker:
	@echo "Docker イメージのビルド中..."
	docker build -t image-converter .

# システム依存関係のインストール (Debian/Ubuntu)
install-system-deps-debian:
	@echo "システム依存関係のインストール中..."
	sudo apt-get update
	sudo apt-get install -y libaom-dev pure-ftpd openssh-server

# システム依存関係のインストール (RHEL/CentOS/Fedora)
install-system-deps-redhat:
	@echo "システム依存関係のインストール中..."
	dnf install -y libaom-devel pure-ftpd openssh-server

# テスト実行
test:
	@echo "テスト実行中..."
	@if [ -f "$(shell go list -f '{{.Dir}}' github.com/user/image-converter)/main_test.go" ]; then \
		go test -v . ; \
	else \
		echo "プロジェクトのテストファイルがありません"; \
	fi

# 依存関係の問題を回避するカスタムテスト
test-safe:
	@echo "安全なテスト実行中（外部ライブラリのテストをスキップ）..."
	@chmod +x ./run-tests.sh
	@./run-tests.sh

# ヘルプ
help:
	@echo "利用可能なコマンド:"
	@echo "  make                 - コードをフォーマットしてビルドする"
	@echo "  make build           - バイナリをビルドする"
	@echo "  make build-safe      - 依存関係の問題を回避して安全にビルドする"
	@echo "  make build-static    - 静的リンクされたバイナリをビルドする"
	@echo "  make clean           - ビルドファイルを削除する"
	@echo "  make run             - ビルドして実行する"
	@echo "  make run-dry         - ドライランモードで実行する"
	@echo "  make remote          - リモートモードで実行する"
	@echo "  make remote-dry      - リモートモードでドライラン実行する"
	@echo "  make fmt             - コードをフォーマットする"
	@echo "  make install-deps    - Go の依存パッケージをインストールする"
	@echo "  make install         - システムにインストールする"
	@echo "  make docker          - Docker イメージをビルドする"
	@echo "  make install-system-deps-debian - Debian系のシステム依存パッケージをインストールする"
	@echo "  make install-system-deps-redhat - RHEL系のシステム依存パッケージをインストールする"
	@echo "  make test            - テストを実行する"
	@echo "  make test-safe       - 依存関係の問題を回避して安全にテストを実行する"
	@echo "  make help            - このヘルプを表示する"
