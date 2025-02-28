package imageutils

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"time"
)

// ImageInfo は画像に関する基本情報を保持する構造体です
type ImageInfo struct {
	Path      string    // ファイルパス
	Format    string    // 画像形式
	Width     int       // 幅（ピクセル）
	Height    int       // 高さ（ピクセル）
	Size      int64     // ファイルサイズ（バイト）
	ModTime   time.Time // 最終更新日時
	Channels  int       // カラーチャンネル数
	BitDepth  int       // ビット深度
	IsValid   bool      // 有効な画像かどうか
	ErrorInfo string    // エラー情報（無効な場合）
}

// GetImageInfo は画像ファイルの基本情報を取得します
func GetImageInfo(path string) (*ImageInfo, error) {
	info := &ImageInfo{
		Path: path,
	}

	// ファイル情報の取得
	fileInfo, err := os.Stat(path)
	if err != nil {
		info.IsValid = false
		info.ErrorInfo = fmt.Sprintf("ファイル情報の取得に失敗しました: %v", err)
		return info, err
	}

	info.Size = fileInfo.Size()
	info.ModTime = fileInfo.ModTime()

	// 0バイトファイルのチェック
	if info.Size == 0 {
		info.IsValid = false
		info.ErrorInfo = "ファイルサイズが0バイトです"
		return info, fmt.Errorf("ファイルサイズが0バイトです")
	}

	// 画像ファイルを開く
	file, err := os.Open(path)
	if err != nil {
		info.IsValid = false
		info.ErrorInfo = fmt.Sprintf("ファイルを開けません: %v", err)
		return info, err
	}
	defer file.Close()

	// 画像情報の取得
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		info.IsValid = false
		info.ErrorInfo = fmt.Sprintf("画像のデコードに失敗しました: %v", err)
		return info, err
	}

	info.Format = format
	info.Width = config.Width
	info.Height = config.Height

	// カラーモデルの情報を推測
	switch {
	case format == "jpeg" || format == "jpg":
		info.Channels = 3
		info.BitDepth = 8
	case format == "png":
		// PNGは様々なビット深度をサポート
		info.Channels = 4 // RGBAと仮定
		info.BitDepth = 8 // 一般的な値
	case format == "gif":
		info.Channels = 4 // RGBA
		info.BitDepth = 8
	case format == "webp":
		info.Channels = 4 // RGBA
		info.BitDepth = 8
	case format == "avif" || format == "heif" || format == "heic":
		info.Channels = 4  // RGBA
		info.BitDepth = 10 // 一般的な値
	default:
		info.Channels = 0
		info.BitDepth = 0
	}

	info.IsValid = true
	return info, nil
}

// GetAspectRatio は画像のアスペクト比を計算します
func GetAspectRatio(width, height int) float64 {
	if height == 0 {
		return 0
	}
	return float64(width) / float64(height)
}

// FormatImageDimensions は画像の寸法を文字列にフォーマットします
func FormatImageDimensions(width, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}

// FormatImageSize はファイルサイズを人間が読みやすい形式にフォーマットします
func FormatImageSize(size int64) string {
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

// GetImageInfoSummary は画像情報の要約を返します
func (info *ImageInfo) GetSummary() string {
	if !info.IsValid {
		return fmt.Sprintf("無効な画像: %s - %s", info.Path, info.ErrorInfo)
	}

	return fmt.Sprintf("%s: %s, %dx%d, %s, %s",
		info.Path,
		info.Format,
		info.Width, info.Height,
		FormatImageSize(info.Size),
		info.ModTime.Format("2006-01-02 15:04:05"))
}
