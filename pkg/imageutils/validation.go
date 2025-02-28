/*
Package imageutils は画像処理に関連する再利用可能なユーティリティ関数を提供します。
*/
package imageutils

import (
	"fmt"
	"image"
	_ "image/gif"  // GIFデコーダを登録
	_ "image/jpeg" // JPEGデコーダを登録
	_ "image/png"  // PNGデコーダを登録
	"log"
	"os"
	"path/filepath"
	"strings"
)

// IsValidImage は画像ファイルが有効かどうかを確認します
func IsValidImage(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("画像ファイルを開けません: %v", err)
		return false
	}
	defer file.Close()

	// 画像のデコードを試みる
	_, format, err := image.DecodeConfig(file)
	if err != nil {
		log.Printf("画像のデコードに失敗しました: %s - %v", path, err)
		return false
	}

	// サポートされているフォーマットかどうかを確認
	switch strings.ToLower(format) {
	case "jpeg", "jpg", "png", "gif", "webp", "avif", "heic", "heif":
		return true
	default:
		log.Printf("サポートされていない画像形式です: %s - %s", path, format)
		return false
	}
}

// IsValidImageDimensions は画像の寸法が指定された範囲内かどうかを確認します
func IsValidImageDimensions(path string, minWidth, minHeight, maxWidth, maxHeight int) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("ファイルを開けません: %v", err)
	}
	defer file.Close()

	// 画像の寸法を取得
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return false, fmt.Errorf("画像のデコードに失敗しました: %v", err)
	}

	// 寸法が範囲内かどうかをチェック
	if config.Width < minWidth || config.Height < minHeight {
		return false, fmt.Errorf("画像が小さすぎます: %dx%d (最小: %dx%d)",
			config.Width, config.Height, minWidth, minHeight)
	}

	if maxWidth > 0 && config.Width > maxWidth || maxHeight > 0 && config.Height > maxHeight {
		return false, fmt.Errorf("画像が大きすぎます: %dx%d (最大: %dx%d)",
			config.Width, config.Height, maxWidth, maxHeight)
	}

	return true, nil
}

// IsValidFile はファイルが有効かどうかを確認し、そのサイズを返します
func IsValidFile(path string) (bool, int64) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("ファイル情報の取得に失敗しました: %v", err)
		return false, 0
	}

	// 0バイトファイルはスキップ
	if fileInfo.Size() == 0 {
		log.Printf("ファイルサイズが0バイトです: %s", path)
		return false, 0
	}

	// 画像ファイルの場合は追加チェック
	if IsImageExt(filepath.Ext(path)) {
		if !IsValidImage(path) {
			return false, fileInfo.Size()
		}
	}

	return true, fileInfo.Size()
}

// IsValidImageSize はファイルサイズが指定された範囲内かどうかを確認します
func IsValidImageSize(path string, minSize, maxSize int64) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("ファイル情報の取得に失敗しました: %v", err)
	}

	size := fileInfo.Size()

	// サイズが範囲内かどうかをチェック
	if size < minSize {
		return false, fmt.Errorf("ファイルサイズが小さすぎます: %d バイト (最小: %d バイト)",
			size, minSize)
	}

	if maxSize > 0 && size > maxSize {
		return false, fmt.Errorf("ファイルサイズが大きすぎます: %d バイト (最大: %d バイト)",
			size, maxSize)
	}

	return true, nil
}

// IsImageBroken は画像ファイルが破損しているかどうかを詳細に確認します
func IsImageBroken(path string) (bool, string) {
	file, err := os.Open(path)
	if err != nil {
		return true, fmt.Sprintf("ファイルを開けません: %v", err)
	}
	defer file.Close()

	// 画像の完全な読み込みを試みる
	_, _, err = image.Decode(file)
	if err != nil {
		return true, fmt.Sprintf("画像の完全なデコードに失敗しました: %v", err)
	}

	return false, ""
}
