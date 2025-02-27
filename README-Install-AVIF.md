# AVIF 依存関係のインストール手順

AVIF変換機能を使用するには、`libaom`ライブラリがシステムにインストールされている必要があります。このドキュメントでは、各OSでのインストール方法について説明します。

## Debian / Ubuntu

```bash
sudo apt-get update
sudo apt-get install libaom-dev
```

## CentOS / RHEL / Fedora

```bash
sudo dnf install libaom-devel
```

または、古いバージョンのCentOSを使用している場合：

```bash
sudo yum install libaom-devel
```

## macOS

Homebrewを使用してインストールします：

```bash
brew install aom
```

## Windows

MSYS2やMinGWを使用している場合：

```bash
pacman -S mingw-w64-x86_64-aom
```

または、vcpkgを使用している場合：

```bash
vcpkg install aom
```

## Docker環境での注意点

Dockerで開発/実行環境を構築する場合、Dockerfileに以下のように依存関係を追加してください：

```dockerfile
# Debian/Ubuntuベースの場合
RUN apt-get update && apt-get install -y libaom-dev

# CentOS/RHELベースの場合
RUN dnf install -y libaom-devel
```

## トラブルシューティング

### ライブラリが見つからないエラー

実行時に以下のようなエラーが表示される場合：

```
error while loading shared libraries: libaom.so.3: cannot open shared object file: No such file or directory
```

システムにlibaomがインストールされていることを確認し、必要に応じてライブラリパスを設定してください：

```bash
export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
```

### CGO関連のエラー

コンパイル時に以下のようなエラーが表示される場合：

```
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in $PATH
```

Cコンパイラとビルドツールがインストールされていることを確認してください：

```bash
# Debian/Ubuntu
sudo apt-get install build-essential

# CentOS/RHEL
sudo dnf group install "Development Tools"
```
