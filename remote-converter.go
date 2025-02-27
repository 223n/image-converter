package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHClient はリモートサーバーとの接続を管理
type SSHClient struct {
	config     *RemoteConfig
	client     *ssh.Client
	sftpClient *SFTPClient
}

// SFTPClient はSFTPプロトコルによるファイル転送を管理
type SFTPClient struct {
	client *ssh.Client
	sftp   *sftp.Client
}

// NewSSHClient は新しいSSHクライアントを作成
func NewSSHClient(config *RemoteConfig) (*SSHClient, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("リモート変換が無効です")
	}

	// SSHクライアント設定
	clientConfig := &ssh.ClientConfig{
		User:            config.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 開発用 - 本番環境では使用しないでください
		Timeout:         time.Duration(config.Timeout) * time.Second,
	}

	// 既知のホストファイルが指定されている場合は使用
	if config.KnownHosts != "" {
		expandedPath := os.ExpandEnv(config.KnownHosts)
		expandedPath = strings.Replace(expandedPath, "~", os.Getenv("HOME"), 1)

		hostKeyCallback, err := knownhosts.New(expandedPath)
		if err != nil {
			log.Printf("警告: 既知のホストファイルの読み込みに失敗しました: %v", err)
		} else {
			clientConfig.HostKeyCallback = hostKeyCallback
		}
	}

	// 認証方法の設定
	if config.UseSSHAgent {
		// SSH Agentを使用した認証
		socket := os.Getenv("SSH_AUTH_SOCK")
		if socket == "" {
			return nil, fmt.Errorf("SSH_AUTH_SOCK環境変数が設定されていません")
		}

		agentConn, err := net.Dial("unix", socket)
		if err != nil {
			return nil, fmt.Errorf("SSH Agentへの接続に失敗しました: %v", err)
		}

		agentClient := agent.NewClient(agentConn)
		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}
	} else if config.KeyPath != "" {
		// 秘密鍵ファイルを使用した認証
		expandedPath := os.ExpandEnv(config.KeyPath)
		expandedPath = strings.Replace(expandedPath, "~", os.Getenv("HOME"), 1)

		keyData, err := os.ReadFile(expandedPath)
		if err != nil {
			return nil, fmt.Errorf("秘密鍵ファイルの読み込みに失敗しました: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("秘密鍵の解析に失敗しました: %v", err)
		}

		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		return nil, fmt.Errorf("認証方法が指定されていません")
	}

	// SSHクライアント接続
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("SSHサーバーへの接続に失敗しました: %v", err)
	}

	// SFTPクライアントの作成
	sftpClient, err := NewSFTPClient(client)
	if err != nil {
		client.Close()
		return nil, err
	}

	return &SSHClient{
		config:     config,
		client:     client,
		sftpClient: sftpClient,
	}, nil
}

// Close はSSH接続を閉じる
func (c *SSHClient) Close() {
	if c.sftpClient != nil && c.sftpClient.sftp != nil {
		c.sftpClient.sftp.Close()
	}

	if c.client != nil {
		c.client.Close()
	}
}

// ExecuteCommand はリモートサーバーでコマンドを実行
func (c *SSHClient) ExecuteCommand(command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("セッションの作成に失敗しました: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("コマンド実行エラー: %v, 出力: %s", err, string(output))
	}

	return string(output), nil
}

// FindRemoteImages はリモートサーバー上の画像ファイルを検索
func (c *SSHClient) FindRemoteImages(extensions []string) ([]string, error) {
	// 拡張子をパイプ区切りの文字列に変換
	var extsFormatted []string
	for _, ext := range extensions {
		ext = strings.TrimPrefix(ext, ".")
		extsFormatted = append(extsFormatted, fmt.Sprintf("-name \"*.%s\"", ext))
	}
	extsStr := strings.Join(extsFormatted, " -o ")

	// findコマンドを作成
	cmd := fmt.Sprintf("find %s -type f \\( %s \\) | sort",
		c.config.RemotePath,
		extsStr)

	output, err := c.ExecuteCommand(cmd)
	if err != nil {
		return nil, err
	}

	// 出力を行に分割
	files := strings.Split(strings.TrimSpace(output), "\n")

	// 空の行を除外
	var result []string
	for _, file := range files {
		if file != "" {
			result = append(result, file)
		}
	}

	return result, nil
}

// DownloadFile はリモートサーバーからファイルをダウンロード
func (c *SSHClient) DownloadFile(remotePath, localPath string) error {
	// ローカルディレクトリを作成
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("ローカルディレクトリの作成に失敗しました: %v", err)
	}

	// リモートファイルを開く
	srcFile, err := c.sftpClient.sftp.Open(remotePath)
	if err != nil {
		return fmt.Errorf("リモートファイルを開くことができません: %v", err)
	}
	defer srcFile.Close()

	// ローカルファイルを作成
	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("ローカルファイルを作成できません: %v", err)
	}
	defer dstFile.Close()

	// ファイルをコピー
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("ファイルのコピーに失敗しました: %v", err)
	}

	log.Printf("リモートファイルのダウンロード: %s -> %s", remotePath, localPath)
	return nil
}

