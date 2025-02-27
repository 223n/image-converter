package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// ssh-agentから登録済みの鍵を取得してauthorized_keysに追加
func setupSSHKeys(authorizedKeysPath string) error {
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
