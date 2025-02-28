# 画像変換ツール（Image Converter）

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Go Version](https://img.shields.io/badge/go-1.16%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)

高速で効率的なWebPおよびAVIF画像変換ツール。ローカル変換とリモートサーバー変換の両方をサポートしています。

## 機能

- **複数フォーマット対応**: JPG、PNG、HEIC、HEIFを最新のフォーマットに変換
- **最適化フォーマット**:
  - **WebP**: ウェブ用に最適化された画像フォーマット
  - **AVIF**: さらに高圧縮かつ高品質な次世代フォーマット
- **柔軟な動作モード**:
  - **ローカルモード**: ローカルディレクトリの画像を変換
  - **リモートモード**: SSH経由でリモートサーバーの画像を変換
  - **ドライランモード**: 実際の変換を行わずに対象ファイルを確認
- **高度な機能**:
  - **並列処理**: マルチスレッドによる高速変換
  - **ディレクトリ再帰**: サブディレクトリも含めた変換
  - **元構造維持**: 元のディレクトリ構造を保持
  - **サーバー機能**: FTPおよびSSHサーバー機能内蔵
- **詳細な進捗表示**: リアルタイムの変換進捗とエラー報告
- **カスタマイズ可能**: 豊富な設定オプション

## スクリーンショット

```bash
変換処理: [████████████████████████░░░░░░░░] 75% (300/400) 経過: 00:01:15 残り: 00:00:25
```

## インストール

### 前提条件（実行）

- Go 1.16以上
- libaom（AVIF変換用）
- WebP変換ツール

### バイナリからのインストール

最新のリリースバイナリは[Releases](https://github.com/yourusername/image-converter/releases)からダウンロードできます。

### ソースからビルド

```bash
# リポジトリのクローン
git clone https://github.com/yourusername/image-converter.git
cd image-converter

# 依存関係のインストール
make install-deps

# Debian/Ubuntu系
make install-system-deps-debian

# RHEL/CentOS/Fedora系
make install-system-deps-redhat

# ビルド
make build-safe

# インストール（オプション）
sudo make install
```

### Docker

```bash
# Dockerイメージのビルド
make docker

# Dockerでの実行
make docker-run
```

## 使用方法

### 基本的な使い方

```bash
# ローカルディレクトリの画像を変換
./bin/image-converter -config=configs/config.yml

# ドライランモード（実際の変換を行わず変換対象のみ表示）
./bin/image-converter -dry-run

# リモートサーバーの画像を変換
./bin/image-converter -remote
```

### 設定ファイル

設定はYAML形式で記述します。主な設定項目:

```yaml
# リモート設定
remote:
  enabled: false
  host: "example.com"
  port: 22
  user: "webuser"
  remote_path: "/var/www/html/images"

# 入力設定
input:
  directory: "./images"
  supported_extensions: [.jpg, .jpeg, .png, .heic, .heif]

# 変換設定
conversion:
  workers: 4
  webp:
    enabled: true
    quality: 80
  avif:
    enabled: true
    quality: 40
    speed: 6
```

詳細な設定オプションは[設定ガイド](docs/CONFIG.md)を参照してください。

## プロジェクト構造

```bash
image-converter/
├── cmd/                    # コマンドラインアプリケーション
│   └── image-converter/    # メインエントリーポイント
├── internal/               # 内部パッケージ
│   ├── config/             # 設定処理
│   ├── converter/          # 変換ロジック
│   ├── remote/             # リモート処理
│   ├── server/             # サーバー機能
│   └── utils/              # ユーティリティ
├── pkg/                    # 公開パッケージ
│   └── imageutils/         # 画像処理ユーティリティ
├── configs/                # 設定ファイル
├── docs/                   # ドキュメント
├── scripts/                # スクリプト
└── bin/                    # ビルド成果物
```

## ドキュメント

詳細なドキュメントは[docs](docs)ディレクトリを参照してください:

- [インストール手順](docs/INSTALL.md)
- [使用方法](docs/USAGE.md)
- [設定ガイド](docs/CONFIG.md)
- [リモート変換](docs/REMOTE.md)
- [トラブルシューティング](docs/TROUBLESHOOTING.md)

## 開発

### 前提条件（開発）

- Go 1.16以上
- VS Code（推奨）+ Remote Containers拡張機能
- Docker（開発コンテナー用）

### 開発環境のセットアップ

```bash
# プロジェクト構造のセットアップ
make setup-project

# 開発環境のセットアップ
make setup-dev

# Lintの実行
make lint-all

# テストの実行
make test
```

### VS Code Dev Containers

このプロジェクトは、VS Code Dev Containersをサポートしています。VS Codeで「Reopen in Container」を選択すると、必要なすべての依存関係が設定された開発環境が自動的に構築されます。

## 貢献

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチをプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

変更を送信する前に、以下を実行してください:

- `make lint-all` でコードスタイルをチェック
- `make test` でテストを実行

## ロードマップ

- [ ] WebP/AVIF品質の自動最適化
- [ ] 画像比較レポート生成
- [ ] ウェブインターフェイス
- [ ] 変換前後のプレビュー機能
- [ ] バッチ処理のパフォーマンス改善

## ライセンス

MITライセンスの下で公開されています - 詳細は[LICENSE](LICENSE)ファイルを参照してください。

## 謝辞

- [libaom](https://aomedia.googlesource.com/aom/) - AVIF変換
- [libwebp](https://developers.google.com/speed/webp/docs/compressing) - WebP変換
- その他、このプロジェクトに貢献してくださったすべての方々
