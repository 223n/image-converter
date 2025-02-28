package local

import (
	"fmt"
	"sync"
	"time"

	"github.com/223n/image-converter/internal/config"
	"github.com/223n/image-converter/internal/converter"
	"github.com/223n/image-converter/internal/utils"
)

// FileProcessor はローカルファイルの処理を担当します
type FileProcessor struct {
	config     *config.Config // ポインタとして設定
	stats      *config.ConversionStats
	converter  *converter.ImageConverter
	logManager *utils.LogManager
}

// NewFileProcessor は新しいファイル処理インスタンスを作成します
func NewFileProcessor(cfg *config.Config, stats *config.ConversionStats, logManager *utils.LogManager) *FileProcessor {
	return &FileProcessor{
		config:     cfg,
		stats:      stats,
		converter:  converter.NewImageConverter(cfg, logManager),
		logManager: logManager,
	}
}

// ProcessFiles は複数のファイルを並行処理します
func (p *FileProcessor) ProcessFiles(files []string, totalFiles int) error {
	// 進捗トラッカーを作成
	tracker := utils.NewMultiProgressTracker(totalFiles, "変換処理")

	// ワーカープールを使用した並列処理
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, p.config.Conversion.Workers)

	// エラー収集用のチャネル
	errorCh := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(file string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := p.processFile(file, tracker); err != nil {
				errorCh <- fmt.Errorf("ファイル %s の処理に失敗しました: %v", file, err)
			}
		}(file)
	}

	// すべてのワーカーの終了を待機
	wg.Wait()
	close(errorCh)

	// 進捗トラッカーを完了
	tracker.Complete()

	// エラーがあれば最初のものを返す
	for err := range errorCh {
		return err
	}

	return nil
}

// processFile は単一ファイルの処理を行います
func (p *FileProcessor) processFile(file string, tracker *utils.MultiProgressTracker) error {
	// ファイル処理の開始時間を記録
	startTime := time.Now()

	// 変換処理の実行
	result, err := p.converter.Convert(file)
	if err != nil {
		p.logManager.LogError("変換エラー [%s]: %v", file, err)
		tracker.IncrementFailed()
		return err
	}

	// 統計情報の更新
	p.updateStats(result)

	// 処理時間をログに記録
	p.logManager.LogInfo("ファイル処理完了 [%s]: 所要時間 %v", file, time.Since(startTime))

	// 成功としてカウント
	p.stats.TotalProcessed++
	tracker.IncrementSuccess()

	return nil
}

// updateStats は変換結果に基づいて統計情報を更新します
func (p *FileProcessor) updateStats(result *converter.ConversionResult) {
	if result.WebPSuccess {
		p.stats.WebPSuccess++
		p.logManager.LogInfo("WebP変換成功: %s (サイズ: %d バイト)", result.WebPPath, result.WebPSize)
	} else if result.WebPAttempted {
		p.stats.WebPFailed++
		p.logManager.LogWarning("WebP変換失敗: %s", result.WebPPath)
	}

	if result.AVIFSuccess {
		p.stats.AVIFSuccess++
		p.logManager.LogInfo("AVIF変換成功: %s (サイズ: %d バイト)", result.AVIFPath, result.AVIFSize)
	} else if result.AVIFAttempted {
		p.stats.AVIFFailed++
		p.logManager.LogWarning("AVIF変換失敗: %s", result.AVIFPath)
	}
}
