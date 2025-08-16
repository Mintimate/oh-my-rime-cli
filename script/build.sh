#!/bin/bash
#!/bin/bash

set -e

echo "🚀 Oh My Rime 完整构建脚本"
echo "================================"

echo "📦 安装基础 GO 依赖..."
go mod tidy

echo ""
echo "🎯 开始全平台构建..."

# 清理旧文件
echo "🧹 清理旧构建文件..."
rm -rf dist/
mkdir -p dist

echo ""
echo "=== 🖥️  构建 Windows 版本 ==="
bash script/build_windows.sh

# 查看是否为 macOS 系统
if [ -f /usr/bin/sw_vers ]; then
    echo ""
    echo "=== 🍎 构建 macOS 版本 ==="
    bash script/build_macOS.sh
fi

# 获取版本信息用于文件名
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_VERSION=$(bash "$SCRIPT_DIR/get_version.sh" version)

# 移动 macOS 构建结果到 dist 目录
if [ -f "Oh My Rime.app" ]; then
    mv "Oh My Rime.app" dist/
fi
if [ -f "Oh-My-Rime-v$APP_VERSION.dmg" ]; then
    mv "Oh-My-Rime-v$APP_VERSION.dmg" dist/
fi

echo ""
echo "=== 🐧 构建 Linux 版本 ==="
echo "📦 构建 Linux CLI..."
mkdir -p dist/linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/linux/oh-my-rime-cli-linux-amd64 ./cmd/cli
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/linux/oh-my-rime-cli-linux-arm64 ./cmd/cli

echo "🎨 构建 Linux GUI..."
# Linux GUI 需要在 Linux 系统上构建，这里只构建 CLI 版本
echo "⚠️  Linux GUI 版本需要在 Linux 系统上构建"

# 创建 Linux 使用说明
cat > dist/linux/README.txt << 'EOF'
Oh My Rime - Linux 版本

文件说明：
- oh-my-rime-cli-linux-amd64: AMD64 架构命令行版本
- oh-my-rime-cli-linux-arm64: ARM64 架构命令行版本

使用方法：
1. 给文件添加执行权限: chmod +x oh-my-rime-cli-linux-*
2. 运行对应架构的版本: ./oh-my-rime-cli-linux-amd64

注意事项：
- GUI 版本需要在 Linux 系统上编译
- 确保系统已安装 Rime 输入法

项目地址：https://github.com/Mintimate/oh-my-rime-cli
EOF

echo ""
echo "✅ 全平台构建完成！"
echo "📁 构建结果："
echo ""
find dist -type f -exec ls -lh {} \; | awk '{print "  " $9 " (" $5 ")"}'

echo ""
echo "📋 发布文件："
echo "  Windows: dist/windows/ (包含 CLI + GUI + 图标)"
echo "  macOS:   dist/Oh-My-Rime-v1.1.1.dmg (完整安装包)"
echo "  Linux:   dist/linux/ (CLI 版本)"
echo ""
echo "🎉 构建完成！可以开始发布了！"
