# 画像変換ツール

このプロジェクトは、指定したディレクトリ内の画像ファイル（JPG、PNG、HEIC、HEIF）をWebPとAVIFフォーマットに変換するGoプログラムです。変換された画像は元のディレクトリに保存されます。また、FTPとSSHによるリモートアクセスもサポートしています。さらに、リモートWebサーバーへのSSH接続による画像変換機能も備えています。

## 機能

- 指定したディレクトリ内の画像を再帰的に検索
- サポートされているフォーマット（JPG、PNG、HEIC、HEIF）を自動検出
- 元のファイル名を保持したまま、WebPとAVIFフォーマットに変換
- 元のディレクトリ構造を維持して出力
- 複数のファイルを並列処理
- FTPサーバーによるリモートアクセス（オプション）
- SSHサーバーによるリモートアクセス（オプション）
- YAML設定ファイルによる柔軟な設定
- リモートWebサーバーへのSSH接続による変換機能

## 開発環境のセットアップ

このプロジェクトはVS Code Dev Containersを使用して開発環境を構築しています。

### 必要なもの

- Visual Studio Code
- Docker
- VS Code Remote - Containers拡張機能

### セットアップ手順

1. リポジトリをクローン
        ```bash
        git clone https://github.com/user/image-converter.git
        cd image-converter
        ```

2. VS Codeでプロジェクトを開く
        ```bash
        code .
        ```

3. VS Codeから「Reopen in Container」を選択
   - 左下の><をクリックし、「Reopen in Container」を選択
   - 初回は開発コンテナーのビルドに時間がかかることがあります

4. ビルドが完了すると、開発環境が自動的に設定されます

## 使用方法

### 設定ファイル

プログラムは`config.yml`ファイルで設定を管理します。主な設定項目は以下の通りです：

- 入力ディレクトリと対象ファイル形式
- 変換オプション（ワーカー数、画質、圧縮率など）
- FTPサーバー設定
- SSHサーバー設定
- ログ設定

設定ファイルのサンプルは[config.yml](./config.yml)を参照してください。

### コマンドライン実行

```bash
# デフォルトの設定ファイル（config.yml）を使用
go run main.go

# カスタム設定ファイルを指定
go run main.go -config=custom-config.yml

# ドライランモード（実際の変換を行わず、対象ファイルのみ表示）
go run main.go -dry-run

# リモートモード（SSH接続して外部サーバーの画像を変換）
go run main.go -remote

# リモートモードとドライランの組み合わせ
go run main.go -remote -dry-run
```

## ビルド方法

バイナリをビルドするには以下のコマンドを実行します：

```bash
go build -o image-converter main.go
```

ビルド後は以下のように実行できます：

```bash
./image-converter -config=config.yml
```

## 設定ファイルの主な項目

```yaml
# リモートサーバー設定
remote:
  enabled: false  # リモート変換の有効/無効
  host: "example.com"  # リモートサーバーのホスト名
  port: 22  # SSHポート
  user: "webuser"  # リモートユーザー名
  remote_path: "/var/www/html/images"  # リモートパス
  use_ssh_agent: true  # SSH Agent認証の使用

# 実行モード設定
mode:
  dry_run: false  # ドライランモード（テスト実行）

# 入力ディレクトリと対象拡張子
input:
  directory: "./images"
  supported_extensions: [.jpg, .jpeg, .png, .heic, .heif]

# 変換設定
conversion:
  workers: 4  # 並列ワーカー数
  webp:
    enabled: true
    quality: 80  # 画質（0-100）
  avif:
    enabled: true
    quality: 80  # 画質（0-100）
    speed: 6     # 処理速度（0-10）
    lossless: false  # ロスレス圧縮

# FTPサーバー設定
ftp:
  enabled: false
  port: 2121

# SSHサーバー設定
ssh:
  enabled: false
  port: 2222
```

詳細な設定オプションは`config.yml`ファイルを参照してください。

## 注意事項

- FTPおよびSSHサーバーを有効にする場合、適切な権限と認証情報の設定が必要です
- 大量の画像を処理する場合は、メモリとCPUの使用量に注意してください
- 画像フォーマットによっては変換時間が異なる場合があります
