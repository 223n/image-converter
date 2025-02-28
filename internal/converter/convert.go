/*
Package converter は画像変換の主要ロジックを提供します。
JPG、PNG、HEIC、HEIFなどの画像フォーマットをWebPとAVIFに変換する機能を実装しています。
*/
package converter

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jdeng/goheif"
	"github.com/yourusername/image-converter/internal/config"
	"github.com/yourusername/image-converter/pkg/imageutils"
)

// Service は画像変換サービスを表します
type Service struct {
	// 将来的な拡張のためのフィールドを追加できます
}

// NewService は新しい変換サービスを作成します
func NewService() *Service {
	return &Service{}
}

// ConvertImage は画像をWebPとAVIFに変換します
func (s *Service) ConvertImage(filePath string) error {
	// 入力画像の読み込み
	img, err := loadImage(filePath)
	if err != nil {
		return err
	}

	// パスの構築
	baseFileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	dir := filepath.Dir(filePath)

	// WebP変換
	if err := s.convertToWebP(img, dir, baseFileName); err != nil {
		return err
	}

	// AVIF変換
	if err := s.convertToAVIF(img, dir, baseFileName); err != nil {
		return err
	}

	log.Printf("変換処理完了: %s", filePath)
	return nil
}

// loadImage は画像を読み込んでデコードします
func loadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けません: %v", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	var img image.Image

	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	case ".heic", ".heif":
		img, err = goheif.Decode(file)
	default:
		return nil, fmt.Errorf("サポートされていない画像形式です: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("画像のデコードに失敗しました: %v", err)
	}

	return img, nil
}

// convertToWebP は画像をWebP形式に変換します
// このメソッドはwebp.goで実装される具体的な変換処理を呼び出します
func (s *Service) convertToWebP(img image.Image, dir, baseFileName string) error {
	if !config.IsWebPEnabled() {
		return nil
	}

	webpPath := filepath.Join(dir, baseFileName+".webp")

	// ドライランモードではスキップ
	if config.IsDryRun() {
		log.Printf("ドライラン: WebP変換のスキップ")
		return nil
	}

	if err := SaveWebP(img, webpPath); err != nil {
		log.Printf("WebP変換に失敗しました: %v", err)
		return err
	}

	// ファイルサイズをチェック
	if fi, err := os.Stat(webpPath); err == nil && fi.Size() > 0 {
		log.Printf("WebP変換成功: %s (サイズ: %d バイト)", webpPath, fi.Size())
		return nil
	}

	log.Printf("警告: WebP変換結果が異常です: %s", webpPath)
	return fmt.Errorf("WebP変換後のファイルが無効です")
}

// convertToAVIF は画像をAVIF形式に変換します
// このメソッドはavif.goで実装される具体的な変換処理を呼び出します
func (s *Service) convertToAVIF(img image.Image, dir, baseFileName string) error {
	if !config.IsAVIFEnabled() {
		return nil
	}

	avifPath := filepath.Join(dir, baseFileName+".avif")

	// ドライランモードではスキップ
	if config.IsDryRun() {
		log.Printf("ドライラン: AVIF変換対象: %s -> %s", baseFileName, avifPath)
		return nil
	}

	if err := SaveAVIF(img, avifPath); err != nil {
		log.Printf("AVIF変換に失敗しました: %v", err)
		return err
	}

	// ファイルサイズと整合性をチェック
	valid, fileSize := imageutils.IsValidFile(avifPath)
	if valid {
		log.Printf("AVIF変換成功: %s (サイズ: %d バイト)", avifPath, fileSize)
		return nil
	}

	log.Printf("警告: AVIF変換結果が無効です: %s", avifPath)
	// 無効なファイルを削除
	os.Remove(avifPath)
	return fmt.Errorf("AVIF変換後のファイルが無効です")
}

// CheckConversionResults は変換結果をチェックし、統計情報を更新します
func (s *Service) CheckConversionResults(file string, stats *config.ConversionStats) {
	ext := filepath.Ext(file)
	baseName := strings.TrimSuffix(file, ext)
	dir := filepath.Dir(file)

	// WebPファイルのチェック
	if config.IsWebPEnabled() {
		s.checkWebPResult(dir, baseName, stats)
	}

	// AVIFファイルのチェック
	if config.IsAVIFEnabled() {
		s.checkAVIFResult(dir, baseName, stats)
	}
}

// checkWebPResult はWebP変換結果をチェックします
func (s *Service) checkWebPResult(dir, baseName string, stats *config.ConversionStats) {
	webpPath := filepath.Join(dir, baseName+".webp")
	if fi, err := os.Stat(webpPath); err == nil && fi.Size() > 0 {
		stats.WebPSuccess++
		log.Printf("WebP変換成功: %s (サイズ: %d バイト)", webpPath, fi.Size())
	} else if err == nil {
		stats.WebPFailed++
		log.Printf("警告: WebP変換結果が0バイトです: %s", webpPath)
	}
}

// checkAVIFResult はAVIF変換結果をチェックします
func (s *Service) checkAVIFResult(dir, baseName string, stats *config.ConversionStats) {
	avifPath := filepath.Join(dir, baseName+".avif")
	if fi, err := os.Stat(avifPath); err == nil && fi.Size() > 0 {
		// ファイルの整合性チェック
		if imageutils.IsValidImage(avifPath) {
			stats.AVIFSuccess++
			log.Printf("AVIF変換成功: %s (サイズ: %d バイト)", avifPath, fi.Size())
		} else {
			stats.AVIFFailed++
			log.Printf("警告: AVIF変換結果が破損しています: %s", avifPath)
			// 破損ファイルを削除
			os.Remove(avifPath)
		}
	} else if err == nil {
		stats.AVIFFailed++
		log.Printf("警告: AVIF変換結果が0バイトです: %s", avifPath)
		// 0バイトファイルを削除
		os.Remove(avifPath)
	}
}

// CleanupFiles は処理済みのファイルを削除します
func (s *Service) CleanupFiles(localPath, baseName string) {
	// 元ファイルを削除
	os.Remove(localPath)

	// 変換後のファイルを削除
	dir := filepath.Dir(localPath)

	webpPath := filepath.Join(dir, baseName+".webp")
	if _, err := os.Stat(webpPath); err == nil {
		os.Remove(webpPath)
	}

	avifPath := filepath.Join(dir, baseName+".avif")
	if _, err := os.Stat(avifPath); err == nil {
		os.Remove(avifPath)
	}
}
