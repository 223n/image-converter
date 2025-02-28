package server

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/223n/image-converter/internal/config"
)

// FTPService はFTPサーバー機能を管理します
type FTPService struct {
	cmd      *exec.Cmd
	running  bool
	port     int
	user     string
	password string
	passive  bool
}

// NewFTPService は新しいFTPサービスを作成します
func NewFTPService() *FTPService {
	cfg := config.GetConfig()
	return &FTPService{
		port:     cfg.FTP.Port,
		user:     cfg.FTP.User.Name,
		password: cfg.FTP.User.Password,
		passive:  cfg.FTP.Passive.Enabled,
		running:  false,
	}
}

// Start はFTPサーバーを起動します
func (s *FTPService) Start() error {
	if s.running {
		return fmt.Errorf("FTPサーバーは既に実行中です")
	}

	// FTPサーバーの引数を準備
	args := s.prepareArgs()

	// コマンド実行
	s.cmd = exec.Command("pure-ftpd", args...)

	// コマンド実行
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("FTPサーバーの起動に失敗しました: %v\nシステムにpure-ftpdがインストールされていることを確認してください", err)
	}

	s.running = true
	log.Printf("FTPサーバーを起動しました（ポート: %d, PID: %d）", s.port, s.cmd.Process.Pid)

	// ユーザー設定を行う場合はここで実装
	// 例: s.configureUsers()

	return nil
}

// Stop はFTPサーバーを停止します
func (s *FTPService) Stop() error {
	if !s.running || s.cmd == nil || s.cmd.Process == nil {
		return nil // 既に停止している
	}

	// Gracefulにプロセスを終了
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// 失敗した場合はKill
		s.cmd.Process.Kill()
		return fmt.Errorf("FTPサーバーの停止に失敗しました: %v", err)
	}

	s.running = false
	log.Printf("FTPサーバーを停止しました")
	return nil
}

// prepareArgs はFTPサーバーのコマンドライン引数を準備します
func (s *FTPService) prepareArgs() []string {
	args := []string{
		"--bind", fmt.Sprintf("0.0.0.0,%d", s.port),
		"--chroot",     // ユーザーをホームディレクトリに制限
		"--createhome", // ホームディレクトリを自動作成
	}

	// 匿名ログインを無効化
	args = append(args, "--noanonymous")

	// パッシブモードが有効な場合
	if s.passive {
		cfg := config.GetConfig()
		args = append(args, "--passiveportrange", cfg.FTP.Passive.PortRange)
	}

	// 認証データベースの設定
	args = append(args, "--login", "puredb:/etc/pure-ftpd/pureftpd.pdb")

	return args
}

// configureUsers はFTPユーザーを設定します
// 実際の環境では、pure-pwコマンドなどを使用してユーザーを追加する必要があります
func (s *FTPService) configureUsers() error {
	// この実装例では、一時的なユーザー設定ファイルを作成
	// pure-pwユーティリティがインストールされていることを前提

	// 一時ファイルの作成
	tmpFile, err := os.CreateTemp("", "pure-pw-*")
	if err != nil {
		return fmt.Errorf("一時ファイルの作成に失敗しました: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// パスワード入力用の文字列を作成
	passwordInput := strings.NewReader(fmt.Sprintf("%s\n%s\n", s.password, s.password))

	// pure-pwコマンドでユーザー追加
	addCmd := exec.Command("pure-pw", "useradd", s.user, "-u", "ftpuser", "-g", "ftpgroup", "-d", "/home/ftpusers/"+s.user, "-m")

	// 標準入力にパスワードを設定
	addCmd.Stdin = passwordInput

	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ユーザー追加に失敗しました: %v, 出力: %s", err, string(output))
	}

	// データベースを更新
	updateCmd := exec.Command("pure-pw", "mkdb")
	if output, err := updateCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("データベース更新に失敗しました: %v, 出力: %s", err, string(output))
	}

	log.Printf("FTPユーザー '%s' を設定しました", s.user)
	return nil
}

// GetStatus はFTPサーバーの状態情報を返します
func (s *FTPService) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})
	status["running"] = s.running
	status["port"] = s.port

	if s.running && s.cmd != nil && s.cmd.Process != nil {
		status["pid"] = s.cmd.Process.Pid
	}

	return status
}

// IsRunning はFTPサーバーが実行中かどうかを返します
func (s *FTPService) IsRunning() bool {
	return s.running
}
