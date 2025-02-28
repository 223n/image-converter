# トラブルシューティングガイド

この文書では、画像変換ツールの使用中に発生する可能性のある問題とその解決方法について説明します。

## 目次

- [インストールの問題](#インストールの問題)
- [設定の問題](#設定の問題)
- [WebP変換の問題](#webp変換の問題)
- [AVIF変換の問題](#avif変換の問題)
- [リモート接続の問題](#リモート接続の問題)
- [サーバー機能の問題](#サーバー機能の問題)
- [パフォーマンスの問題](#パフォーマンスの問題)
- [一般的なエラーメッセージ](#一般的なエラーメッセージ)
- [ログの読み方](#ログの読み方)

## インストールの問題

### 依存関係のエラー

**問題**: ビルド時に依存関係に関するエラーが表示される。

**解決策**:

1. Goの依存関係を更新：

```bash
go mod tidy
```

2. システム依存関係をインストール：

```bash
# Debian/Ubuntu
sudo apt-get install libaom-dev webp
# RHEL/Fedora
sudo dnf install libaom-devel libwebp-tools
```

### CGO関連のエラー

**問題**: `cgo: C compiler "gcc" not found` というエラーが表示される。

**解決策**:

1. Cコンパイラをインストール：

```bash
# Debian/Ubuntu
sudo apt-get install build-essential
# RHEL/Fedora
sudo dnf group install "Development Tools"
```

2. CGOを無効にしてビルド（機能制限あり）：

```bash
CGO_ENABLED=0 go build
```

## 設定の問題

### 設定ファイルが読み込まれない

**問題**: カスタム設定ファイルを指定しても適用されない。

**解決策**:

1. 設定ファイルのパスが正しいことを確認：

```bash
./image-converter -config=/正確な/パス/config.yml
```

2. 設定ファイルの構文が正しいことを確認：

```bash
yamllint config.yml
```

### 無効な設定値

**問題**: 設定値に関するエラーが表示される。

**解決策**:

- YAML形式を確認（インデントや特殊文字の使用に注意）
- 数値パラメーターの範囲を確認（例：AVIF品質は1-63）

## WebP変換の問題

### WebP変換失敗

**問題**: WebP変換に失敗する。

**解決策**:

1. cwebpコマンドがインストールされているか確認：

```bash
which cwebp
```

2. 代替変換方法を使用：

```yaml
# 設定ファイルでWebP品質を調整
conversion:
        webp:
        quality: 80
```

3. インストール状態を確認：

```bash
# Debian/Ubuntu
sudo apt-get install --reinstall webp
```

### WebP変換結果が0バイト

**問題**: WebP変換後のファイルが0バイト。

**解決策**:

1. 元の画像ファイルが正常かを確認
2. 一時的に異なる品質設定を試す
3. メモリ不足でないか確認

## AVIF変換の問題

### AVIF変換失敗

**問題**:「options error: bad quality value」というエラーが表示される。

**解決策**:

- AVIF品質値を1-63の範囲内に設定：

```yaml
conversion:
        avif:
        quality: 40  # 1-63の間で設定
```

### libaomが見つからない

**問題**: `libaom.so.3: cannot open shared object file` というエラーが表示される。

**解決策**:

1. libaomライブラリをインストール：

```bash
# Debian/Ubuntu
sudo apt-get install libaom-dev
# RHEL/Fedora
sudo dnf install libaom-devel
```

2. ライブラリパスを更新：

```bash
export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
```

## リモート接続の問題

### SSH接続失敗

**問題**: リモートサーバーへの接続が失敗する。

**解決策**:

1. ホスト名、ポート、ユーザー名が正しいか確認
2. 手動でSSH接続を試してみる：

```bash
ssh webuser@example.com
```

3. SSHデバッグ情報を確認：

```bash
ssh -vv webuser@example.com
```

### 接続が切断される

**問題**:「connection lost」などのエラーでリモート処理が途中で中断される。

**解決策**:

1. タイムアウト設定を増やす：

```yaml
remote:
        timeout: 120  # 秒単位
```

2. 処理するファイル数を減らす
3. ネットワーク接続の安定性を確認

### 認証エラー

**問題**: SSH認証エラーが発生する。

**解決策**:

1. SSH Agentが実行されているか確認：

```bash
echo $SSH_AUTH_SOCK
```

2. 鍵が登録されているか確認：

```bash
ssh-add -l
```

3. 秘密鍵のパーミッションを確認：

```bash
chmod 600 ~/.ssh/id_rsa
```

## サーバー機能の問題

### FTPサーバー起動失敗

**問題**: FTPサーバーが起動しない。

**解決策**:

1. pure-ftpdがインストールされているか確認：

```bash
which pure-ftpd
```

2. ポートがすでに使用されていないか確認：

```bash
sudo netstat -tulpn | grep :2121
```

3. 管理者権限で実行しているか確認

### SSHサーバー起動失敗

**問題**: SSHサーバーが起動しない。

**解決策**:

1. sshdがインストールされているか確認：

```bash
which sshd
```

2. ポートがすでに使用されていないか確認：

```bash
sudo netstat -tulpn | grep :2222
```

3. ホストキーファイルが存在するか確認

## パフォーマンスの問題

### 処理が遅い

**問題**: 画像変換処理が予想より遅い。

**解決策**:

1. ワーカー数を増やす：

```yaml
conversion:
        workers: 8  # CPUコア数に応じて調整
```

2. AVIFの速度設定を調整：

```yaml
conversion:
        avif:
        speed: 8  # 高速化（品質は低下）
```

3. 一部のフォーマットのみを有効にする：

```yaml
conversion:
        webp:
        enabled: true
        avif:
        enabled: false  # 無効化
```

### メモリ使用量が多い

**問題**: メモリ使用量が多く、システムが遅くなる。

**解決策**:

1. ワーカー数を減らす：

```yaml
conversion:
        workers: 2
```

2. リモートモードでバッチサイズを減らす

## 一般的なエラーメッセージ

### "unsupported image format"

**原因**: サポートされていない画像形式。

**解決策**:

- 対象ファイルがJPG、PNG、HEIC、HEIFのいずれかであることを確認

### "file is corrupted or truncated"

**原因**: 画像ファイルが破損している。

**解決策**:

- 元のファイルを確認
- 別の画像ビューアーで開けるか試す

### "permission denied"

**原因**: ファイルまたはディレクトリへのアクセス権がない。

**解決策**:

- ファイルの所有者と権限を確認
- 必要に応じて権限を変更

## ログの読み方

ログファイルには以下の情報が含まれています：

### エラーレベル

- `INFO`: 一般的な情報（通常の動作）
- `WARNING`/`警告`: 潜在的な問題（処理は継続）
- `ERROR`/`エラー`: 特定の操作に失敗（他の処理は継続）
- `FATAL`/`致命的`: プログラム全体が停止するエラー

### ログの例

```log
2025/02/28 12:34:56 main.go:50: === 画像変換処理開始: 2025-02-28 12:34:56 ===
2025/02/28 12:34:56 main.go:51: 設定ファイル: config.yml
2025/02/28 12:34:57 converter.go:65: WebP変換成功: /path/to/image.webp (サイズ: 12345 バイト)
2025/02/28 12:34:58 converter.go:97: 警告: AVIF変換結果が破損しています: /path/to/image.avif
```

### トラブルシューティングに役立つ情報の見つけ方

1. エラーメッセージを検索：

```bash
grep -i error image-converter_*.log
```

2. 特定のファイルに関する情報を検索：

```bash
grep "filename.jpg" image-converter_*.log
```

3. 警告メッセージを検索：

```bash
grep -i "警告\|warning" image-converter_*.log
```

4. 処理結果の統計を確認：

```bash
grep -A 5 "=== 変換処理結果 ===" image-converter_*.log
```
