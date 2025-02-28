/*
Package server はFTPとSSHサーバー機能を提供します。
設定に基づいて、適切なサーバーを起動し管理します。
*/
package server

import (
	"fmt"
	"log"

	"github.com/yourusername/image-converter/internal/config"
)

// Service はサーバー管理サービスを表します
type Service struct {
	ftpService *FTPService
	sshService *SSHService
}

// NewService は新しいサーバーサービスを作成します
func NewService() *Service {
	return &Service{
		ftpService: NewFTPService(),
		sshService: NewSSHService(),
	}
}

// Start は設定に基づいてサーバーを起動します
func (s *Service) Start() error {
	// サーバーが有効かどうかをチェック
	if !config.IsFTPEnabled() && !config.IsSSHEnabled() {
		log.Println("サーバー機能が有効ではありません")
		return nil
	}

	// FTPサーバーの起動
	if config.IsFTPEnabled() {
		if err := s.ftpService.Start(); err != nil {
			log.Printf("FTPサーバーの起動に失敗しました: %v", err)
		}
	}

	// SSHサーバーの起動
	if config.IsSSHEnabled() {
		if err := s.sshService.Start(); err != nil {
			log.Printf("SSHサーバーの起動に失敗しました: %v", err)
		}
	}

	// いずれかのサーバーが起動している場合
	if config.IsFTPEnabled() || config.IsSSHEnabled() {
		fmt.Println("サーバーが稼働中です。Ctrl+Cで終了してください。")
		// 無限待機
		select {}
	}

	return nil
}

// Stop はすべてのサーバーを停止します
func (s *Service) Stop() error {
	var ftpErr, sshErr error

	// FTPサーバーの停止
	if config.IsFTPEnabled() {
		ftpErr = s.ftpService.Stop()
		if ftpErr != nil {
			log.Printf("FTPサーバーの停止に失敗しました: %v", ftpErr)
		}
	}

	// SSHサーバーの停止
	if config.IsSSHEnabled() {
		sshErr = s.sshService.Stop()
		if sshErr != nil {
			log.Printf("SSHサーバーの停止に失敗しました: %v", sshErr)
		}
	}

	// どちらかのサーバーでエラーがあった場合
	if ftpErr != nil || sshErr != nil {
		return fmt.Errorf("サーバーの停止に失敗しました")
	}

	return nil
}

// GetStatus はサーバーのステータス情報を返します
func (s *Service) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})

	if config.IsFTPEnabled() {
		status["ftp"] = s.ftpService.GetStatus()
	}

	if config.IsSSHEnabled() {
		status["ssh"] = s.sshService.GetStatus()
	}

	return status
}
