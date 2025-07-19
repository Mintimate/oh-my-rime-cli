# 帮助文档地址: https://docs.cnb.cool/zh/vscode/quick-start.html
FROM golang:1.24.5

# 腾讯云软件源使用示例: https://cnb.cool/examples/mirrors/mirrors.cloud.tencent.com

# 安装 ssh 服务(用于支持 VSCode 客户端通过 Remote-SSH 访问开发环境);
# 脚本必须依赖: unzip、wget/curl、zsh、git
RUN apt update && \
    apt install -y \
    curl unzip zsh git \
    openssh-server \
    vim \
    webp && \
    chsh -s /bin/zsh

RUN yes | zsh -c "$(wget -qO- 'https://cnb.cool/Mintimate/tool-forge/oh-my-zsh/-/git/raw/main/AutoInstallZshWithPlugin.sh')"

# 安装 code-server 和 vscode 常用插件
RUN curl -fsSL https://code-server.dev/install.sh | sh \
  && code-server --install-extension redhat.vscode-yaml \
  && code-server --install-extension eamodio.gitlens \
  && code-server --install-extension tencent-cloud.coding-copilot \
  && code-server --install-extension golang.go \
  && echo done

# 指定字符集支持命令行输入中文（根据需要选择字符集）
ENV LANG C.UTF-8
ENV LANGUAGE C.UTF-8
