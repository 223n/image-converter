# インストール手順

この文書では、画像変換ツールのインストール方法と依存関係のセットアップについて説明します。

## 目次

- [インストール手順](#インストール手順)
  - [目次](#目次)
  - [環境要件](#環境要件)
  - [依存関係のインストール](#依存関係のインストール)
    - [Debian/Ubuntu](#debianubuntu)
    - [RHEL/CentOS/Fedora](#rhelcentosfedora)
    - [macOS](#macos)
    - [Windows](#windows)
  - [Goのインストール](#goのインストール)
  - [プロジェクトのビルド](#プロジェクトのビルド)
  - [開発環境のセットアップ](#開発環境のセットアップ)
  - [Docker環境](#docker環境)
  - [トラブルシューティング](#トラブルシューティング)
    - [libaomが見つからない場合](#libaomが見つからない場合)
    - [CGO関連のエラー](#cgo関連のエラー)

## 環境要件

- **オペレーティングシステム**: Linux, macOS, Windows（WSL推奨）
- **Go**: バージョン1.16以降
- **メモリ**: 最低4GB（大量の画像を処理する場合は8GB以上推奨）
- **ディスク容量**: 最低1GB

## 依存関係のインストール

画像変換ツールでは、WebPとAVIFの変換に外部ライブラリを使用します。それぞれのOSでのインストール方法は以下の通りです。

### Debian/Ubuntu

```bash
# 必要なパッケージのインストール
sudo apt-get update
sudo apt-get install -y build-essential libaom-dev webp

# スクリプトによるインストール
sudo make install-system-deps-debian
```

### RHEL/CentOS/Fedora

```bash
# 必要なパッケージのインストール
sudo dnf install -y gcc-c++ libaom-devel libwebp-tools

# スクリプトによるインストール
sudo make install-system-deps-redhat
```

### macOS

Homebrewを使用してインストールします：

```bash
# Homebrewがなければインストール
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 依存パッケージをインストール
brew install aom webp

# Go（インストールされていない場合）
brew install go
```

### Windows

Windows環境では、Windows Subsystem for Linux (WSL)の使用を推奨します：

1. WSLをインストール：[https://docs.microsoft.com/ja-jp/windows/wsl/install](https://docs.microsoft.com/ja-jp/windows/wsl/install)
2. Ubuntu等のディストリビューションをインストール
3. 上記のDebian/Ubuntuの手順に従う

あるいは、MSYSやMinGWを使用する場合：

```bash
pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-aom mingw-w64-x86_64-libwebp
```

## Goのインストール

公式サイトからGoをインストールしてください：[https://golang.org/doc/install](https://golang.org/doc/install)

バージョン1.16以降が必要です。

## プロジェクトのビルド

リポジトリをクローンしてビルドします：

```bash
# リポジトリのクローン
git clone https://github.com/yourusername/image-converter.git
cd image-converter

# 依存パッケージのインストール
make install-deps

# ビルド
make build-safe

# または代替ビルド方法
./scripts/build.sh
```

## 開発環境のセットアップ

VS Code Dev Containersを使用した開発環境のセットアップ：

1. [VS Code](https://code.visualstudio.com/)をインストール
2. [Remote - Containers拡張機能](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)をインストール
3. [Docker Desktop](https://www.docker.com/products/docker-desktop)をインストール
4. VS Codeでプロジェクトを開き、「Reopen in Container」を選択

これにより、すべての依存関係があらかじめ設定された開発環境が自動的に構築されます。

## Docker環境

Dockerを使用してビルド・実行することも可能です：

```bash
# Dockerイメージのビルド
make docker

# Dockerでの実行
make docker-run
```

## トラブルシューティング

### libaomが見つからない場合

```bash
error while loading shared libraries: libaom.so.3: cannot open shared object file: No such file or directory
```

解決策:

```bash
# ライブラリパスの更新
export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

# または、システムに適切にインストールされているか確認
sudo ldconfig
```

### CGO関連のエラー

```bash
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in $PATH
```

解決策:

```bash
# Debian/Ubuntu
sudo apt-get install build-essential

# CentOS/RHEL
sudo dnf group install "Development Tools"
```

その他のトラブルシューティングについては[トラブルシューティングガイド](TROUBLESHOOTING.md)を参照してください。
