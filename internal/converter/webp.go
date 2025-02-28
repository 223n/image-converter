/*
Package converter の一部として、WebP変換に特化した関数を提供します。
*/
package converter

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/chai2010/webp"
	"github.com/yourusername/image-converter/internal/config"
)

// SaveWebP は画像をWebPとして保存します
func SaveWebP(img image.Image, outputPath string) error {
	// 最適なWebPエンコーダーを選択
	encoder := selectBestWebPEncoder()

	switch encoder {
	case "cwebp":
		// cwebpコマンドを使用
		return saveWebPUsingCommand(img, outputPath, config.GetWebPQuality())
	case "libwebp":
		// libwebpを直接使用（必要に応じて実装）
		// 現在はsaveWebPUsingCommandを使用
		return saveWebPUsingCommand(img, outputPath, config.GetWebPQuality())
	default:
		// Goのwebpライブラリを使用
		return saveWebPUsingLibrary(img, outputPath)
	}
}

// saveWebPUsingLibrary はGoのWebPライブラリを使用して保存します
func saveWebPUsingLibrary(img image.Image, outputPath string) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("出力ファイルの作成に失敗しました: %v", err)
	}
	defer output.Close()

	opts := &webp.Options{
		Lossless: false,
		Quality:  float32(config.GetWebPQuality()),
	}

	if err := webp.Encode(output, img, opts); err != nil {
		return fmt.Errorf("WebPエンコードに失敗しました: %v", err)
	}

	return nil
}

// saveWebPUsingCommand は外部コマンド（cwebpツール）を使用してWebP画像を保存します
func saveWebPUsingCommand(img image.Image, outputPath string, quality int) error {
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

// selectBestWebPEncoder はWebP変換の最適な方法を選択します
func selectBestWebPEncoder() string {
	// 優先順位:
	// 1. cwebp コマンド (最も信頼性が高い)
	// 2. libwebp ライブラリ (ヘッダーファイルが正しくインストールされている場合)
	// 3. Go製webpライブラリ (最後の手段)

	// cwebpコマンドが利用可能か確認
	if _, err := exec.LookPath("cwebp"); err == nil {
		log.Printf("WebP変換: cwebpコマンドを使用します")
		return "cwebp"
	}

	// libwebpライブラリが使用可能か確認
	if isWebPLibraryAvailable() {
		log.Printf("WebP変換: libwebpライブラリを使用します")
		return "libwebp"
	}

	// どちらも利用できない場合はGoのwebpライブラリを使用
	log.Printf("WebP変換: Goのwebpライブラリを使用します")
	return "gowebp"
}

// isWebPLibraryAvailable はlibwebpライブラリが適切に使用可能かを確認します
func isWebPLibraryAvailable() bool {
	// libwebp-devがインストールされているか確認（ヘッダーファイルの存在をチェック）
	if _, err := os.Stat("/usr/include/webp/encode.h"); err != nil {
		return false
	}

	return true
}
