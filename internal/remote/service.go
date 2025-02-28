package remote

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/223n/image-converter/internal/config"
	"github.com/223n/image-converter/internal/utils"
)

// Service はリモート変換サービスを表します
type Service struct {
	config *config.RemoteConfig
}

// NewService は新しいリモート変換サービスを作成します
func NewService() *Service {
	return &Service{
		config: config.GetRemoteConfig(),
	}
}

// Execute はリモート変換を実行します
func (s *Service) Execute() error {
	// 設定の検証
	if err := s.validateConfig(); err != nil {
		return err
	}

	// ログ設定
	logFileName, logFile := s.setupLogging()
	if logFile != nil {
		defer logFile.Close()
	}

	s.logStartInfo()

	// SSHクライアント作成
	client, err := NewClient(s.config)
	if err != nil {
		s.logFatalError("SSHクライアントの作成に失敗しました", err)
		return fmt.Errorf("SSHクライアントの作成に失敗しました: %w", err)
	}
	defer client.Close()

	// リモートファイル検索
	imageFiles, totalFiles, err := s.findRemoteImages(client)
	if err != nil {
		return err
	}

	// 一時ディレクトリの準備
	tempDir, err := s.prepareTempDirectory()
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// 統計情報の初期化
	stats := config.NewConversionStats()

	// バッチ処理
	if err := s.processBatches(client, imageFiles, totalFiles, tempDir, stats); err != nil {
		return err
	}

	// 結果の出力
	s.logConversionResults(stats, totalFiles, logFileName)

	return nil
}

// validateConfig は設定を検証します
func (s *Service) validateConfig() error {
	if !s.config.Enabled {
		return fmt.Errorf("リモート変換が無効です")
	}

	// タイムアウト設定を増やして、より長い接続時間を可能に
	if s.config.Timeout < 60 {
		log.Printf("警告: リモート接続タイムアウトが短すぎます。60秒に設定します: %d -> 60", s.config.Timeout)
		s.config.Timeout = 60
	}

	return nil
}

// setupLogging はリモート用ログファイルを設定します
func (s *Service) setupLogging() (string, *os.File) {
	logFileName := fmt.Sprintf("remote-converter_%s.log", time.Now().Format("20060102_150405"))

	// 設定からログディレクトリを取得（デフォルトは "logs"）
	cfg := config.GetConfig()
	logsDir := "logs"
	if cfg.Logging.Directory != "" {
		logsDir = cfg.Logging.Directory
	}

	// ログディレクトリを作成
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("警告: ログディレクトリの作成に失敗しました: %v - 標準出力にログを出力します", err)
		return logFileName, nil
	}

	// ログファイルのパスを設定
	logFilePath := filepath.Join(logsDir, logFileName)

	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Printf("警告: ログファイルの作成に失敗しました: %v", err)
		return logFileName, nil
	}

	log.SetOutput(logFile)
	return logFileName, logFile
}

