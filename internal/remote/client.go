/*
Package remote はリモートサーバーとの接続と操作に関する機能を提供します。
*/
package remote

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

	"github.com/223n/image-converter/internal/config"
	"github.com/223n/image-converter/pkg/imageutils"
)

// Client はリモートサーバーとの接続を管理します
type Client struct {
	config     *config.RemoteConfig
	client     *ssh.Client
	sftpClient *SFTPClient
}

// SFTPClient はSFTPプロトコルによるファイル転送を管理します
type SFTPClient struct {
	client *ssh.Client
	sftp   *sftp.Client
}

// NewClient は新しいリモートクライアントを作成します
func NewClient(cfg *config.RemoteConfig) (*Client, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("リモート変換が無効です")
	}

	// SSHクライアント設定
	clientConfig := &ssh.ClientConfig{
		User:            cfg.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 開発用 - 本番環境では使用しないでください
		Timeout:         time.Duration(cfg.Timeout) * time.Second,
	}

	// 既知のホストファイルが指定されている場合は使用
	if cfg.KnownHosts != "" {
		expandedPath := os.ExpandEnv(cfg.KnownHosts)
		expandedPath = strings.Replace(expandedPath, "~", os.Getenv("HOME"), 1)

		hostKeyCallback, err := knownhosts.New(expandedPath)
		if err != nil {
			log.Printf("警告: 既知のホストファイルの読み込みに失敗しました: %v", err)
		} else {
			clientConfig.HostKeyCallback = hostKeyCallback
		}
	}

	// 認証方法の設定
	if err := setupAuthentication(cfg, clientConfig); err != nil {
		return nil, err
	}

	// SSHクライアント接続
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("SSHサーバーへの接続に失敗しました: %v", err)
	}

	// SFTPクライアントの作成
	sftpClient, err := newSFTPClient(client)
	if err != nil {
		client.Close()
		return nil, err
	}

	return &Client{
		config:     cfg,
		client:     client,
		sftpClient: sftpClient,
	}, nil
}

// setupAuthentication は認証設定を行います
func setupAuthentication(cfg *config.RemoteConfig, clientConfig *ssh.ClientConfig) error {
	if cfg.UseSSHAgent {
		// SSH Agentを使用した認証
		socket := os.Getenv("SSH_AUTH_SOCK")
		if socket == "" {
			return fmt.Errorf("SSH_AUTH_SOCK環境変数が設定されていません")
		}

		agentConn, err := net.Dial("unix", socket)
		if err != nil {
			return fmt.Errorf("SSH Agentへの接続に失敗しました: %v", err)
		}

		agentClient := agent.NewClient(agentConn)
		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}
	} else if cfg.KeyPath != "" {
		// 秘密鍵ファイルを使用した認証
		expandedPath := os.ExpandEnv(cfg.KeyPath)
		expandedPath = strings.Replace(expandedPath, "~", os.Getenv("HOME"), 1)

		keyData, err := os.ReadFile(expandedPath)
		if err != nil {
			return fmt.Errorf("秘密鍵ファイルの読み込みに失敗しました: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return fmt.Errorf("秘密鍵の解析に失敗しました: %v", err)
		}

		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		return fmt.Errorf("認証方法が指定されていません")
	}

	return nil
}

// Close は接続を閉じます
func (c *Client) Close() {
	if c.sftpClient != nil && c.sftpClient.sftp != nil {
		c.sftpClient.sftp.Close()
	}

	if c.client != nil {
		c.client.Close()
	}
}

// ExecuteCommand はリモートサーバーでコマンドを実行します
func (c *Client) ExecuteCommand(command string) (string, error) {
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

// DownloadFile はリモートサーバーからファイルをダウンロードします
func (c *Client) DownloadFile(remotePath, localPath string) error {
	// リトライ設定
	retryConfig := newDefaultRetryConfig()

	return withRetry(func() error {
		// ローカルディレクトリを作成
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return fmt.Errorf("ローカルディレクトリの作成に失敗しました: %v", err)
		}

		// クライアントの接続状態を確認
		if c.client == nil || c.sftpClient == nil || c.sftpClient.sftp == nil {
			log.Printf("警告: SSH/SFTP接続が閉じられています。再接続を試みます...")
			if err := c.reconnect(); err != nil {
				return fmt.Errorf("再接続に失敗しました: %v", err)
			}
		}

		// リモートファイルを開く
		srcFile, err := c.sftpClient.sftp.Open(remotePath)
		if err != nil {
			// 接続エラーの場合は再接続を試みる
			if isConnectionError(err) {
				log.Printf("接続エラーが発生しました。再接続を試みます...")
				if reconnErr := c.reconnect(); reconnErr != nil {
					return fmt.Errorf("リモートファイルのオープンに失敗し、再接続もできませんでした: %v, 再接続エラー: %v", err, reconnErr)
				}

				// 再接続後に再度ファイルを開く
				srcFile, err = c.sftpClient.sftp.Open(remotePath)
				if err != nil {
					return fmt.Errorf("再接続後もリモートファイルを開くことができません: %v", err)
				}
			} else {
				return fmt.Errorf("リモートファイルを開くことができません: %v", err)
			}
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
			// 接続エラーの場合、ファイルを閉じて削除し、次のリトライでまた最初から
			os.Remove(localPath)
			return fmt.Errorf("ファイルのコピーに失敗しました: %v", err)
		}

		log.Printf("リモートファイルのダウンロード: %s -> %s", remotePath, localPath)
		return nil
	}, retryConfig)
}

