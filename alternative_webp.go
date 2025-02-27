package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
)

// saveWebPAlternative は外部コマンド（cwebpツール）を使用してWebP画像を保存します
// libwebpの開発ヘッダーが問題を起こす場合の代替手段として使用します
func saveWebPAlternative(img image.Image, outputPath string, quality int) error {
	// 一時的にPNGとして保存
	tempDir, err := os.MkdirTemp("", "webp-conversion-")
	if err != nil {
		return fmt.Errorf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempPNGPath := filepath.Join(tempDir, "temp.png")

	// 一時PNGファイルの作成
	tempFile, err := os.Create(tempPNGPath)
	if err != nil {
		return fmt.Errorf("一時ファイルの作成に失敗しました: %v", err)
	}

	// PNGとして一時保存
	if err := png.Encode(tempFile, img); err != nil {
		tempFile.Close()
		return fmt.Errorf("PNGエンコードに失敗しました: %v", err)
	}
	tempFile.Close()

	// cwebpコマンドが利用可能か確認
	if _, err := exec.LookPath("cwebp"); err != nil {
		// cwebpがインストールされていない場合はインストールを促す
		return fmt.Errorf("cwebpコマンドが見つかりません。次のコマンドでインストールしてください: sudo apt-get install webp")
	}

	// cwebpを使ってWebPに変換
	cmd := exec.Command("cwebp", "-q", fmt.Sprintf("%d", quality), tempPNGPath, "-o", outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cwebpコマンドの実行に失敗しました: %v\n出力: %s", err, string(output))
	}

	return nil
}

// isWebPLibraryAvailable はlibwebpライブラリが適切に使用可能かを確認します
func isWebPLibraryAvailable() bool {
	// libwebp-devがインストールされているか確認（ヘッダーファイルの存在をチェック）
	if _, err := os.Stat("/usr/include/webp/encode.h"); err != nil {
		return false
	}

	return true
}
