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
	// ssh-agentから鍵を取得
	agentKeys, err := s.getSSHAgentKeys()
	if err != nil {
		return err
	}

	// 既存の鍵ファイルを読み込む
	existingKeys, err := s.readExistingKeys(authorizedKeysPath)
	if err != nil {
		return err
	}

	// 新しい鍵をファイルに書き込む
	if err := s.writeKeysToFile(authorizedKeysPath, existingKeys, agentKeys); err != nil {
		return err
	}

	log.Printf("ssh-agentから%d個の鍵をauthorized_keysに追加しました", strings.Count(string(agentKeys), "\n")+1)
	return nil
}

// getSSHAgentKeys はssh-agentから公開鍵を取得します
func (s *SSHService) getSSHAgentKeys() ([]byte, error) {
	cmd := exec.Command("ssh-add", "-L")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ssh-add -L の実行に失敗しました: %v", err)
	}

	output := out.Bytes()

	// 鍵が登録されていない場合のチェック
	if strings.TrimSpace(string(output)) == "The agent has no identities." {
		return nil, fmt.Errorf("ssh-agentに登録された鍵がありません")
	}

	return output, nil
}

// readExistingKeys は既存の認証鍵ファイルを読み込みます
func (s *SSHService) readExistingKeys(authorizedKeysPath string) ([]byte, error) {
	// ファイルが存在しない場合は空のバイト配列を返す
	if _, err := os.Stat(authorizedKeysPath); os.IsNotExist(err) {
		return []byte{}, nil
	}

	// ファイルを読み込む
	existingKeys, err := os.ReadFile(authorizedKeysPath)
	if err != nil {
		return nil, fmt.Errorf("既存のauthorized_keysファイルの読み込みに失敗しました: %v", err)
	}

	return existingKeys, nil
}

// writeKeysToFile は鍵をファイルに書き込みます
func (s *SSHService) writeKeysToFile(authorizedKeysPath string, existingKeys, newKeys []byte) error {
	// 一時ファイルに書き込む（先にssh-agent鍵を保存）
	tempKeys, err := os.CreateTemp("", "ssh-keys-*")
	if err != nil {
		return fmt.Errorf("一時ファイルの作成に失敗しました: %v", err)
	}
	defer os.Remove(tempKeys.Name())

	if _, err := tempKeys.Write(newKeys); err != nil {
		return fmt.Errorf("一時ファイルへの書き込みに失敗しました: %v", err)
	}

	// 本番ファイルを作成
	authorizedKeysFile, err := os.Create(authorizedKeysPath)
	if err != nil {
		return fmt.Errorf("authorized_keysファイルの作成に失敗しました: %v", err)
	}
	defer authorizedKeysFile.Close()

	// 既存の鍵がある場合はコピー
	if err := s.appendExistingKeys(authorizedKeysFile, existingKeys); err != nil {
		return err
	}

	// 新しい鍵を追加
	if _, err := authorizedKeysFile.Write(newKeys); err != nil {
		return fmt.Errorf("ssh-agent鍵の書き込みに失敗しました: %v", err)
	}

	// 権限を設定
	return s.setFilePermissions(authorizedKeysPath)
}

// appendExistingKeys は既存の鍵をファイルに追加します
func (s *SSHService) appendExistingKeys(file *os.File, existingKeys []byte) error {
	if len(existingKeys) == 0 {
		return nil
	}

	if _, err := file.Write(existingKeys); err != nil {
		return fmt.Errorf("既存の鍵の書き込みに失敗しました: %v", err)
	}

	// 改行の確認と追加
	if !bytes.HasSuffix(existingKeys, []byte("\n")) {
		if _, err := file.Write([]byte("\n")); err != nil {
			return fmt.Errorf("改行の書き込みに失敗しました: %v", err)
		}
	}

	return nil
}

// setFilePermissions はファイルに適切な権限を設定します
func (s *SSHService) setFilePermissions(filePath string) error {
	// 権限を設定（600 = ユーザーのみ読み書き可能）
	if err := os.Chmod(filePath, 0600); err != nil {
		return fmt.Errorf("ファイルの権限設定に失敗しました: %v", err)
	}
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
