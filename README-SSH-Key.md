# SSH鍵の設定

このプログラムでは、SSHサーバー機能を有効にすると、`ssh-agent`に登録されている鍵を自動的に認識し、利用できるようになります。以下では、SSH鍵の登録方法と確認方法について説明します。

## 1. SSHキーペアの生成（まだ持っていない場合）

SSHキーペアをまだ持っていない場合は、以下のコマンドで生成できます：

```bash
ssh-keygen -t ed25519 -C "your_email@example.com"
```

または、より古い形式のRSA鍵を生成する場合：

```bash
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
```

生成されたキーペアは通常、`~/.ssh/id_ed25519`（または`~/.ssh/id_rsa`）と`~/.ssh/id_ed25519.pub`（または`~/.ssh/id_rsa.pub`）に保存されます。

## 2. ssh-agentの起動

`ssh-agent`がまだ実行されていない場合は、以下のコマンドで起動します：

```bash
eval "$(ssh-agent -s)"
```

## 3. 鍵をssh-agentに登録

作成した秘密鍵を`ssh-agent`に登録します：

```bash
ssh-add ~/.ssh/id_ed25519
```

または、RSA鍵の場合：

```bash
ssh-add ~/.ssh/id_rsa
```

## 4. 登録された鍵の確認

`ssh-agent`に登録されている鍵を確認するには：

```bash
ssh-add -l
```

登録されている鍵のフィンガープリントが表示されます。

完全な公開鍵を表示するには：

```bash
ssh-add -L
```

## 5. プログラムでのSSH鍵の利用

プログラムのSSHサーバー機能を有効にすると（`config.yml`の`ssh.enabled`を`true`に設定）、自動的に`ssh-agent`から鍵を取得し、アクセスを許可します。

特別な設定は必要ありません。設定ファイルで`ssh.auth.password_auth`を`true`にすることで、パスワード認証と併用することもできます。

## トラブルシューティング

### 鍵が認識されない場合

1. `ssh-agent`が実行されていることを確認：
   何も表示されない場合は、`ssh-agent`を起動してください。
        ```bash
        echo $SSH_AUTH_SOCK
        ```

2. 鍵が登録されていることを確認：
  「The agent has no identities.」と表示される場合は、鍵を登録してください。
        ```bash
        ssh-add -l
        ```

3. 権限の確認：
   秘密鍵ファイルの権限が正しいことを確認します：
        ```bash
        chmod 600 ~/.ssh/id_ed25519
        ```

4. ログの確認：
   プログラムのログを確認して、SSH鍵の設定に関するエラーメッセージを確認してください。
