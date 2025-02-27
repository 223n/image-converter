# Makefileの使い方

## 基本的なビルドと実行

- `make` または `make all` - コードのフォーマットとビルドを実行
- `make build` - プログラムをビルドしてバイナリを生成
- `make run` - ビルドして実行
- `make run-dry` - ドライランモードで実行（変換処理は行わず、対象ファイルのみ確認）

## クリーンアップとフォーマット

- `make clean` - ビルドされたバイナリを削除
- `make fmt` - Goコードをフォーマット

## 依存関係の管理

- `make install-deps` - Goの依存パッケージをインストール
- `make install-system-deps-debian` - Debian/Ubuntu系のシステム依存パッケージをインストール
- `make install-system-deps-redhat` - RHEL/CentOS/Fedora系のシステム依存パッケージをインストール

## デプロイメント

- `make install` - ビルドしたバイナリとコンフィグファイルをシステムにインストール
- `make build-static` - 静的リンクされたポータブルバイナリを生成
- `make docker` - Dockerイメージをビルド

## その他

- `make test` - テストを実行
- `make help` - 使用可能なコマンドの一覧とその説明を表示

## 使用例

```bash
# ビルドと実行
make run

# ドライランモードで実行
make run-dry

# 依存関係のインストールとビルド
make install-deps
make

# システムにインストール（管理者権限が必要）
sudo make install

# Debian/Ubuntu での依存パッケージのインストール
sudo make install-system-deps-debian
```

```bash
make install-system-deps-debian
make build-safe
```