// logStartInfo はリモート処理開始情報をログに記録します
func (s *Service) logStartInfo() {
	log.Printf("=== リモート画像変換処理開始: %s ===", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("リモートサーバー上の画像変換を開始します...")
	log.Printf("対象サーバー: %s:%d, ユーザー: %s, パス: %s",
		s.config.Host, s.config.Port, s.config.User, s.config.RemotePath)
}

// logFatalError は致命的なエラーをログに記録します
func (s *Service) logFatalError(message string, err error) {
	log.Printf("致命的エラー: %s: %v", message, err)
}

// findRemoteImages はリモートサーバー上の画像ファイルを検索します
func (s *Service) findRemoteImages(client *Client) ([]string, int, error) {
	imageFiles, err := client.FindRemoteImages(config.GetSupportedExtensions())
	if err != nil {
		s.logFatalError("リモート画像の検索に失敗しました", err)
		return nil, 0, fmt.Errorf("リモート画像の検索に失敗しました: %w", err)
	}

	totalFiles := len(imageFiles)
	log.Printf("リモートサーバーで変換対象の画像: %d個", totalFiles)

	// 一時停止して接続を確保
	log.Printf("処理を開始する前に5秒間待機します...")
	time.Sleep(5 * time.Second)

	return imageFiles, totalFiles, nil
}

// prepareTempDirectory は一時ディレクトリを作成します
func (s *Service) prepareTempDirectory() (string, error) {
	tempDir, err := os.MkdirTemp("", "remote-images-")
	if err != nil {
		s.logFatalError("一時ディレクトリの作成に失敗しました", err)
		return "", fmt.Errorf("一時ディレクトリの作成に失敗しました: %w", err)
	}
	return tempDir, nil
}

// processBatches はファイルをバッチ処理します
func (s *Service) processBatches(client *Client, imageFiles []string, totalFiles int, tempDir string, stats *config.ConversionStats) error {
	// 進捗トラッカーを作成
	tracker := utils.NewMultiProgressTracker(totalFiles, "リモート変換")

	// バッチサイズを設定（メモリ使用量削減のため小さいサイズに変更）
	const batchSize = 10
	log.Printf("バッチ処理を使用します: %d個のファイルごとに処理", batchSize)

	// ファイルをバッチごとに処理
	for i := 0; i < len(imageFiles); i += batchSize {
		end := i + batchSize
		if end > len(imageFiles) {
			end = len(imageFiles)
		}

		log.Printf("バッチ処理: %d - %d / %d ファイル", i+1, end, totalFiles)

		// 各バッチの間で休止してSSH接続を安定させる
		if i > 0 {
			log.Printf("バッチ間休止: 5秒間待機...")
			time.Sleep(5 * time.Second)
		}

		// このバッチのファイルを処理
		if err := s.processFileBatch(client, imageFiles[i:end], tempDir, tracker, stats); err != nil {
			return err
		}

		// 中間統計情報をログに出力
		LogIntermediateStats(stats, end, totalFiles)

		// メモリ使用状況を出力しガベージコレクションを強制実行
		s.performMemoryManagement()
	}

	// 進捗トラッカーを完了
	tracker.Complete()

	return nil
}

// performMemoryManagement はメモリ使用状況の出力とガベージコレクションを実行します
func (s *Service) performMemoryManagement() {
	// 明示的にガベージコレクションを呼び出す
	runtime.GC()

	// メモリ使用状況を出力
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("メモリ使用量: Alloc=%v MiB, Sys=%v MiB", m.Alloc/1024/1024, m.Sys/1024/1024)
}

// processFileBatch はファイルのバッチを処理します
func (s *Service) processFileBatch(client *Client, files []string, tempDir string, tracker *utils.MultiProgressTracker, stats *config.ConversionStats) error {
	for _, remoteFile := range files {
		if err := s.processFile(client, remoteFile, tempDir, tracker, stats); err != nil {
			// エラーがあっても続行
			log.Printf("ファイル処理エラー [%s]: %v", remoteFile, err)
		}
	}
	return nil
}

// processFile は単一のリモートファイルを処理します
func (s *Service) processFile(client *Client, remoteFile, tempDir string, tracker *utils.MultiProgressTracker, stats *config.ConversionStats) error {
	err := client.ProcessRemoteFile(remoteFile, tempDir, stats)

	if err != nil {
		tracker.IncrementFailed()
		return err
	}

	tracker.IncrementSuccess()
	return nil
}

// logConversionResults はリモート変換結果をログに出力します
func (s *Service) logConversionResults(stats *config.ConversionStats, _ int, logFileName string) {
	log.Println("=== 変換処理結果 ===")
	log.Printf("処理ファイル数: %d", stats.TotalProcessed)
	log.Printf("ダウンロード失敗: %d, 変換失敗: %d", stats.DownloadFailed, stats.ConvertFailed)
	log.Printf("WebP変換成功: %d, 失敗: %d", stats.WebPSuccess, stats.WebPFailed)
	log.Printf("AVIF変換成功: %d, 失敗: %d", stats.AVIFSuccess, stats.AVIFFailed)
	log.Printf("アップロード成功: %d, スキップ: %d", stats.UploadedFiles, stats.SkippedUploads)
	log.Printf("処理時間: %s", time.Since(stats.StartTime))
	log.Printf("=== 画像変換処理終了: %s ===", time.Now().Format("2006-01-02 15:04:05"))

	fmt.Printf("変換処理の詳細ログは logs/%s に保存されました\n", logFileName)
}
