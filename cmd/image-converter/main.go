/*
Package main は画像変換ツールのエントリーポイントです。

このツールは指定されたディレクトリ内のJPG、PNG、HEIC、HEIFなどの画像ファイルを
WebPとAVIFフォーマットに変換し、元のディレクトリ構造を維持したまま保存します。
FTPとSSHによるリモートアクセスもサポートしており、リモートWebサーバー上での
画像変換機能も実装しています。
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/image-converter/internal/config"
	"github.com/yourusername/image-converter/internal/local"
	"github.com/yourusername/image-converter/internal/remote"
	"github.com/yourusername/image-converter/internal/utils"
)

var (
	configPath string
	dryRun     bool
	remoteMode bool
	startTime  time.Time
)

func init() {
	flag.StringVar(&configPath, "config", "configs/config.yml", "設定ファイルのパス")
	flag.BoolVar(&dryRun, "dry-run", false, "ドライランモード（実際の変換は行わない）")
	flag.BoolVar(&remoteMode, "remote", false, "リモートモード（SSHで接続して変換）")

	// 処理開始時間を記録
	startTime = time.Now()
}

// main はプログラムのエントリーポイントです
func main() {
	// 初期化と設定の読み込み
	if err := initializeApplication(); err != nil {
		log.Fatalf("初期化に失敗しました: %v", err)
	}

	// リモートモードの処理
	if config.GetConfig().Remote.Enabled {
		if err := executeRemoteMode(); err != nil {
			log.Fatalf("リモート変換に失敗しました: %v", err)
		}
		return
	}

	// ローカルモードの処理
	if err := executeLocalMode(); err != nil {
		log.Fatalf("ローカル変換に失敗しました: %v", err)
	}
}

// initializeApplication はアプリケーションの初期化と設定を行います
func initializeApplication() error {
	// コマンドライン引数の解析
	flag.Parse()

	// 設定ファイルを読み込む
	if err := config.LoadConfig(configPath); err != nil {
		return err
	}

	// コマンドラインオプションが設定されていればYAML設定よりも優先
	if dryRun {
		config.SetDryRun(true)
	}

	if remoteMode {
		config.SetRemoteMode(true)
	}

	// ログファイル名に開始日時を含める
	logFileName := utils.GetLogFileName(startTime)

	// ログ設定を適用
	utils.SetupLogger(logFileName)

	// 開始ログを出力
	utils.LogStartupInfo(configPath)

	return nil
}

// executeRemoteMode はリモートモード処理を実行します
func executeRemoteMode() error {
	log.Printf("リモートモードで実行中 - ホスト: %s", config.GetConfig().Remote.Host)
	fmt.Printf("リモートモードで実行中 - ホスト: %s\n", config.GetConfig().Remote.Host)

	// リモート変換の実行
	remoteService := remote.NewService()
	if err := remoteService.Execute(); err != nil {
		return fmt.Errorf("リモート変換に失敗しました: %v", err)
	}

	log.Println("リモート変換が完了しました")
	log.Printf("処理時間: %s", time.Since(startTime))
	log.Printf("=== 画像変換処理終了: %s ===", time.Now().Format("2006-01-02 15:04:05"))

	return nil
}

// executeLocalMode はローカルモード処理を実行します
func executeLocalMode() error {
	// ローカル変換サービスを作成して実行
	localService := local.NewService()
	if err := localService.Execute(); err != nil {
		return fmt.Errorf("ローカル変換に失敗しました: %v", err)
	}

	return nil
}
