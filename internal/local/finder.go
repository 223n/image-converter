package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/223n/image-converter/internal/config"
)

// FileFinder はローカルファイルシステムからの画像ファイル検索を担当します
type FileFinder struct {
	config              *config.Config // ポインタとして設定
	supportedExtensions map[string]bool
}

// NewFileFinder は新しいファイル検索インスタンスを作成します
func NewFileFinder(cfg *config.Config) *FileFinder {
	// サポートされている拡張子をマップに変換
	supportedExtensions := make(map[string]bool)
	for _, ext := range cfg.Input.SupportedExtensions {
		supportedExtensions[strings.ToLower(ext)] = true
	}

	return &FileFinder{
		config:              cfg,
		supportedExtensions: supportedExtensions,
	}
}

// FindFiles は対象ディレクトリから変換対象の画像ファイルを検索します
func (f *FileFinder) FindFiles() ([]string, int, error) {
	// 入力ディレクトリの存在チェック
	if err := f.validateDirectory(); err != nil {
		return nil, 0, err
	}

	// ファイル検索
	files, err := f.searchFiles()
	if err != nil {
		return nil, 0, fmt.Errorf("ファイル検索に失敗しました: %w", err)
	}

	return files, len(files), nil
}

// validateDirectory は入力ディレクトリの存在を確認します
func (f *FileFinder) validateDirectory() error {
	info, err := os.Stat(f.config.Input.Directory)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("指定されたディレクトリが存在しません: %s", f.config.Input.Directory)
		}
		return fmt.Errorf("ディレクトリの情報取得に失敗しました: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("指定されたパスはディレクトリではありません: %s", f.config.Input.Directory)
	}

	return nil
}

// searchFiles は再帰的にファイルを検索します
func (f *FileFinder) searchFiles() ([]string, error) {
	var filesToConvert []string

	err := filepath.Walk(f.config.Input.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 拡張子がサポート対象かチェック
		ext := strings.ToLower(filepath.Ext(path))
		if f.supportedExtensions[ext] {
			filesToConvert = append(filesToConvert, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// サポートされるファイルが見つからない場合
	if len(filesToConvert) == 0 {
		return nil, fmt.Errorf("対象ディレクトリに変換対象のファイルが見つかりません: %s", f.config.Input.Directory)
	}

	return filesToConvert, nil
}

// GetSupportedExtensions はサポートされている拡張子のマップを返します
func (f *FileFinder) GetSupportedExtensions() map[string]bool {
	return f.supportedExtensions
}

// FilterDuplicates は既に変換済みのファイルをフィルタリングします
func (f *FileFinder) FilterDuplicates(files []string) []string {
	var filtered []string

	for _, file := range files {
		// 既にWebPまたはAVIFファイルが存在するかチェック
		baseName := strings.TrimSuffix(file, filepath.Ext(file))
		dir := filepath.Dir(file)

		// WebPファイルの存在チェック
		webpPath := filepath.Join(dir, baseName+".webp")
		webpExists := fileExists(webpPath)

		// AVIFファイルの存在チェック
		avifPath := filepath.Join(dir, baseName+".avif")
		avifExists := fileExists(avifPath)

		// WebPとAVIFの両方が既に存在する場合はスキップ
		if f.config.Conversion.WebP.Enabled && f.config.Conversion.AVIF.Enabled && webpExists && avifExists {
			continue
		}

		// WebPのみが有効で、WebPが既に存在する場合はスキップ
		if f.config.Conversion.WebP.Enabled && !f.config.Conversion.AVIF.Enabled && webpExists {
			continue
		}

		// AVIFのみが有効で、AVIFが既に存在する場合はスキップ
		if !f.config.Conversion.WebP.Enabled && f.config.Conversion.AVIF.Enabled && avifExists {
			continue
		}

		// それ以外はフィルタリングされたリストに追加
		filtered = append(filtered, file)
	}

	return filtered
}

// fileExists はファイルが存在するかどうかをチェックします
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
