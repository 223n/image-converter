package sshutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// 有効な known_hosts の中身。ssh-keygen が出す形式の1行。
const sampleKnownHosts = "example.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB2sCJ0Zc0Nn+3jvVLQMBd0zkjRLtM6BdMPz9GJfKrHt\n"

func TestResolveHostKeyCallback(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "known_hosts")
	if err := os.WriteFile(valid, []byte(sampleKnownHosts), 0o600); err != nil {
		t.Fatalf("テスト用 known_hosts を作成できません: %v", err)
	}
	missing := filepath.Join(dir, "does-not-exist")

	tests := []struct {
		name    string
		opts    HostKeyOptions
		wantErr bool
		errHas  string
	}{
		{
			name:    "known_hosts が読めれば検証コールバックを返す",
			opts:    HostKeyOptions{KnownHosts: valid},
			wantErr: false,
		},
		{
			// 修正前はここで警告を出すだけで InsecureIgnoreHostKey のまま接続していた
			name:    "known_hosts が存在しなければ接続させない",
			opts:    HostKeyOptions{KnownHosts: missing},
			wantErr: true,
			errHas:  "known_hosts",
		},
		{
			name:    "known_hosts が未設定なら接続させない",
			opts:    HostKeyOptions{KnownHosts: ""},
			wantErr: true,
			errHas:  "insecure_skip_host_key_check",
		},
		{
			name:    "明示的に指定したときだけ検証を省略する",
			opts:    HostKeyOptions{KnownHosts: "", InsecureSkipHostKeyCheck: true},
			wantErr: false,
		},
		{
			name:    "検証省略は known_hosts の不備より優先する",
			opts:    HostKeyOptions{KnownHosts: missing, InsecureSkipHostKeyCheck: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb, err := ResolveHostKeyCallback(tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("エラーを期待したが nil。ホスト鍵の検証を素通りさせている")
				}
				if tt.errHas != "" && !strings.Contains(err.Error(), tt.errHas) {
					t.Errorf("エラーに %q を含むことを期待したが: %v", tt.errHas, err)
				}
				if cb != nil {
					t.Errorf("エラー時にコールバックを返してはいけない")
				}
				return
			}
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if cb == nil {
				t.Fatal("コールバックが nil")
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("ホームディレクトリを解決できないため省略: %v", err)
	}

	t.Run("先頭の ~ をホームディレクトリへ展開する", func(t *testing.T) {
		got, err := ExpandPath("~/.ssh/known_hosts")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		want := filepath.Join(home, ".ssh", "known_hosts")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("~ 単体はホームディレクトリになる", func(t *testing.T) {
		got, err := ExpandPath("~")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if got != home {
			t.Errorf("got %q, want %q", got, home)
		}
	})

	t.Run("HOME が未設定でも展開できる", func(t *testing.T) {
		// 修正前は os.Getenv("HOME") を使っており、Windows で HOME が無いと
		// ~ が空文字へ潰れて \.ssh\known_hosts のような不正なパスになった。
		if runtime.GOOS != "windows" {
			t.Skip("HOME を空にすると UserHomeDir も失敗するため Windows のみで確認する")
		}
		t.Setenv("HOME", "")
		got, err := ExpandPath("~/.ssh/known_hosts")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if strings.HasPrefix(got, string(filepath.Separator)) || got == filepath.Join(".ssh", "known_hosts") {
			t.Errorf("ホームが解決できていない: %q", got)
		}
	})

	t.Run("~ で始まらないパスはそのまま返す", func(t *testing.T) {
		in := filepath.Join("etc", "ssh", "ssh_known_hosts")
		got, err := ExpandPath(in)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if got != in {
			t.Errorf("got %q, want %q", got, in)
		}
	})

	t.Run("環境変数を展開する", func(t *testing.T) {
		t.Setenv("WC_TEST_DIR", "custom")
		got, err := ExpandPath("$WC_TEST_DIR/known_hosts")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if got != "custom/known_hosts" {
			t.Errorf("got %q, want %q", got, "custom/known_hosts")
		}
	})
}
