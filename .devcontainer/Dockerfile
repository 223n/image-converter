FROM mcr.microsoft.com/devcontainers/base:debian-12

# ユーザー設定
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=${USER_UID}

# 必要なパッケージのインストール
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    build-essential \
    pkg-config \
    libjpeg-dev \
    libpng-dev \
    libwebp-dev \
    libheif-dev \
    openssh-client \
    libaom-dev \
    pure-ftpd \
    openssh-server \
    nodejs \
    npm \
    git \
    webp \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# WebP関連のパッケージを確保
RUN apt-get update && apt-get install -y libwebp-dev webp

# Go言語のインストール
ARG GO_VERSION=1.20.4
RUN curl -sSL https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf -
ENV PATH=$PATH:/usr/local/go/bin:/home/${USERNAME}/go/bin

# 必要なGoパッケージは初期化したモジュール内でインストールするように変更
# go.modファイルが存在するときに依存関係を取得するようにpost-create.shで処理

# textlintと日本語関連ルールのインストール
RUN npm install -g textlint \
    textlint-rule-preset-ja-technical-writing \
    textlint-rule-preset-japanese \
    textlint-rule-spellcheck-tech-word

# SSHの設定
RUN mkdir -p /var/run/sshd
RUN echo 'PermitRootLogin no' >> /etc/ssh/sshd_config
RUN echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config

# Pure-FTPdの設定
RUN mkdir -p /etc/pure-ftpd/auth
RUN groupadd ftpgroup
RUN useradd -g ftpgroup -d /home/ftpuser -s /dev/null ftpuser
RUN mkdir -p /home/ftpuser
RUN chown -R ftpuser:ftpgroup /home/ftpuser

# vscodeユーザーをsudoersに追加
RUN echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} \
    && chmod 0440 /etc/sudoers.d/${USERNAME}

# vscodeユーザーをftpgroupに追加
RUN usermod -a -G ftpgroup ${USERNAME}

# ワークディレクトリの設定
WORKDIR /workspace

USER ${USERNAME}
