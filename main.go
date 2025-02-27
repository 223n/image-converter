package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Kagami/go-avif"
	"github.com/chai2010/webp"
	"github.com/jdeng/goheif"
	"gopkg.in/yaml.v3"
)

// Config はYAML設定ファイルの構造を表します
type Config struct {
	Remote struct {
		Enabled     bool   `yaml:"enabled"`
		Host        string `yaml:"host"`
		Port        int    `yaml:"port"`
		User        string `yaml:"user"`
		KeyPath     string `yaml:"key_path"`
		KnownHosts  string `yaml:"known_hosts"`
		RemotePath  string `yaml:"remote_path"`
		UseSSHAgent bool   `yaml:"use_ssh_agent"`
		Timeout     int    `yaml:"timeout"`
	} `yaml:"remote"`

	Mode struct {
		DryRun bool `yaml:"dry_run"`
	} `yaml:"mode"`

	Input struct {
		Directory           string   `yaml:"directory"`
		SupportedExtensions []string `yaml:"supported_extensions"`
	} `yaml:"input"`

	Conversion struct {
		Workers int `yaml:"workers"`
		WebP    struct {
			Enabled          bool `yaml:"enabled"`
			Quality          int  `yaml:"quality"`
			CompressionLevel int  `yaml:"compression_level"`
		} `yaml:"webp"`
		AVIF struct {
			Enabled  bool `yaml:"enabled"`
			Quality  int  `yaml:"quality"`
			Speed    int  `yaml:"speed"`
			Lossless bool `yaml:"lossless"`
		} `yaml:"avif"`
	} `yaml:"conversion"`

	FTP struct {
		Enabled bool `yaml:"enabled"`
		Port    int  `yaml:"port"`
		User    struct {
			Name     string `yaml:"name"`
			Password string `yaml:"password"`
		} `yaml:"user"`
		Passive struct {
			Enabled   bool   `yaml:"enabled"`
			PortRange string `yaml:"port_range"`
		} `yaml:"passive"`
	} `yaml:"ftp"`

	SSH struct {
		Enabled bool `yaml:"enabled"`
		Port    int  `yaml:"port"`
		Auth    struct {
			PasswordAuth bool   `yaml:"password_auth"`
			PubkeyAuth   bool   `yaml:"pubkey_auth"`
			AuthKeysFile string `yaml:"auth_keys_file"`
		} `yaml:"auth"`
	} `yaml:"ssh"`

	Logging struct {
		Level      string `yaml:"level"`
		File       string `yaml:"file"`
		MaxSize    int    `yaml:"max_size"`
		MaxBackups int    `yaml:"max_backups"`
		MaxAge     int    `yaml:"max_age"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"logging"`
}

// RemoteConfig はリモートサーバーの接続設定
type RemoteConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	User        string `yaml:"user"`
	KeyPath     string `yaml:"key_path"`
	KnownHosts  string `yaml:"known_hosts"`
	RemotePath  string `yaml:"remote_path"`
	UseSSHAgent bool   `yaml:"use_ssh_agent"`
	Timeout     int    `yaml:"timeout"`
}

var (
	configPath string
	config     Config
	dryRun     bool
	remoteMode bool
)

func init() {
	flag.StringVar(&configPath, "config", "config.yml", "設定ファイルのパス")
	flag.BoolVar(&dryRun, "dry-run", false, "ドライランモード（実際の変換は行わない）")
	flag.BoolVar(&remoteMode, "remote", false, "リモートモード（SSHで接続して変換）")
	flag.Parse()
}

// 設定ファイルを読み込む
func loadConfig() error {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %v", err)
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return fmt.Errorf("設定ファイルの解析に失敗しました: %v", err)
	}

	// サポートされている拡張子をマップに変換
	supportedExtensions = make(map[string]bool)
	for _, ext := range config.Input.SupportedExtensions {
		supportedExtensions[ext] = true
	}

	return nil
}

// 変換対象のファイル拡張子
var supportedExtensions map[string]bool

