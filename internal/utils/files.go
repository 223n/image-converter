package utils

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// EnsureDirectoryExists はディレクトリが存在することを確認し、必要に応じて作成します
func EnsureDirectoryExists(path string) error {
	if path == "" {
		return fmt.Errorf("ディレクトリパスが空です")
	}

	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("パス %s はディレクトリではありません", path)
		}
		return nil // ディレクトリは既に存在する
	}

	if os.IsNotExist(err) {
		// ディレクトリが存在しないので作成
		return os.MkdirAll(path, 0755)
	}

	// 他のエラー
	return err
}

// GetFilesWithExtensions は指定されたディレクトリ内から、指定された拡張子を持つファイルの一覧を取得します
func GetFilesWithExtensions(directory string, extensions []string) ([]string, error) {
	var result []string

	// 拡張子のマップを作成
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extMap[ext] = true
	}

	// ディレクトリを歩いてファイルを見つける
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if extMap[ext] {
			result = append(result, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ディレクトリの検索中にエラーが発生しました: %v", err)
	}

	return result, nil
}

// IsValidImage は画像ファイルが有効かどうかを確認します
func IsValidImage(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("画像ファイルを開けません: %v", err)
		return false
	}
	defer file.Close()

	// 画像のデコードを試みる
	_, _, err = image.DecodeConfig(file)
	if err != nil {
		log.Printf("画像のデコードに失敗しました: %s - %v", path, err)
		return false
	}

	return true
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

// IsImageExt は拡張子が画像ファイルかどうかを判断します
func IsImageExt(ext string) bool {
	ext = strings.ToLower(ext)
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".avif": true,
		".heic": true,
		".heif": true,
	}
	return imageExts[ext]
}

// CleanupFiles は一時ファイルを削除します
func CleanupFiles(filePaths ...string) {
	for _, path := range filePaths {
		if path != "" {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				log.Printf("ファイルの削除に失敗しました: %s - %v", path, err)
			}
		}
	}
}

// GetFileSize はファイルのサイズをバイト単位で返します
func GetFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// FormatFileSize はファイルサイズを人間が読みやすい形式にフォーマットします
func FormatFileSize(size int64) string {
	const (
		B  = 1
		KB = B * 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d バイト", size)
	}
}
