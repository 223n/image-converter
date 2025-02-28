package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/223n/image-converter/internal/config"
)

// LogLevel は利用可能なログレベルを定義します
type LogLevel int

const (
	// LogLevelDebug はデバッグログレベルです
	LogLevelDebug LogLevel = iota
	// LogLevelInfo は情報ログレベルです
	LogLevelInfo
	// LogLevelWarn は警告ログレベルです
	LogLevelWarn
	// LogLevelError はエラーログレベルです
	LogLevelError
	// LogLevelFatal は致命的エラーログレベルです
	LogLevelFatal
)

// String はLogLevelを文字列に変換します
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogManager はログ管理機能を提供します
type LogManager struct {
	level LogLevel
}

// NewLogManager は新しいLogManagerインスタンスを作成します
func NewLogManager() *LogManager {
	cfg := config.GetConfig()
	return &LogManager{
		level: stringToLogLevel(cfg.Logging.Level),
	}
}

// LogInfo は情報メッセージをログに出力します
func (lm *LogManager) LogInfo(format string, args ...interface{}) {
	lm.logWithLevel(LogLevelInfo, format, args...)
}

// LogWarning は警告メッセージをログに出力します
func (lm *LogManager) LogWarning(format string, args ...interface{}) {
	lm.logWithLevel(LogLevelWarn, format, args...)
}

// LogError はエラーメッセージをログに出力します
func (lm *LogManager) LogError(format string, args ...interface{}) {
	lm.logWithLevel(LogLevelError, format, args...)
}

// LogDebug はデバッグメッセージをログに出力します
func (lm *LogManager) LogDebug(format string, args ...interface{}) {
	lm.logWithLevel(LogLevelDebug, format, args...)
}

// logWithLevel は指定されたレベルでメッセージをログに出力します
func (lm *LogManager) logWithLevel(level LogLevel, format string, args ...interface{}) {
	// 設定されたレベル以上の場合のみログを出力
	if level >= lm.level {
		message := fmt.Sprintf(format, args...)
		log.Printf("[%s] %s", level.String(), message)
	}
}

// logFile はログファイルへの参照を保持します
var logFile *os.File

// SetupLogger はロガーを設定します
func SetupLogger(logFileName string) {
	// 基本的なログ設定
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 設定からログディレクトリを取得（デフォルトは "logs"）
	cfg := config.GetConfig()
	logsDir := "logs"
	if cfg.Logging.Directory != "" {
		logsDir = cfg.Logging.Directory
	}

	// ログディレクトリを作成
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("警告: ログディレクトリの作成に失敗しました: %v - 標準出力にログを出力します", err)
		return
	}

	// 出力ログファイル名の設定
	outputLogFile := logFileName
	if cfg.Logging.File != "" {
		outputLogFile = cfg.Logging.File
	}

	// ログファイルのパスをlogsディレクトリ内に設定
	outputLogFile = filepath.Join(logsDir, filepath.Base(outputLogFile))

	// ログファイルを作成
	var err error
	logFile, err = os.Create(outputLogFile)
	if err != nil {
		log.Printf("警告: ログファイルの作成に失敗しました: %v - 標準出力にログを出力します", err)
		return
	}

	// 標準出力とファイルの両方にログを書き込む
	log.SetOutput(logFile)

	fmt.Printf("ログファイル: %s\n", outputLogFile)
}

// CloseLogger はロガーのリソースを解放します
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}

// GetLogFileName は日時を含むログファイル名を生成します
func GetLogFileName(timestamp time.Time) string {
	return fmt.Sprintf("image-converter_%s.log", timestamp.Format("20060102_150405"))
}

// LogStartupInfo は起動情報をログに出力します
func LogStartupInfo(configPath string) {
	log.Printf("=== 画像変換処理開始: %s ===", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("設定ファイル: %s", configPath)

	cfg := config.GetConfig()
	// ドライランモードの場合は通知
	if cfg.Mode.DryRun {
		log.Println("ドライランモードで実行中 - 実際の変換は行われません")
		fmt.Println("ドライランモード: 実際の変換は行われません")
	}
}

// LogDebug はデバッグメッセージをログに出力します
func LogDebug(format string, args ...interface{}) {
	logWithLevel(LogLevelDebug, format, args...)
}

// LogInfo は情報メッセージをログに出力します
func LogInfo(format string, args ...interface{}) {
	logWithLevel(LogLevelInfo, format, args...)
}

// LogWarn は警告メッセージをログに出力します
func LogWarn(format string, args ...interface{}) {
	logWithLevel(LogLevelWarn, format, args...)
}

// LogError はエラーメッセージをログに出力します
func LogError(format string, args ...interface{}) {
	logWithLevel(LogLevelError, format, args...)
}

// LogFatal は致命的エラーメッセージをログに出力し、プログラムを終了します
func LogFatal(format string, args ...interface{}) {
	logWithLevel(LogLevelFatal, format, args...)
	os.Exit(1)
}

// logWithLevel は指定されたレベルでメッセージをログに出力します
func logWithLevel(level LogLevel, format string, args ...interface{}) {
	cfg := config.GetConfig()
	configLevel := stringToLogLevel(cfg.Logging.Level)

	// 設定されたレベル以上の場合のみログを出力
	if level >= configLevel {
		message := fmt.Sprintf(format, args...)
		log.Printf("[%s] %s", level.String(), message)
	}
}

// stringToLogLevel は文字列をLogLevelに変換します
func stringToLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error", "err":
		return LogLevelError
	case "fatal":
		return LogLevelFatal
	default:
		return LogLevelInfo // デフォルトはInfo
	}
}
