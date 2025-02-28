# 設定ガイド

この文書では、画像変換ツールの設定ファイルの詳細と設定オプションについて説明します。

## 目次

- [設定ガイド](#設定ガイド)
  - [目次](#目次)
  - [設定ファイルの概要](#設定ファイルの概要)
  - [設定オプション](#設定オプション)
    - [リモート設定](#リモート設定)
    - [実行モード設定](#実行モード設定)
    - [入力設定](#入力設定)
    - [変換設定](#変換設定)
    - [FTPサーバー設定](#ftpサーバー設定)
    - [SSHサーバー設定](#sshサーバー設定)
    - [ログ設定](#ログ設定)
  - [設定例](#設定例)
    - [高品質変換設定](#高品質変換設定)
    - [高速変換設定](#高速変換設定)
    - [容量優先設定](#容量優先設定)
  - [環境別の推奨設定](#環境別の推奨設定)
    - [開発環境](#開発環境)
    - [本番環境](#本番環境)
    - [リソース制限環境](#リソース制限環境)
    - [大量処理環境](#大量処理環境)

## 設定ファイルの概要

設定ファイルはYAML形式で記述され、通常 `config.yml` という名前で保存されます。コマンドラインから `-config` オプションを使用して別の設定ファイルを指定することもできます。

設定ファイルは以下の主要なセクションで構成されています：

- `remote`: リモートサーバーへの接続設定
- `mode`: 実行モードの設定
- `input`: 入力ディレクトリと対象拡張子の設定
- `conversion`: 変換設定（並列数、品質等）
- `ftp`: FTPサーバー設定
- `ssh`: SSHサーバー設定
- `logging`: ログ設定

## 設定オプション

### リモート設定

リモートサーバー接続のための設定です。

```yaml
remote:
  # リモート変換を有効/無効
  enabled: false
  # リモートサーバーのホスト名またはIPアドレス
  host: "example.com"
  # SSHポート
  port: 22
  # リモートユーザー名
  user: "webuser"
  # 秘密鍵のパス（空の場合はSSH Agentを使用）
  key_path: ""
  # 既知のホストファイルのパス（空の場合は検証を無効化）
  known_hosts: "~/.ssh/known_hosts"
  # リモートサーバー上の変換対象パス
  remote_path: "/var/www/html/images"
  # SSH Agentを使用するかどうか
  use_ssh_agent: true
  # タイムアウト（秒）
  timeout: 60
```

### 実行モード設定

実行モードに関する設定です。

```yaml
mode:
  # ドライラン（true=実際の変換を行わず、ログのみ出力）
  dry_run: false
```

### 入力設定

入力ディレクトリと対象ファイル形式の設定です。

```yaml
input:
  # 画像を検索するディレクトリ
  directory: "./images"
  # サポートする拡張子（変換対象ファイル形式）
  supported_extensions:
    - .jpg
    - .jpeg
    - .png
    - .heic
    - .heif
```

### 変換設定

変換処理に関する設定です。

```yaml
conversion:
  # 並列処理するワーカー数
  workers: 4
  # WebP変換設定
  webp:
    # 変換を有効/無効
    enabled: true
    # 画質設定（0-100）
    quality: 80
    # 圧縮レベル（0-6、値が大きいほど圧縮率が高い）
    compression_level: 4
  # AVIF変換設定
  avif:
    # 変換を有効/無効
    enabled: true
    # 画質設定（1-63）
    quality: 40
    # 速度設定（0-10、値が小さいほど品質が高いが処理は遅くなる）
    # 0=最高品質/最低速度、10=最低品質/最高速度
    speed: 6
    # ロスレス圧縮（trueの場合、qualityは無視される）
    lossless: false
```

### FTPサーバー設定

組み込みFTPサーバーの設定です。

```yaml
ftp:
  # FTPサーバーを有効/無効
  enabled: false
  # FTPサーバーのポート
  port: 2121
  # FTPユーザー設定
  user:
    # ユーザー名
    name: "ftpuser"
    # パスワード（本番環境では平文で保存しないでください）
    password: "ftppassword"
  # パッシブモード設定
  passive:
    # パッシブモードを有効/無効
    enabled: true
    # パッシブポート範囲
    port_range: "50000-50100"
```

### SSHサーバー設定

組み込みSSHサーバーの設定です。

```yaml
ssh:
  # SSHサーバーを有効/無効
  enabled: false
  # SSHサーバーのポート
  port: 2222
  # SSH認証設定
  auth:
    # パスワード認証を有効/無効
    password_auth: true
    # 公開鍵認証（常に有効、ssh-agentの登録済み鍵が使用されます）
    pubkey_auth: true
    auth_keys_file: "~/.ssh/authorized_keys"
```

### ログ設定

ログ出力に関する設定です。

```yaml
logging:
  # ログレベル（debug, info, warn, error）
  level: "info"
  # ログファイルの出力先（空の場合は標準出力のみ）
  file: "image-converter.log"
  # ログファイルの最大サイズ（MB）
  max_size: 10
  # 保持するログファイルの数
  max_backups: 3
  # ログファイルの最大保持日数
  max_age: 28
  # ログを圧縮するかどうか
  compress: true
```

## 設定例

### 高品質変換設定

```yaml
conversion:
  workers: 2
  webp:
    enabled: true
    quality: 95
    compression_level: 2
  avif:
    enabled: true
    quality: 60
    speed: 2
    lossless: false
```

### 高速変換設定

```yaml
conversion:
  workers: 8
  webp:
    enabled: true
    quality: 75
    compression_level: 6
  avif:
    enabled: true
    quality: 30
    speed: 9
    lossless: false
```

### 容量優先設定

```yaml
conversion:
  workers: 4
  webp:
    enabled: true
    quality: 65
    compression_level: 6
  avif:
    enabled: true
    quality: 20
    speed: 7
    lossless: false
```

## 環境別の推奨設定

### 開発環境

```yaml
mode:
  dry_run: false
conversion:
  workers: 2
logging:
  level: "debug"
```

### 本番環境

```yaml
mode:
  dry_run: false
conversion:
  workers: 4
logging:
  level: "info"
  file: "/var/log/image-converter.log"
  max_size: 100
  max_backups: 10
```

### リソース制限環境

```yaml
conversion:
  workers: 1
  webp:
    enabled: true
    quality: 80
  avif:
    enabled: false  # AVIFは無効化
```

### 大量処理環境

```yaml
conversion:
  workers: 8
logging:
  level: "warn"  # 警告以上のみログ記録
```