// UploadFile はリモートサーバーにファイルをアップロードします
func (c *Client) UploadFile(localPath, remotePath string) error {
	// リトライ設定
	retryConfig := newDefaultRetryConfig()

	return withRetry(func() error {
		// ファイルサイズと整合性をチェック
		valid, fileSize := imageutils.IsValidFile(localPath)
		if !valid {
			return fmt.Errorf("無効なファイルです: %s", localPath)
		}

		// クライアントの接続状態を確認
		if c.client == nil || c.sftpClient == nil || c.sftpClient.sftp == nil {
			log.Printf("警告: SSH/SFTP接続が閉じられています。再接続を試みます...")
			if err := c.reconnect(); err != nil {
				return fmt.Errorf("再接続に失敗しました: %v", err)
			}
		}

		// リモートディレクトリを作成
		remoteDir := filepath.Dir(remotePath)
		if err := c.sftpClient.sftp.MkdirAll(remoteDir); err != nil {
			// 接続エラーの場合は再接続を試みる
			if isConnectionError(err) {
				log.Printf("接続エラーが発生しました。再接続を試みます...")
				if reconnErr := c.reconnect(); reconnErr != nil {
					return fmt.Errorf("リモートディレクトリの作成に失敗し、再接続もできませんでした: %v, 再接続エラー: %v", err, reconnErr)
				}

				// 再接続後に再度ディレクトリを作成
				err = c.sftpClient.sftp.MkdirAll(remoteDir)
				if err != nil {
					return fmt.Errorf("再接続後もリモートディレクトリを作成できません: %v", err)
				}
			} else {
				return fmt.Errorf("リモートディレクトリの作成に失敗しました: %v", err)
			}
		}

		// ローカルファイルを開く
		srcFile, err := os.Open(localPath)
		if err != nil {
			return fmt.Errorf("ローカルファイルを開くことができません: %v", err)
		}
		defer srcFile.Close()

		// リモートファイルを作成
		dstFile, err := c.createRemoteFile(remotePath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// ファイルをコピー
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return fmt.Errorf("ファイルのコピーに失敗しました: %v", err)
		}

		log.Printf("ローカルファイルのアップロード: %s -> %s (サイズ: %d バイト)", localPath, remotePath, fileSize)
		return nil
	}, retryConfig)
}

// createRemoteFile はリモートファイルを作成します
func (c *Client) createRemoteFile(remotePath string) (*sftp.File, error) {
	dstFile, err := c.sftpClient.sftp.Create(remotePath)

	// 接続エラーの場合は再接続を試みる
	if err != nil && isConnectionError(err) {
		log.Printf("接続エラーが発生しました。再接続を試みます...")
		if reconnErr := c.reconnect(); reconnErr != nil {
			return nil, fmt.Errorf("リモートファイルの作成に失敗し、再接続もできませんでした: %v, 再接続エラー: %v", err, reconnErr)
		}

		// 再接続後に再度ファイルを作成
		dstFile, err = c.sftpClient.sftp.Create(remotePath)
		if err != nil {
			return nil, fmt.Errorf("再接続後もリモートファイルを作成できません: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("リモートファイルを作成できません: %v", err)
	}

	return dstFile, nil
}

// reconnect はSSHおよびSFTP接続を再確立します
func (c *Client) reconnect() error {
	// 既存の接続をクローズ
	if c.sftpClient != nil && c.sftpClient.sftp != nil {
		c.sftpClient.sftp.Close()
	}
	if c.client != nil {
		c.client.Close()
	}

	// 新しいSSHクライアントの作成
	client, err := NewClient(c.config)
	if err != nil {
		return fmt.Errorf("SSH再接続に失敗しました: %v", err)
	}

	// 接続情報を更新
	c.client = client.client
	c.sftpClient = client.sftpClient

	log.Printf("SSH/SFTP接続を再確立しました")
	return nil
}

// newSFTPClient は新しいSFTPクライアントを作成します
func newSFTPClient(client *ssh.Client) (*SFTPClient, error) {
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
