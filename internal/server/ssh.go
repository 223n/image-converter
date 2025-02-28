package server

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/223n/image-converter/internal/config"
)

// SSHService はSSHサーバー機能を管理します
type SSHService struct {
	cmd            *exec.Cmd
	running        bool
	port           int
	passwordAuth   bool
	authorizedKeys string
}

// NewSSHService は新しいSSHサービスを作成します
func NewSSHService() *SSHService {
	cfg := config.GetConfig()
	return &SSHService{
		port:           cfg.SSH.Port,
		passwordAuth:   cfg.SSH.Auth.PasswordAuth,
		authorizedKeys: cfg.SSH.Auth.AuthKeysFile,
		running:        false,
	}
}

// Start はSSHサーバーを起動します
func (s *SSHService) Start() error {
	if s.running {
		return fmt.Errorf("SSHサーバーは既に実行中です")
	}

	// SSHディレクトリとauthorized_keysの準備
	_, authorizedKeysPath, err := s.prepareSSHDirectory()
	if err != nil {
		return err
	}

	// ssh-agent から登録済みの鍵を取得して authorized_keys に追加
	if err := s.setupSSHKeys(authorizedKeysPath); err != nil {
		log.Printf("警告: SSH鍵の設定に失敗しました: %v", err)
		// 続行する（エラーがあっても他の認証方法が使える可能性あり）
	}

	// SSHコマンドの引数を構築
	args := s.prepareArgs(authorizedKeysPath)

	// コマンド実行
	s.cmd = exec.Command("sshd", args...)

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("SSHサーバーの起動に失敗しました: %v\nシステムにOpenSSHがインストールされていることを確認してください", err)
	}

	s.running = true
	log.Printf("SSHサーバーを起動しました（ポート: %d, PID: %d）", s.port, s.cmd.Process.Pid)
	return nil
}

// Stop はSSHサーバーを停止します
func (s *SSHService) Stop() error {
	if !s.running || s.cmd == nil || s.cmd.Process == nil {
		return nil // 既に停止している
	}

	// Gracefulにプロセスを終了
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// 失敗した場合はKill
		s.cmd.Process.Kill()
		return fmt.Errorf("SSHサーバーの停止に失敗しました: %v", err)
	}

	s.running = false
	log.Printf("SSHサーバーを停止しました")
	return nil
}

// prepareSSHDirectory はSSHディレクトリを準備します
func (s *SSHService) prepareSSHDirectory() (string, string, error) {
	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("ホームディレクトリの取得に失敗しました: %v", err)
	}

	// SSH設定ディレクトリの作成
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return "", "", fmt.Errorf("SSH設定ディレクトリの作成に失敗しました: %v", err)
	}

	// パスの展開
	authorizedKeysPath := s.authorizedKeys
	if authorizedKeysPath == "" {
		authorizedKeysPath = filepath.Join(sshDir, "authorized_keys")
	} else {
		// チルダやパス変数を展開
		authorizedKeysPath = os.ExpandEnv(authorizedKeysPath)
		authorizedKeysPath = strings.Replace(authorizedKeysPath, "~", homeDir, 1)
	}

	return sshDir, authorizedKeysPath, nil
}

// prepareArgs はSSHサーバーのコマンドライン引数を準備します
func (s *SSHService) prepareArgs(authorizedKeysPath string) []string {
	args := []string{
		"-D", // フォアグラウンドで実行
		"-p", fmt.Sprintf("%d", s.port),
	}

	// パスワード認証の設定
	if s.passwordAuth {
		args = append(args, "-o", "PasswordAuthentication=yes")
	} else {
		args = append(args, "-o", "PasswordAuthentication=no")
	}

	// 公開鍵認証の設定（常に有効）
	args = append(args, "-o", "PubkeyAuthentication=yes")
	args = append(args, "-o", fmt.Sprintf("AuthorizedKeysFile=%s", authorizedKeysPath))

	return args
}

// setupSSHKeys はSSH鍵を設定します
func (s *SSHService) setupSSHKeys(authorizedKeysPath string) error {
	// ssh-add -L コマンドを実行して登録済みの鍵を取得
	cmd := exec.Command("ssh-add", "-L")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh-add -L の実行に失敗しました: %v", err)
	}

	// 出力が "The agent has no identities." の場合は鍵が登録されていない
	if strings.TrimSpace(out.String()) == "The agent has no identities." {
		return fmt.Errorf("ssh-agentに登録された鍵がありません")
	}

	// 既存のauthorized_keysファイルを読み込む（存在する場合）
	var existingKeys []byte
	if _, err := os.Stat(authorizedKeysPath); err == nil {
		existingKeys, err = os.ReadFile(authorizedKeysPath)
		if err != nil {
			return fmt.Errorf("既存のauthorized_keysファイルの読み込みに失敗しました: %v", err)
		}
	}

	// ssh-agent鍵を一時ファイルに書き込む
	tempKeys, err := os.CreateTemp("", "ssh-keys-*")
	if err != nil {
		return fmt.Errorf("一時ファイルの作成に失敗しました: %v", err)
	}
	defer os.Remove(tempKeys.Name())

	if _, err := tempKeys.Write(out.Bytes()); err != nil {
		return fmt.Errorf("一時ファイルへの書き込みに失敗しました: %v", err)
	}

	// 既存のキーと新しいキーをマージ
	// まず既存のキーをコピー
	authorizedKeysFile, err := os.Create(authorizedKeysPath)
	if err != nil {
		return fmt.Errorf("authorized_keysファイルの作成に失敗しました: %v", err)
	}
	defer authorizedKeysFile.Close()

	if len(existingKeys) > 0 {
		if _, err := authorizedKeysFile.Write(existingKeys); err != nil {
			return fmt.Errorf("既存の鍵の書き込みに失敗しました: %v", err)
		}

		// 既存のファイルが改行で終わっていない場合は追加
		if !bytes.HasSuffix(existingKeys, []byte("\n")) {
			if _, err := authorizedKeysFile.Write([]byte("\n")); err != nil {
				return fmt.Errorf("改行の書き込みに失敗しました: %v", err)
			}
		}
	}

	// 新しい鍵を追加
	if _, err := authorizedKeysFile.Write(out.Bytes()); err != nil {
		return fmt.Errorf("ssh-agent鍵の書き込みに失敗しました: %v", err)
	}

	// 権限を設定（600 = ユーザーのみ読み書き可能）
	if err := os.Chmod(authorizedKeysPath, 0600); err != nil {
		return fmt.Errorf("authorized_keysファイルの権限設定に失敗しました: %v", err)
	}

	log.Printf("ssh-agentから%d個の鍵をauthorized_keysに追加しました", strings.Count(out.String(), "\n")+1)
	return nil
}

// GetStatus はSSHサーバーの状態情報を返します
func (s *SSHService) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})
	status["running"] = s.running
	status["port"] = s.port
	status["password_auth"] = s.passwordAuth

	if s.running && s.cmd != nil && s.cmd.Process != nil {
		status["pid"] = s.cmd.Process.Pid
	}

	return status
}

// IsRunning はSSHサーバーが実行中かどうかを返します
func (s *SSHService) IsRunning() bool {
	return s.running
}
