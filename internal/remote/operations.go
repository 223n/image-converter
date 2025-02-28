package remote

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/image-converter/internal/config"
	"github.com/yourusername/image-converter/internal/converter"
	"github.com/yourusername/image-converter/pkg/imageutils"
)

// retryConfig はリトライ処理の設定です
type retryConfig struct {
	MaxRetries  int           // 最大リトライ回数
	InitialWait time.Duration // 初回のリトライ待機時間
	MaxWait     time.Duration // 最大リトライ待機時間
	Factor      float64       // リトライ待機時間の増加係数
}

// newDefaultRetryConfig はデフォルトのリトライ設定を返します
func newDefaultRetryConfig() *retryConfig {
	return &retryConfig{
		MaxRetries:  3,
		InitialWait: 2 * time.Second,
		MaxWait:     30 * time.Second,
		Factor:      2.0,
	}
}

// withRetry は指定された関数をリトライ付きで実行します
func withRetry(fn func() error, config *retryConfig) error {
	var err error
	wait := config.InitialWait

	for attempt := 1; attempt <= config.MaxRetries+1; attempt++ {
		// 関数を実行
		err = fn()
		if err == nil {
			// 成功した場合は終了
			return nil
		}

		// 最後の試行の場合はエラーを返す
		if attempt > config.MaxRetries {
			return fmt.Errorf("最大リトライ回数(%d)に達しました: %w", config.MaxRetries, err)
		}

		// エラーログを出力
		log.Printf("操作に失敗しました（試行 %d/%d）: %v - %d秒後に再試行します",
			attempt, config.MaxRetries+1, err, int(wait.Seconds()))

		// 待機時間を調整（指数バックオフ）
		time.Sleep(wait)
		wait = time.Duration(float64(wait) * config.Factor)
		if wait > config.MaxWait {
			wait = config.MaxWait
		}
	}

	return err // ここには到達しない
}

// isConnectionError は接続関連のエラーかどうかを判断します
func isConnectionError(err error) bool {
	// エラーメッセージに特定の文字列が含まれているかチェック
	errorMsg := err.Error()
	connectionErrors := []string{
		"connection lost",
		"connection reset",
		"broken pipe",
		"timeout",
		"connection refused",
		"no route to host",
		"network is unreachable",
		"i/o timeout",
	}

	for _, errText := range connectionErrors {
		if strings.Contains(strings.ToLower(errorMsg), errText) {
			return true
		}
	}
	return false
}

// FindRemoteImages はリモートサーバー上の画像ファイルを検索します
func (c *Client) FindRemoteImages(extensions []string) ([]string, error) {
	// 拡張子をパイプ区切りの文字列に変換
	var extsFormatted []string
	for _, ext := range extensions {
		ext = strings.TrimPrefix(ext, ".")
		extsFormatted = append(extsFormatted, fmt.Sprintf("-name \"*.%s\"", ext))
	}
	extsStr := strings.Join(extsFormatted, " -o ")

	// findコマンドを作成
	cmd := fmt.Sprintf("find %s -type f \\( %s \\) | sort",
		c.config.RemotePath,
		extsStr)

	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		return nil, err
	}

	// 出力を行に分割
	files := strings.Split(strings.TrimSpace(output), "\n")

	// 空の行を除外
	var result []string
	for _, file := range files {
		if file != "" {
			result = append(result, file)
		}
	}

	return result, nil
}

// ProcessRemoteFile は単一のリモートファイルを処理します
func (c *Client) ProcessRemoteFile(remoteFile, tempDir string, stats *config.ConversionStats) error {
	// ベース名とディレクトリを取得
	baseFileName := filepath.Base(remoteFile)
	relPath, err := filepath.Rel(c.config.RemotePath, filepath.Dir(remoteFile))
	if err != nil {
		log.Printf("警告: 相対パスの計算に失敗しました: %v", err)
		relPath = ""
	}

	// ローカルのパスを作成
	localPath := filepath.Join(tempDir, relPath, baseFileName)

	// ファイルをダウンロード
	if err := c.DownloadFile(remoteFile, localPath); err != nil {
		log.Printf("エラー: ファイルのダウンロードに失敗しました %s: %v", remoteFile, err)
		stats.DownloadFailed++
		return err
	}

	// 変換サービスを作成
	convService := converter.NewService()

	// 画像を変換
	if err := convService.ConvertImage(localPath); err != nil {
		log.Printf("エラー: 画像の変換に失敗しました %s: %v", localPath, err)
		stats.ConvertFailed++
		return err
	}

	stats.TotalProcessed++

	// 変換結果をアップロード
	uploadSuccess := c.UploadConvertedFiles(localPath, remoteFile, baseFileName, stats)

	// 処理済みファイルを削除して一時ディレクトリの肥大化を防ぐ
	cleanupFiles(localPath, baseFileName)

	if !uploadSuccess {
		return fmt.Errorf("変換結果のアップロードに失敗しました: %s", localPath)
	}

	return nil
}

