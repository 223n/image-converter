/*
Package config は設定ファイルの読み込みと設定値の管理を行います。
*/
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		Directory  string `yaml:"directory"`
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

// ConversionStats は変換統計情報を保持する構造体
type ConversionStats struct {
	TotalProcessed int
	DownloadFailed int
	ConvertFailed  int
	WebPSuccess    int
	WebPFailed     int
	AVIFSuccess    int
	AVIFFailed     int
	UploadedFiles  int
	SkippedUploads int
	StartTime      time.Time
}

// NewConversionStats は新しい統計情報構造体を作成します
func NewConversionStats() *ConversionStats {
	return &ConversionStats{
		StartTime: time.Now(),
	}
}

// グローバル変数
var (
	config              Config
	supportedExtensions map[string]bool
)

// LoadConfig は設定ファイルを読み込みます
func LoadConfig(configPath string) error {
	// configPathが相対パスの場合、実行ディレクトリからの相対パスとして解釈
	if !filepath.IsAbs(configPath) {
		// 現在の作業ディレクトリを取得
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("現在の作業ディレクトリの取得に失敗しました: %v", err)
		}
		configPath = filepath.Join(wd, configPath)
	}

	// ファイルが存在するか確認
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("設定ファイルが存在しません: %s", configPath)
	}

	// 設定ファイルを読み込む
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %v", err)
	}

	// デフォルト設定を適用
	config = DefaultConfig()

	// YAMLデータを構造体にアンマーシャル
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return fmt.Errorf("設定ファイルの解析に失敗しました: %v", err)
	}

	// 設定値の検証と調整
	validateConfig()

	// サポートされている拡張子をマップに変換
	supportedExtensions = make(map[string]bool)
	for _, ext := range config.Input.SupportedExtensions {
		supportedExtensions[strings.ToLower(ext)] = true
	}

	return nil
}

// validateConfig は設定値を検証し、必要に応じて調整します
func validateConfig() {
	// ワーカー数の検証（少なくとも1以上）
	if config.Conversion.Workers < 1 {
		config.Conversion.Workers = 1
	}

	// WebP品質の検証（0〜100の範囲）
	if config.Conversion.WebP.Quality < 0 {
		config.Conversion.WebP.Quality = 0
	} else if config.Conversion.WebP.Quality > 100 {
		config.Conversion.WebP.Quality = 100
	}

	// AVIF品質の検証（1〜63の範囲）
	if config.Conversion.AVIF.Quality < 1 {
		config.Conversion.AVIF.Quality = 1
	} else if config.Conversion.AVIF.Quality > 63 {
		config.Conversion.AVIF.Quality = 63
	}

	// AVIF速度の検証（0〜10の範囲）
	if config.Conversion.AVIF.Speed < 0 {
		config.Conversion.AVIF.Speed = 0
	} else if config.Conversion.AVIF.Speed > 10 {
		config.Conversion.AVIF.Speed = 10
	}

	// リモートタイムアウトが短すぎる場合は調整
	if config.Remote.Enabled && config.Remote.Timeout < 60 {
		config.Remote.Timeout = 60
	}
}

// GetConfig は現在の設定を返します
func GetConfig() Config {
	return config
}

// GetRemoteConfig はリモート設定を作成します
func GetRemoteConfig() *RemoteConfig {
	return &RemoteConfig{
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
}

// SetDryRun はドライランモードを設定します
func SetDryRun(enabled bool) {
	config.Mode.DryRun = enabled
}

// SetRemoteMode はリモートモードを設定します
func SetRemoteMode(enabled bool) {
	config.Remote.Enabled = enabled
}

// IsDryRun はドライランモードかどうかを返します
func IsDryRun() bool {
	return config.Mode.DryRun
}

// IsRemoteMode はリモートモードかどうかを返します
func IsRemoteMode() bool {
	return config.Remote.Enabled
}

// IsSupportedExtension は指定された拡張子がサポートされているかどうかを返します
func IsSupportedExtension(ext string) bool {
	return supportedExtensions[strings.ToLower(ext)]
}

// GetSupportedExtensions はサポートされている拡張子のリストを返します
func GetSupportedExtensions() []string {
	return config.Input.SupportedExtensions
}

// GetInputDirectory は入力ディレクトリのパスを返します
func GetInputDirectory() string {
	return config.Input.Directory
}

// GetWorkerCount はワーカー数を返します
func GetWorkerCount() int {
	return config.Conversion.Workers
}

// IsWebPEnabled はWebP変換が有効かどうかを返します
func IsWebPEnabled() bool {
	return config.Conversion.WebP.Enabled
}

// GetWebPQuality はWebP品質設定を返します
func GetWebPQuality() int {
	return config.Conversion.WebP.Quality
}

// IsAVIFEnabled はAVIF変換が有効かどうかを返します
func IsAVIFEnabled() bool {
	return config.Conversion.AVIF.Enabled
}

// GetAVIFQuality はAVIF品質設定を返します
func GetAVIFQuality() int {
	return config.Conversion.AVIF.Quality
}

// GetAVIFSpeed はAVIF速度設定を返します
func GetAVIFSpeed() int {
	return config.Conversion.AVIF.Speed
}

// IsFTPEnabled はFTPサーバーが有効かどうかを返します
func IsFTPEnabled() bool {
	return config.FTP.Enabled
}

// IsSSHEnabled はSSHサーバーが有効かどうかを返します
func IsSSHEnabled() bool {
	return config.SSH.Enabled
}
