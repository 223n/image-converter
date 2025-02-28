# 画像変換ツール用 Makefile

# 変数定義
BINARY_NAME = image-converter
VERSION = 1.0.0
BUILD_TIME = $(shell date +%FT%T%z)
BUILD_DIR = ./bin
GO = go
GOFLAGS = -v
BUILD_FLAGS = -tags 'nocgo'
LDFLAGS = -ldflags="-s -w -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'"

# デフォルトターゲット
.PHONY: all
all: fmt lint-fix build

# ビルド関連ターゲット
.PHONY: build build-safe build-static clean

# 通常ビルド
build:
	@echo "ビルド実行中..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/image-converter
	@echo "ビルド完了: $(BUILD_DIR)/$(BINARY_NAME)"

# 安全なビルド（CGO依存なし）
build-safe:
	@echo "安全なビルドを実行中..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/image-converter
	@echo "安全なビルド完了: $(BUILD_DIR)/$(BINARY_NAME)"

# 静的リンクビルド
build-static:
	@echo "静的ビルドを実行中..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/image-converter
	@echo "静的ビルド完了: $(BUILD_DIR)/$(BINARY_NAME)"

# すべてのプラットフォーム用にビルド
build-all: build-linux build-macos build-windows

# Linux用ビルド
build-linux:
	@echo "Linux用ビルドを実行中..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/linux/$(BINARY_NAME) ./cmd/image-converter
	@echo "Linux用ビルド完了: $(BUILD_DIR)/linux/$(BINARY_NAME)"

# macOS用ビルド
build-macos:
	@echo "macOS用ビルドを実行中..."
	@mkdir -p $(BUILD_DIR)/macos
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/macos/$(BINARY_NAME) ./cmd/image-converter
	@echo "macOS用ビルド完了: $(BUILD_DIR)/macos/$(BINARY_NAME)"

# Windows用ビルド
build-windows:
	@echo "Windows用ビルドを実行中..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe ./cmd/image-converter
	@echo "Windows用ビルド完了: $(BUILD_DIR)/windows/$(BINARY_NAME).exe"

# クリーンアップ
clean:
	@echo "クリーンアップ中..."
	rm -rf $(BUILD_DIR)
	@echo "クリーンアップ完了"

# 実行関連ターゲット
.PHONY: run run-dry remote

# 通常実行
run: build
	@echo "プログラムを実行中..."
	$(BUILD_DIR)/$(BINARY_NAME) -config=configs/config.yml

# ドライランモードで実行
run-dry: build
	@echo "ドライランモードで実行中..."
	$(BUILD_DIR)/$(BINARY_NAME) -config=configs/config.yml -dry-run

# リモートモードで実行
remote: build
	@echo "リモートモードで実行中..."
	$(BUILD_DIR)/$(BINARY_NAME) -config=configs/config.yml -remote

# テスト関連ターゲット
.PHONY: test test-verbose test-coverage

# テスト実行
test:
	@echo "テスト実行中..."
	$(GO) test -v ./...
	@echo "テスト完了"

# 詳細なテスト実行
test-verbose:
	@echo "詳細なテスト実行中..."
	$(GO) test -v ./... -count=1
	@echo "テスト完了"

# テストカバレッジ
test-coverage:
	@echo "テストカバレッジを計測中..."
	$(GO) test -v ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "テストカバレッジの計測完了: coverage.html"

# コード関連ターゲット
.PHONY: fmt mod-tidy

# コードフォーマット
fmt:
	@echo "コードフォーマット実行中..."
	$(GO) fmt ./...
	@echo "コードフォーマット完了"

# Go依存関係の整理
mod-tidy:
	@echo "Go依存関係を整理中..."
	$(GO) mod tidy
	@echo "依存関係の整理完了"

# 依存パッケージインストール
.PHONY: install-deps install-system-deps-debian install-system-deps-redhat

# Go依存パッケージのインストール
install-deps:
	@echo "依存パッケージインストール中..."
	$(GO) mod download
	@echo "依存パッケージインストール完了"

# Debian/Ubuntu系システム依存パッケージ
install-system-deps-debian:
	@echo "Debian/Ubuntu用システム依存パッケージをインストール中..."
	sudo apt-get update
	sudo apt-get install -y libaom-dev webp
	@echo "システム依存パッケージインストール完了"

# RHEL/Fedora系システム依存パッケージ
install-system-deps-redhat:
	@echo "RHEL/Fedora用システム依存パッケージをインストール中..."
	sudo dnf install -y libaom-devel libwebp-tools
	@echo "システム依存パッケージインストール完了"

# インストール関連ターゲット
.PHONY: install uninstall

# システムへのインストール
install: build
	@echo "システムにインストール中..."
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo mkdir -p /etc/image-converter
	sudo install -m 644 configs/config.yml /etc/image-converter/
	@echo "インストール完了: /usr/local/bin/$(BINARY_NAME)"

# システムからのアンインストール
uninstall:
	@echo "システムからアンインストール中..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -rf /etc/image-converter
	@echo "アンインストール完了"

# Docker関連ターゲット
.PHONY: docker docker-run

# Dockerイメージビルド
docker:
	@echo "Dockerイメージをビルド中..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	@echo "Dockerイメージビルド完了"

# Dockerコンテナ実行
docker-run: docker
	@echo "Dockerコンテナを実行中..."
	docker run -v $(PWD)/images:/app/images $(BINARY_NAME):$(VERSION)

# Lint関連ターゲット
.PHONY: install-lint-tools lint lint-verbose lint-fix lint-all lint-ci

