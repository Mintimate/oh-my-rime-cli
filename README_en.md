# oh-my-rime-cli

A command-line tool for automatic update and management of Rime [Oh-my-rime](https://github.com/Mintimate/oh-my-rime) configuration, supporting Windows, Linux, and macOS.

## Features
- One-click download and update of Rime main scheme, model, and dictionaries
- Auto-detect operating system and select configuration directory intelligently
- Windows: automatically locate Rime user directory via registry
- Interactive and user-friendly menu (CLI), supports multiple input method environments

```mermaid
flowchart TD
    A[🌟 启动程序 <br/> macOS/Linux使用终端<br/>Windows双击即可] --> B[🖥️ 检测操作系统]
    B --支持的操作系统--> C[📋 显示主菜单]
    B --不支持/检测失败--> I
    
    C --> D{"✨用户选择"}
    D -->|① 更新方案| E[💼 下载薄荷方案包]
    D -->|② 更新模型| F[🧠 下载万象模型]
    D -->|③ 更新词库| G[📚 下载薄荷方案包<br/>提取词库]
    D -->|④ 自定义| H[🔗 粘贴配置链接]
    D -->|⏹️ 退出| I[👋 结束程序]
    
    E --> J[📂 选择输入法目录<br/>「Linux 和 macOS 需要」]
    F --> J
    G --> J
    H --> J
    
    J --> K{🛠️ 更新类型}
    K -->|方案更新| L[📤 解压&替换方案包]
    K -->|模型替换| M[💾 替换模型文件]
    K -->|词库更新| N[🔄 更新词库]
    K -->|自定义| O{🔍 文件类型}
    
    O -->|📦 ZIP文件| L
    O -->|🧾 模型文件| M
    
    L --> P[✅ 更新成功！]
    M --> P
    N --> P
    
    P --> C

    style A fill:#7E57C2,stroke:#5E35B1,color:white,stroke-width:2px,stroke-dasharray:5
    style I fill:#EF5350,stroke:#D32F2F,color:white,stroke-width:2px
    style C fill:#42A5F5,stroke:#1976D2,color:white
    style D fill:#FFCA28,stroke:#FFA000,color:black,stroke-width:2px
    style E fill:#26A69A,stroke:#00897B
    style F fill:#5C6BC0,stroke:#3949AB
    style G fill:#29B6F6,stroke:#039BE5
    style H fill:#FF7043,stroke:#F4511E
    style J fill:#AB47BC,stroke:#8E24AA,color:white
    style K fill:#FFEE58,stroke:#FDD835,color:black
    style L fill:#66BB6A,stroke:#43A047,color:white
    style M fill:#5C6BC0,stroke:#3949AB,color:white
    style N fill:#29B6F6,stroke:#039BE5,color:white
    style O fill:#FFA726,stroke:#FB8C00,color:black
    style P fill:#66BB6A,stroke:#43A047,color:white,stroke-width:2px
    linkStyle default stroke:#90A4AE,stroke-width:2px
```

## Dependencies
- Go 1.24 or above
- On Windows, requires `golang.org/x/sys/windows/registry` package

## Usage

### 1. Build

#### Windows
Build directly on Windows:
```sh
# Recommended to run in Windows command line
cd oh-my-rime-cli
go build -o oh-my-rime-cli.exe .
```

#### macOS/Linux
Build on macOS or Linux:
```sh
cd oh-my-rime-cli
go build -o oh-my-rime-cli .
```

#### Cross-compile for Windows
Build Windows executable on macOS/Linux:
```sh
GOOS=windows GOARCH=amd64 go build -o oh-my-rime-cli.exe .
```

### 2. Run

Double-click or run the compiled program in the command line, and follow the prompts to select actions and configuration directory.

## Logic Details

Windows Registry Support
- On Windows, automatically reads the `RimeUserDir` value from `HKEY_CURRENT_USER\Software\Rime\Weasel`.
- If the registry key does not exist or fails to read, falls back to `%APPDATA%\Rime` directory.

## Windows Antivirus Issues

Windows Defender and other antivirus software may falsely report Go-compiled executables as viruses. Please ensure you download from the official channel or compile by yourself, and add trust if necessary.

> No code signing certificate is provided, so the executable is unsigned. (o′ω`o)ノ

## License
- MIT License
- Issues and PRs are welcome

## Support

- [Mintimate's Blog: https://www.mintimate.cn](https://www.mintimate.cn)
- [Mintimate的爱发电: 加入电圈，支持创造!](https://afdian.net/a/mintimate)
- [Bilibili：@Mintimate](https://space.bilibili.com/355567627)
- [Youtube：@Mintimate](https://www.youtube.com/channel/UCI7LLdUGNzkcKOE7grAqCoA)
