/*
Package sshutil は SSH 接続の設定のうち、画像変換の依存を持たない部分を提供します。

internal/remote は internal/converter 経由で cgo を要する libde265 に依存するため、
cgo が使えない環境ではビルドもテストもできません。ホスト鍵の検証は安全性に直結し、
回帰を検知できる状態に保ちたいので、この部分だけ独立させています。
*/
package sshutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyOptions はホスト鍵の検証方法を決める設定です。
type HostKeyOptions struct {
	// KnownHosts は known_hosts ファイルのパス。~ と環境変数を展開します。
	KnownHosts string
	// InsecureSkipHostKeyCheck が true のときだけ検証を省略します。
	InsecureSkipHostKeyCheck bool
}

// ResolveHostKeyCallback はホスト鍵の検証方法を決めます。
//
// 検証を省略するのは InsecureSkipHostKeyCheck が明示的に true の場合だけです。
// それ以外で known_hosts を読めなければ、接続させずエラーを返します。
//
// 以前の実装は ssh.InsecureIgnoreHostKey を初期値に置き、known_hosts の
// 読み込みに失敗しても警告を出すだけで接続を続けていました。ファイルが無い
// 環境では検証なしで接続し、中間者攻撃を検知できませんでした。
func ResolveHostKeyCallback(opts HostKeyOptions) (ssh.HostKeyCallback, error) {
	if opts.InsecureSkipHostKeyCheck {
		log.Printf("警告: ホスト鍵の検証を省略します（insecure_skip_host_key_check=true）。中間者攻撃を検知できません")
		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // 設定で明示的に選択された場合のみ
	}

	if opts.KnownHosts == "" {
		return nil, fmt.Errorf("known_hosts が未設定です。ホスト鍵を検証できないため接続しません。" +
			"known_hosts を設定するか、検証を省略する場合は insecure_skip_host_key_check を true にしてください")
	}

	expandedPath, err := ExpandPath(opts.KnownHosts)
	if err != nil {
		return nil, fmt.Errorf("known_hosts のパスを解決できません (%s): %w", opts.KnownHosts, err)
	}

	hostKeyCallback, err := knownhosts.New(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("known_hosts を読み込めません (%s): %w", expandedPath, err)
	}

	return hostKeyCallback, nil
}

// ExpandPath は環境変数と先頭の ~ を展開します。
//
// 以前は os.Getenv("HOME") を使っていました。Windows では HOME が設定されて
// いないことがあり、その場合 ~ が空文字へ潰れて不正なパスになります。
// os.UserHomeDir は Windows では USERPROFILE を参照します。
func ExpandPath(path string) (string, error) {
	expanded := os.ExpandEnv(path)

	if expanded == "~" || strings.HasPrefix(expanded, "~/") || strings.HasPrefix(expanded, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		rest := strings.TrimLeft(expanded[1:], `/\`)
		if rest == "" {
			return home, nil
		}
		return filepath.Join(home, rest), nil
	}

	return expanded, nil
}
