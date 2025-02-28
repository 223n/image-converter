package config

// DefaultConfig はデフォルトの設定を返します
func DefaultConfig() Config {
	config := Config{}

	// リモート設定のデフォルト値
	config.Remote.Enabled = false
	config.Remote.Host = "localhost"
	config.Remote.Port = 22
	config.Remote.User = "user"
	config.Remote.KeyPath = ""
	config.Remote.KnownHosts = "~/.ssh/known_hosts"
	config.Remote.RemotePath = "/var/www/html/images"
	config.Remote.UseSSHAgent = true
	config.Remote.Timeout = 60

	// モード設定のデフォルト値
	config.Mode.DryRun = false

	// 入力設定のデフォルト値
	config.Input.Directory = "./images"
	config.Input.SupportedExtensions = []string{
		".jpg", ".jpeg", ".png", ".heic", ".heif",
	}

	// 変換設定のデフォルト値
	config.Conversion.Workers = 4
	config.Conversion.WebP.Enabled = true
	config.Conversion.WebP.Quality = 80
	config.Conversion.WebP.CompressionLevel = 4
	config.Conversion.AVIF.Enabled = true
	config.Conversion.AVIF.Quality = 40
	config.Conversion.AVIF.Speed = 6
	config.Conversion.AVIF.Lossless = false

	// FTPサーバー設定のデフォルト値
	config.FTP.Enabled = false
	config.FTP.Port = 2121
	config.FTP.User.Name = "ftpuser"
	config.FTP.User.Password = "ftppassword"
	config.FTP.Passive.Enabled = true
	config.FTP.Passive.PortRange = "50000-50100"

	// SSHサーバー設定のデフォルト値
	config.SSH.Enabled = false
	config.SSH.Port = 2222
	config.SSH.Auth.PasswordAuth = true
	config.SSH.Auth.PubkeyAuth = true
	config.SSH.Auth.AuthKeysFile = "~/.ssh/authorized_keys"

	// ログ設定のデフォルト値
	config.Logging.Level = "info"
	config.Logging.File = ""
	config.Logging.Directory = "logs" // デフォルトディレクトリを設定
	config.Logging.MaxSize = 10
	config.Logging.MaxBackups = 3
	config.Logging.MaxAge = 28
	config.Logging.Compress = true

	return config
}

// DefaultConversionConfig は変換設定のデフォルト値を返します
func DefaultConversionConfig() struct {
	Workers int
	WebP    struct {
		Enabled          bool
		Quality          int
		CompressionLevel int
	}
	AVIF struct {
		Enabled  bool
		Quality  int
		Speed    int
		Lossless bool
	}
} {
	var config struct {
		Workers int
		WebP    struct {
			Enabled          bool
			Quality          int
			CompressionLevel int
		}
		AVIF struct {
			Enabled  bool
			Quality  int
			Speed    int
			Lossless bool
		}
	}

	config.Workers = 4
	config.WebP.Enabled = true
	config.WebP.Quality = 80
	config.WebP.CompressionLevel = 4
	config.AVIF.Enabled = true
	config.AVIF.Quality = 40
	config.AVIF.Speed = 6
	config.AVIF.Lossless = false

	return config
}

// DefaultRemoteConfig はリモート設定のデフォルト値を返します
func DefaultRemoteConfig() RemoteConfig {
	return RemoteConfig{
		Enabled:     false,
		Host:        "localhost",
		Port:        22,
		User:        "user",
		KeyPath:     "",
		KnownHosts:  "~/.ssh/known_hosts",
		RemotePath:  "/var/www/html/images",
		UseSSHAgent: true,
		Timeout:     60,
	}
}

// DefaultFTPConfig はFTPサーバー設定のデフォルト値を返します
func DefaultFTPConfig() struct {
	Enabled bool
	Port    int
	User    struct {
		Name     string
		Password string
	}
	Passive struct {
		Enabled   bool
		PortRange string
	}
} {
	var config struct {
		Enabled bool
		Port    int
		User    struct {
			Name     string
			Password string
		}
		Passive struct {
			Enabled   bool
			PortRange string
		}
	}

	config.Enabled = false
	config.Port = 2121
	config.User.Name = "ftpuser"
	config.User.Password = "ftppassword"
	config.Passive.Enabled = true
	config.Passive.PortRange = "50000-50100"

	return config
}

// DefaultSSHConfig はSSHサーバー設定のデフォルト値を返します
func DefaultSSHConfig() struct {
	Enabled bool
	Port    int
	Auth    struct {
		PasswordAuth bool
		PubkeyAuth   bool
		AuthKeysFile string
	}
} {
	var config struct {
		Enabled bool
		Port    int
		Auth    struct {
			PasswordAuth bool
			PubkeyAuth   bool
			AuthKeysFile string
		}
	}

	config.Enabled = false
	config.Port = 2222
	config.Auth.PasswordAuth = true
	config.Auth.PubkeyAuth = true
	config.Auth.AuthKeysFile = "~/.ssh/authorized_keys"

	return config
}

// DefaultLoggingConfig はログ設定のデフォルト値を返します
func DefaultLoggingConfig() struct {
	Level      string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
} {
	var config struct {
		Level      string
		File       string
		MaxSize    int
		MaxBackups int
		MaxAge     int
		Compress   bool
	}

	config.Level = "info"
	config.File = ""
	config.MaxSize = 10
	config.MaxBackups = 3
	config.MaxAge = 28
	config.Compress = true

	return config
}