func main() {
	// 設定ファイルを読み込む
	if err := loadConfig(); err != nil {
		log.Fatalf("設定ファイルのロードに失敗しました: %v", err)
	}

	// コマンドラインオプションが設定されていればYAML設定よりも優先
	if dryRun {
		config.Mode.DryRun = true
	}

	if remoteMode {
		config.Remote.Enabled = true
	}

	// ロガーの設定
	setupLogger()

	// ドライランモードの場合は通知
	if config.Mode.DryRun {
		log.Println("ドライランモードで実行中 - 実際の変換は行われません")
		fmt.Println("ドライランモード: 実際の変換は行われません")
	}

	// リモートモードの処理
	if config.Remote.Enabled {
		log.Printf("リモートモードで実行中 - ホスト: %s", config.Remote.Host)
		fmt.Printf("リモートモードで実行中 - ホスト: %s\n", config.Remote.Host)

		// RemoteConfig構造体に変換して渡す
		remoteConfig := &RemoteConfig{
			Enabled:     config.Remote.Enabled,
			Host:        config.Remote.Host,
			Port:        config.Remote.Port,
			User:        config.Remote.User,
			KeyPath:     config.Remote.KeyPath,
			KnownHosts:  config.Remote.KnownHosts,
			RemotePath:  config.Remote.RemotePath,
			UseSSHAgent: config.Remote.UseSSHAgent,
			Timeout:     config.Remote.Timeout,
		}

		// リモート変換の実行
		if err := ConvertRemoteImages(remoteConfig, config.Input.SupportedExtensions); err != nil {
			log.Fatalf("リモート変換に失敗しました: %v", err)
		}

		log.Println("リモート変換が完了しました")
		return
	}

	// 変換対象ファイルの収集
	var filesToConvert []string
	err := filepath.Walk(config.Input.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if supportedExtensions[ext] {
			filesToConvert = append(filesToConvert, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("ファイル検索に失敗しました: %v", err)
	}

	totalFiles := len(filesToConvert)
	log.Printf("変換対象ファイル数: %d\n", totalFiles)

	// 進捗トラッカーを作成
	tracker := NewMultiProgressTracker(totalFiles, "変換処理")

	// ワーカープールを使用した並列処理
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.Conversion.Workers)
	for _, file := range filesToConvert {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(file string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			err := convertImage(file)
			if err != nil {
				log.Printf("変換エラー [%s]: %v", file, err)
				tracker.IncrementFailed()
			} else {
				tracker.IncrementSuccess()
			}
		}(file)
	}
	wg.Wait()

	// 進捗トラッカーを完了
	tracker.Complete()

	log.Println("変換完了")

	// FTPとSSHサーバーの起動（設定ファイルで有効な場合）
	if config.FTP.Enabled {
		startFTPServer()
	}

	if config.SSH.Enabled {
		startSSHServer()
	}

	// FTPまたはSSHが有効な場合、プログラムを終了しないように待機
	if config.FTP.Enabled || config.SSH.Enabled {
		fmt.Println("サーバーが稼働中です。Ctrl+Cで終了してください。")
		select {} // 無限待機
	}
}

// ロガーの設定
func setupLogger() {
	// 基本的なログ設定
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// ログファイルが指定されている場合
	if config.Logging.File != "" {
		logFile, err := os.OpenFile(config.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("ログファイルのオープンに失敗しました: %v", err)
		}
		// 標準出力とファイルの両方にログを書き込む
		// 本格的なロガーライブラリを使用する場合は、ここでzap、logrusなどの設定を行う
		log.SetOutput(logFile)
	}
}

// 画像を読み込んでデコード
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

// 画像をWebPとAVIFに変換
func convertImage(filePath string) error {
	img, err := loadImage(filePath)
	if err != nil {
		return err
	}

	// 元のファイル名を取得（拡張子を除く）
	baseFileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// 元のファイルと同じディレクトリに保存するためのパスを作成
	dir := filepath.Dir(filePath)

	// WebP変換（config.Conversion.WebP.Enabledがtrueの場合）
	if config.Conversion.WebP.Enabled {
		webpPath := filepath.Join(dir, baseFileName+".webp")

		// 代替手段を優先的に使用（より信頼性が高い）
		if config.Mode.DryRun {
			// ドライランモードではスキップして成功としてログ出力
			log.Printf("ドライラン: WebP変換のスキップ")
			return nil
		}
		if err := saveWebP(img, webpPath); err != nil {
			log.Printf("WebP変換に失敗しました: %v", err)
		} else {
			log.Printf("WebP変換成功: %s", webpPath)
		}
	}

	// AVIF変換（config.Conversion.AVIF.Enabledがtrueの場合）
	if config.Conversion.AVIF.Enabled {
		avifPath := filepath.Join(dir, baseFileName+".avif")

		// ドライランモードの場合は変換をスキップしてログだけ出力
		if config.Mode.DryRun {
			log.Printf("ドライラン: AVIF変換対象: %s -> %s", filePath, avifPath)
		} else if err := saveAVIF(img, avifPath); err != nil {
			log.Printf("AVIF変換に失敗しました: %v", err)
		} else {
			log.Printf("AVIF変換成功: %s", avifPath)
		}
	}

	log.Printf("変換成功: %s", filePath)
	return nil
}

// 画像をWebPとして保存
func saveWebP(img image.Image, outputPath string) error {
	// 優先順位:
	// 1. cwebp コマンド
	// 2. chai2010/webp ライブラリ

	// cwebpが利用可能か確認
	if _, err := exec.LookPath("cwebp"); err == nil {
		// cwebpコマンドを使用
		return saveWebPAlternative(img, outputPath, config.Conversion.WebP.Quality)
	}

	// cwebpが使用できない場合はchai2010/webpライブラリを使用
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("出力ファイルの作成に失敗しました: %v", err)
	}
	defer output.Close()

	opts := &webp.Options{
		Lossless: false,
		Quality:  float32(config.Conversion.WebP.Quality),
	}
	if err := webp.Encode(output, img, opts); err != nil {
		return fmt.Errorf("WebPエンコードに失敗しました: %v", err)
	}

	return nil
}

