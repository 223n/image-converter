/*
Package utils はアプリケーション全体で使用される共通ユーティリティを提供します。
*/
package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressBar はコンソールに進捗バーを表示するための構造体です
type ProgressBar struct {
	total       int
	current     int
	width       int
	description string
	mu          sync.Mutex
	startTime   time.Time
	lastUpdate  time.Time
	isDone      bool
}

// NewProgressBar は新しい進捗バーを作成します
func NewProgressBar(total int, description string) *ProgressBar {
	return &ProgressBar{
		total:       total,
		current:     0,
		width:       50, // バーの幅
		description: description,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
	}
}

// Increment は進捗バーを指定されたステップ数だけ増加させます
func (p *ProgressBar) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current++
	now := time.Now()

	// 更新頻度を制限（100msに1回まで）
	if now.Sub(p.lastUpdate) < 100*time.Millisecond && p.current < p.total {
		return
	}
	p.lastUpdate = now

	p.printProgress()
}

// IncrementBy は進捗バーを指定されたステップ数だけ増加させます
func (p *ProgressBar) IncrementBy(steps int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current += steps
	if p.current > p.total {
		p.current = p.total
	}

	now := time.Now()

	// 更新頻度を制限（100msに1回まで）
	if now.Sub(p.lastUpdate) < 100*time.Millisecond && p.current < p.total {
		return
	}
	p.lastUpdate = now

	p.printProgress()
}

// SetTotal は進捗バーの合計値を設定します
func (p *ProgressBar) SetTotal(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = total
}

// Complete は進捗バーを完了状態にします
func (p *ProgressBar) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isDone {
		return
	}

	p.current = p.total
	p.printProgress()
	fmt.Println() // 進捗バーの下に改行を追加

	duration := time.Since(p.startTime)
	fmt.Printf("%s: 完了 (所要時間: %s)\n", p.description, FormatDuration(duration))
	p.isDone = true
}

// printProgress は現在の進捗状況を表示します
func (p *ProgressBar) printProgress() {
	percent := float64(p.current) / float64(p.total)
	if percent > 1.0 {
		percent = 1.0
	}

	// 進捗バーに表示する文字数を計算
	filled := int(float64(p.width) * percent)
	if filled > p.width {
		filled = p.width
	}

	// 経過時間と推定残り時間を計算
	elapsed := time.Since(p.startTime)
	var eta time.Duration
	if percent > 0 {
		eta = time.Duration(float64(elapsed) / percent * (1 - percent))
	}

	// 進捗バーを構築
	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	// ステータス行を出力（\rで行頭に戻る）
	fmt.Printf("\r%s: [%s] %3.0f%% (%d/%d) 経過: %s 残り: %s",
		p.description, bar, percent*100, p.current, p.total,
		FormatDuration(elapsed), FormatDuration(eta))
}

// FormatDuration は時間を見やすい形式にフォーマットします
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// MultiProgressTracker は複数の処理の進捗を追跡する構造体です
type MultiProgressTracker struct {
	totalFiles  int
	processed   int
	succeeded   int
	failed      int
	skipped     int
	progressBar *ProgressBar
	mu          sync.Mutex
}

// NewMultiProgressTracker は新しい進捗トラッカーを作成します
func NewMultiProgressTracker(totalFiles int, description string) *MultiProgressTracker {
	return &MultiProgressTracker{
		totalFiles:  totalFiles,
		progressBar: NewProgressBar(totalFiles, description),
	}
}

// IncrementSuccess は成功したファイルの数を増やします
func (m *MultiProgressTracker) IncrementSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processed++
	m.succeeded++
	m.progressBar.Increment()
}

// IncrementFailed は失敗したファイルの数を増やします
func (m *MultiProgressTracker) IncrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processed++
	m.failed++
	m.progressBar.Increment()
}

// IncrementSkipped はスキップされたファイルの数を増やします
func (m *MultiProgressTracker) IncrementSkipped() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processed++
	m.skipped++
	m.progressBar.Increment()
}

// Complete は処理を完了し、最終的な統計情報を表示します
func (m *MultiProgressTracker) Complete() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.progressBar.Complete()
	fmt.Printf("処理結果: 成功: %d, 失敗: %d, スキップ: %d, 合計: %d\n",
		m.succeeded, m.failed, m.skipped, m.totalFiles)
}

// GetStats は現在の統計情報を返します
func (m *MultiProgressTracker) GetStats() (int, int, int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.processed, m.succeeded, m.failed, m.skipped
}