// UploadFile はリモートサーバーにファイルをアップロード
func (c *SSHClient) UploadFile(localPath, remotePath string) error {
	// リモートディレクトリを作成
	remoteDir := filepath.Dir(remotePath)
	if err := c.sftpClient.sftp.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("リモートディレクトリの作成に失敗しました: %v", err)
	}

	// ローカルファイルを開く
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("ローカルファイルを開くことができません: %v", err)
	}
	defer srcFile.Close()

	// リモートファイルを作成
	dstFile, err := c.sftpClient.sftp.Create(remotePath)
	if err != nil {
		return fmt.Errorf("リモートファイルを作成できません: %v", err)
	}
	defer dstFile.Close()

	// ファイルをコピー
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("ファイルのコピーに失敗しました: %v", err)
	}

	log.Printf("ローカルファイルのアップロード: %s -> %s", localPath, remotePath)
	return nil
}

// NewSFTPClient は新しいSFTPクライアントを作成
func NewSFTPClient(client *ssh.Client) (*SFTPClient, error) {
	// SFTPクライアントを作成
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("SFTPクライアントの作成に失敗しました: %v", err)
	}

	return &SFTPClient{
		client: client,
		sftp:   sftpClient,
	}, nil
}

// ConvertRemoteImages はリモートサーバー上の画像を変換
func ConvertRemoteImages(config *RemoteConfig, supportedExtensions []string) error {
	log.Println("リモートサーバー上の画像変換を開始します...")

	// SSHクライアント作成
	client, err := NewSSHClient(config)
	if err != nil {
		return fmt.Errorf("SSHクライアントの作成に失敗しました: %v", err)
	}
	defer client.Close()

	// リモートサーバー上の画像ファイルを検索
	imageFiles, err := client.FindRemoteImages(supportedExtensions)
	if err != nil {
		return fmt.Errorf("リモート画像の検索に失敗しました: %v", err)
	}

	totalFiles := len(imageFiles)
	log.Printf("リモートサーバーで変換対象の画像: %d個", totalFiles)

	// 進捗トラッカーを作成
	tracker := NewMultiProgressTracker(totalFiles, "リモート変換")

	// 一時ディレクトリの作成
	tempDir, err := os.MkdirTemp("", "remote-images-")
	if err != nil {
		return fmt.Errorf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, remoteFile := range imageFiles {
		// ベース名とディレクトリを取得
		baseFileName := filepath.Base(remoteFile)
		relPath, err := filepath.Rel(config.RemotePath, filepath.Dir(remoteFile))
		if err != nil {
			log.Printf("警告: 相対パスの計算に失敗しました: %v", err)
			relPath = ""
		}

		// ローカルのパスを作成
		localPath := filepath.Join(tempDir, relPath, baseFileName)

		// ファイルをダウンロード
		if err := client.DownloadFile(remoteFile, localPath); err != nil {
			log.Printf("ファイルのダウンロードに失敗しました %s: %v", remoteFile, err)
			tracker.IncrementFailed()
			continue
		}

		// 画像を変換（既存の変換機能を利用）
		if err := convertImage(localPath); err != nil {
			log.Printf("画像の変換に失敗しました %s: %v", localPath, err)
			tracker.IncrementFailed()
			continue
		}

		// 変換されたファイルをアップロード
		ext := filepath.Ext(localPath)
		baseName := strings.TrimSuffix(baseFileName, ext)

		// WebPファイルをアップロード
		webpLocalPath := filepath.Join(filepath.Dir(localPath), baseName+".webp")
		webpRemotePath := filepath.Join(filepath.Dir(remoteFile), baseName+".webp")

		if _, err := os.Stat(webpLocalPath); err == nil {
			if err := client.UploadFile(webpLocalPath, webpRemotePath); err != nil {
				log.Printf("WebPファイルのアップロードに失敗しました: %v", err)
			}
		}

		// AVIFファイルをアップロード
		avifLocalPath := filepath.Join(filepath.Dir(localPath), baseName+".avif")
		avifRemotePath := filepath.Join(filepath.Dir(remoteFile), baseName+".avif")

		if _, err := os.Stat(avifLocalPath); err == nil {
			if err := client.UploadFile(avifLocalPath, avifRemotePath); err != nil {
				log.Printf("AVIFファイルのアップロードに失敗しました: %v", err)
				// 変換自体は成功したのでエラーとはしない
			}
		}

		// この時点で1ファイルの処理が完了
		tracker.IncrementSuccess()
	}

	// 進捗トラッカーを完了
	tracker.Complete()

	log.Println("リモートサーバー上の画像変換が完了しました")
	return nil
}