// 画像をAVIFとして保存
func saveAVIF(img image.Image, outputPath string) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	// AVIFエンコードオプションの設定
	options := &avif.Options{
		// Quality: 品質 (0-100)
		Quality: config.Conversion.AVIF.Quality,
		// Speed: 処理速度 (0-10, 値が大きいほど速いが品質は下がる)
		Speed: config.Conversion.AVIF.Speed,
	}

	// AVIF形式で保存
	if err := avif.Encode(output, img, options); err != nil {
		return err
	}

	return nil
}

// FTPサーバーを起動
func startFTPServer() {
	// Pure-FTPdを使用する例（事前にインストールが必要）
	log.Printf("FTPサーバーを起動中（ポート: %d）...\n", config.FTP.Port)

	// FTPコマンドの引数を構築
	args := []string{
		"--bind", fmt.Sprintf("0.0.0.0,%d", config.FTP.Port),
		"--chroot",     // ユーザーをホームディレクトリに制限
		"--createhome", // ホームディレクトリを自動作成
	}

	// 匿名ログインを無効化
	args = append(args, "--noanonymous")

	// パッシブモードが有効な場合
	if config.FTP.Passive.Enabled {
		args = append(args, "--passiveportrange", config.FTP.Passive.PortRange)
	}

	// 認証データベースの設定
	args = append(args, "--login", "puredb:/etc/pure-ftpd/pureftpd.pdb")

	cmd := exec.Command("pure-ftpd", args...)

	// FTPユーザーとパスワードの設定
	// 実際の環境では、pure-pw コマンドなどを使用してユーザーを追加する処理が必要

	err := cmd.Start()
	if err != nil {
		log.Printf("FTPサーバーの起動に失敗しました: %v\nシステムにpure-ftpdがインストールされていることを確認してください。", err)
		return
	}

	log.Printf("FTPサーバーを起動しました（PID: %d）\n", cmd.Process.Pid)
}

// SSHサーバーを起動
func startSSHServer() {
	// OpenSSHを使用する例（事前にインストールが必要）
	log.Printf("SSHサーバーを起動中（ポート: %d）...\n", config.SSH.Port)

	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("ホームディレクトリの取得に失敗しました: %v", err)
		return
	}

	// SSH設定ディレクトリの作成
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		log.Printf("SSH設定ディレクトリの作成に失敗しました: %v", err)
		return
	}

	// ssh-agent から登録済みの鍵を取得して authorized_keys に追加
	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	if err := setupSSHKeys(authorizedKeysPath); err != nil {
		log.Printf("SSH鍵の設定に失敗しました: %v", err)
		// 続行する（エラーがあっても他の認証方法が使える可能性あり）
	}

	// SSHコマンドの引数を構築
	args := []string{
		"-D", // フォアグラウンドで実行
		"-p", fmt.Sprintf("%d", config.SSH.Port),
	}

	// パスワード認証の設定
	if config.SSH.Auth.PasswordAuth {
		args = append(args, "-o", "PasswordAuthentication=yes")
	} else {
		args = append(args, "-o", "PasswordAuthentication=no")
	}

	// 公開鍵認証の設定（常に有効）
	args = append(args, "-o", "PubkeyAuthentication=yes")
	args = append(args, "-o", fmt.Sprintf("AuthorizedKeysFile=%s", authorizedKeysPath))

	cmd := exec.Command("sshd", args...)

	err = cmd.Start()
	if err != nil {
		log.Printf("SSHサーバーの起動に失敗しました: %v\nシステムにOpenSSHがインストールされていることを確認してください。", err)
		return
	}

	log.Printf("SSHサーバーを起動しました（PID: %d）\n", cmd.Process.Pid)
}
