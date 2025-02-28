// Package local は、ローカルファイルシステムでの画像変換処理を提供します。
package local

import (
	"fmt"
	"log"
	"time"

	"github.com/223n/image-converter/internal/config"
	"github.com/223n/image-converter/internal/utils"
)

// Service はローカルモードでの画像変換サービスを表します
type Service struct {
	config     *config.Config
	stats      *config.ConversionStats
	startTime  time.Time
	logManager *utils.LogManager
}

// NewService は新しいローカルサービスインスタンスを作成します
func NewService(cfg *config.Config, logManager *utils.LogManager) *Service {
	return &Service{
		config:     cfg,
		stats:      config.NewConversionStats(),
		startTime:  time.Now(),
		logManager: logManager,
	}
}

// Execute はローカル変換処理を実行します
func (s *Service) Execute() error {
	log.Printf("ローカルモードでの変換を開始します...")
	s.logManager.LogInfo("ローカルモードでの変換を開始します。設定: %s", s.config.Input.Directory)

	// ファイル検索
	finder := NewFileFinder(s.config)
	files, totalFiles, err := finder.FindFiles()
	if err != nil {
		return fmt.Errorf("ファイル検索に失敗しました: %w", err)
	}

	s.logManager.LogInfo("検索完了: %d個のファイルが見つかりました", totalFiles)

	// ドライランモードの場合
	if s.config.Mode.DryRun {
		s.logManager.LogInfo("ドライランモード: 変換は行われません")
		s.printFileList(files)
		return nil
	}

	// 処理実行
	processor := NewFileProcessor(s.config, s.stats, s.logManager)
	if err := processor.ProcessFiles(files, totalFiles); err != nil {
		return fmt.Errorf("ファイル処理に失敗しました: %w", err)
	}

	// 結果出力
	s.logSummary(totalFiles)
	return nil
}

// logSummary は変換結果のサマリーをログに出力します
func (s *Service) logSummary(totalFiles int) {
	s.logManager.LogInfo("=== 変換処理結果 ===")
	s.logManager.LogInfo("処理ファイル数: %d", totalFiles)
	s.logManager.LogInfo("WebP変換成功: %d, 失敗: %d", s.stats.WebPSuccess, s.stats.WebPFailed)
	s.logManager.LogInfo("AVIF変換成功: %d, 失敗: %d", s.stats.AVIFSuccess, s.stats.AVIFFailed)
	s.logManager.LogInfo("処理時間: %s", time.Since(s.startTime))
	s.logManager.LogInfo("=== 画像変換処理終了: %s ===", time.Now().Format("2006-01-02 15:04:05"))
}

// printFileList はドライランモードでファイルリストを表示します
func (s *Service) printFileList(files []string) {
	s.logManager.LogInfo("=== 変換対象ファイル ===")
	for i, file := range files {
		s.logManager.LogInfo("%d: %s", i+1, file)
	}
	s.logManager.LogInfo("合計: %d個のファイル", len(files))
}

// GetStats は現在の統計情報を返します
func (s *Service) GetStats() *config.ConversionStats {
	return s.stats
}