// UploadConvertedFiles は変換されたファイルをアップロードします
func (c *Client) UploadConvertedFiles(localPath, remoteFile, baseFileName string, stats *config.ConversionStats) bool {
	ext := filepath.Ext(localPath)
	baseName := strings.TrimSuffix(baseFileName, ext)

	// アップロード成功フラグ
	webpUploaded := c.uploadWebPFile(localPath, remoteFile, baseName, stats)
	avifUploaded := c.uploadAVIFFile(localPath, remoteFile, baseName, stats)

	return webpUploaded || avifUploaded
}

// uploadWebPFile はWebPファイルをアップロードします
func (c *Client) uploadWebPFile(localPath, remoteFile, baseName string, stats *config.ConversionStats) bool {
	if !config.IsWebPEnabled() {
		return false
	}

	webpLocalPath := filepath.Join(filepath.Dir(localPath), baseName+".webp")
	webpRemotePath := filepath.Join(filepath.Dir(remoteFile), baseName+".webp")

	if _, err := os.Stat(webpLocalPath); err == nil {
		valid, fileSize := imageutils.IsValidFile(webpLocalPath)
		if valid {
			if err := c.UploadFile(webpLocalPath, webpRemotePath); err != nil {
				log.Printf("エラー: WebPファイルのアップロードに失敗しました %s: %v", webpLocalPath, err)
				stats.WebPFailed++
				return false
			}
			stats.WebPSuccess++
			stats.UploadedFiles++
			log.Printf("WebPファイルのアップロード成功: %s (サイズ: %d バイト)", webpRemotePath, fileSize)
			return true
		} else {
			log.Printf("警告: WebPファイルが無効なためスキップします: %s", webpLocalPath)
			stats.WebPFailed++
			stats.SkippedUploads++
		}
	}
	return false
}

// uploadAVIFFile はAVIFファイルをアップロードします
func (c *Client) uploadAVIFFile(localPath, remoteFile, baseName string, stats *config.ConversionStats) bool {
	if !config.IsAVIFEnabled() {
		return false
	}

	avifLocalPath := filepath.Join(filepath.Dir(localPath), baseName+".avif")
	avifRemotePath := filepath.Join(filepath.Dir(remoteFile), baseName+".avif")

	if _, err := os.Stat(avifLocalPath); err == nil {
		valid, fileSize := imageutils.IsValidFile(avifLocalPath)
		if valid {
			if err := c.UploadFile(avifLocalPath, avifRemotePath); err != nil {
				log.Printf("エラー: AVIFファイルのアップロードに失敗しました %s: %v", avifLocalPath, err)
				stats.AVIFFailed++
				return false
			}
			stats.AVIFSuccess++
			stats.UploadedFiles++
			log.Printf("AVIFファイルのアップロード成功: %s (サイズ: %d バイト)", avifRemotePath, fileSize)
			return true
		} else {
			log.Printf("警告: AVIFファイルが無効なためスキップします: %s", avifLocalPath)
			stats.AVIFFailed++
			stats.SkippedUploads++
		}
	}
	return false
}

// cleanupFiles は処理済みのファイルを削除します
func cleanupFiles(localPath, baseName string) {
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

// ProcessFileBatch はファイルのバッチを処理します
func (c *Client) ProcessFileBatch(files []string, tempDir string, stats *config.ConversionStats) error {
	for _, remoteFile := range files {
		if err := c.ProcessRemoteFile(remoteFile, tempDir, stats); err != nil {
			// エラーがあっても続行
			log.Printf("ファイル処理エラー [%s]: %v", remoteFile, err)
		}
	}
	return nil
}

// LogIntermediateStats は中間処理結果をログに出力します
func LogIntermediateStats(stats *config.ConversionStats, processed, total int) {
	log.Printf("=== 中間処理統計 (%d/%d ファイル) ===", processed, total)
	log.Printf("処理ファイル数: %d", stats.TotalProcessed)
	log.Printf("ダウンロード失敗: %d, 変換失敗: %d", stats.DownloadFailed, stats.ConvertFailed)
	log.Printf("WebP変換成功: %d, 失敗: %d", stats.WebPSuccess, stats.WebPFailed)
	log.Printf("AVIF変換成功: %d, 失敗: %d", stats.AVIFSuccess, stats.AVIFFailed)
	log.Printf("アップロード成功: %d, スキップ: %d", stats.UploadedFiles, stats.SkippedUploads)
	log.Printf("現在の処理時間: %s", time.Since(stats.StartTime))
}