# Reviveをインストール
install-lint-tools:
	@echo "Reviveをインストール中..."
	go install github.com/mgechev/revive@latest
	@echo "Reviveインストール完了"

# Lintチェック実行
lint: install-lint-tools
	@echo "Lintチェックを実行中..."
	@revive -config revive.toml ./...

# Lintチェック（詳細出力）
lint-verbose: install-lint-tools
	@echo "詳細なLintチェックを実行中..."
	@revive -config revive.toml -formatter friendly ./...

# 自動的に修正できる問題を修正
lint-fix:
	@echo "コードフォーマットを実行中..."
	@go fmt ./...
	@echo "コードフォーマット完了"

# Lintチェックとフォーマット修正を同時に実行
lint-all: lint-fix lint
	@echo "Lintチェックとフォーマット修正が完了しました"

# CI環境用のLintチェック（エラーがあれば終了）
lint-ci: install-lint-tools
	@echo "CI用Lintチェックを実行中..."
	@revive -config revive.toml -formatter friendly -error ./...

# プロジェクト構成関連ターゲット
.PHONY: setup-project setup-dev update-docs

# プロジェクト構造の初期セットアップ
setup-project:
	@echo "プロジェクト構造を作成中..."
	@mkdir -p cmd/image-converter
	@mkdir -p internal/{config,converter,remote,server,utils}
	@mkdir -p pkg/imageutils
	@mkdir -p configs/config_examples
	@mkdir -p scripts
	@mkdir -p docs
	@mkdir -p $(BUILD_DIR)
	@echo "プロジェクト構造の作成完了"

# 開発環境のセットアップ
setup-dev: install-deps install-lint-tools
	@echo "開発環境のセットアップ完了"

# ドキュメントの更新
update-docs:
	@echo "ドキュメントを更新中..."
	@echo "これはドキュメント生成ツールを使用するターゲットです。実装に応じて更新してください。"

# リリース関連ターゲット
.PHONY: release release-patch release-minor release-major

# リリースビルド
release: clean mod-tidy test lint build-all
	@echo "リリースビルド完了"

# パッチバージョンアップ
release-patch:
	@echo "パッチバージョンアップリリース..."
	# バージョン番号更新ロジックを実装

# マイナーバージョンアップ
release-minor:
	@echo "マイナーバージョンアップリリース..."
	# バージョン番号更新ロジックを実装

# メジャーバージョンアップ
release-major:
	@echo "メジャーバージョンアップリリース..."
	# バージョン番号更新ロジックを実装

# ヘルプ
.PHONY: help
help:
	@echo "画像変換ツール Makefile ヘルプ"
	@echo ""
	@echo "使用可能なターゲット:"
	@echo "  === ビルド関連 ==="
	@echo "  all            - コードフォーマット、Lint修正、ビルドを実行（デフォルト）"
	@echo "  build          - アプリケーションをビルド"
	@echo "  build-safe     - CGO依存なしでビルド"
	@echo "  build-static   - 静的リンクでビルド"
	@echo "  build-all      - 全プラットフォーム用にビルド"
	@echo "  build-linux    - Linux用にビルド"
	@echo "  build-macos    - macOS用にビルド"
	@echo "  build-windows  - Windows用にビルド"
	@echo "  clean          - ビルド成果物を削除"
	@echo ""
	@echo "  === 実行関連 ==="
	@echo "  run            - アプリケーションを実行"
	@echo "  run-dry        - ドライランモードで実行"
	@echo "  remote         - リモートモードで実行"
	@echo ""
	@echo "  === テスト関連 ==="
	@echo "  test           - テスト実行"
	@echo "  test-verbose   - 詳細なテスト実行"
	@echo "  test-coverage  - テストカバレッジを計測"
	@echo ""
	@echo "  === コード関連 ==="
	@echo "  fmt            - コードフォーマット"
	@echo "  mod-tidy       - Go依存関係を整理"
	@echo ""
	@echo "  === 依存関係 ==="
	@echo "  install-deps   - Go依存パッケージをインストール"
	@echo "  install-system-deps-debian - Debian/Ubuntu用システム依存パッケージをインストール"
	@echo "  install-system-deps-redhat - RHEL/Fedora用システム依存パッケージをインストール"
	@echo ""
	@echo "  === インストール ==="
	@echo "  install        - システムにアプリケーションをインストール"
	@echo "  uninstall      - システムからアプリケーションをアンインストール"
	@echo ""
	@echo "  === Docker ==="
	@echo "  docker         - Dockerイメージをビルド"
	@echo "  docker-run     - Dockerコンテナでアプリケーションを実行"
	@echo ""
	@echo "  === Lint ==="
	@echo "  lint           - Reviveによるlintチェック"
	@echo "  lint-verbose   - 詳細なlintチェック"
	@echo "  lint-fix       - 自動修正可能な問題を修正"
	@echo "  lint-all       - lintチェックとフォーマット修正を実行"
	@echo "  lint-ci        - CI環境用lintチェック"
	@echo ""
	@echo "  === プロジェクト構成 ==="
	@echo "  setup-project  - プロジェクト構造を初期セットアップ"
	@echo "  setup-dev      - 開発環境をセットアップ"
	@echo "  update-docs    - ドキュメントを更新"
	@echo ""
	@echo "  === リリース ==="
	@echo "  release        - リリースビルドを作成"
	@echo "  release-patch  - パッチバージョンアップリリース"
	@echo "  release-minor  - マイナーバージョンアップリリース"
	@echo "  release-major  - メジャーバージョンアップリリース"
	@echo ""
	@echo "  help           - このヘルプを表示"
