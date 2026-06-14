#!/bin/bash

# 简化的 Windows 构建脚本
# 快速构建带图标的 Windows 版本

set -e

# 获取版本信息
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"

APP_VERSION=$(bash "$SCRIPT_DIR/get_version.sh" version)
APP_NAME=$(bash "$SCRIPT_DIR/get_version.sh" name)
APP_AUTHOR=$(bash "$SCRIPT_DIR/get_version.sh" author)
APP_OPENSOURCE=$(bash "$SCRIPT_DIR/get_version.sh" opensource)

echo "🚀 开始构建 Windows 版本 v$APP_VERSION..."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到 Go 编译器"
    exit 1
fi

# 创建构建目录
mkdir -p dist/windows

# 设置交叉编译环境
export GOOS=windows
export GOARCH=amd64

echo "📦 构建 CLI 版本..."
export CGO_ENABLED=0
go build -ldflags="-s -w" -o dist/windows/oh-my-rime-cli.exe ./cmd/cli

echo "🎨 构建 GUI 版本..."
# 检查是否有 mingw-w64 支持
if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    echo "✅ 找到 mingw-w64，启用 CGO"
    export CGO_ENABLED=1
    export CC=x86_64-w64-mingw32-gcc
    
    echo "🔧 使用 Wails 构建带图标的 GUI 版本..."
    wails build -platform windows/amd64 -ldflags "-X 'oh-my-rime-cli/internal/constants.AppVersion=$APP_VERSION'" -clean
    
    # 临时复制结果到 dist 目录
    cp build/bin/oh-my-rime-gui.exe dist/windows/
    
    echo "✅ GUI 版本构建完成"
else
    echo "⚠️  未找到 mingw-w64，跳过 GUI 版本"
    echo "   安装方法："
    echo "   macOS: brew install mingw-w64"
    echo "   Ubuntu/Debian: sudo apt-get install gcc-mingw-w64"
fi

# 创建启动脚本
echo "📝 创建启动脚本..."

cat > dist/windows/start-cli.bat << 'EOF'
@echo off
chcp 65001 > nul
echo 启动 Oh My Rime CLI...
"%~dp0oh-my-rime-cli.exe"
pause
EOF

if [ -f "dist/windows/oh-my-rime-gui.exe" ]; then
    cat > dist/windows/start-gui.bat << 'EOF'
@echo off
echo 启动 Oh My Rime GUI...
start "" "%~dp0oh-my-rime-gui.exe"
EOF
fi

# 复制说明文件
cp README.md dist/windows/ 2>/dev/null || true
cp LICENSE dist/windows/ 2>/dev/null || true

# 创建使用说明
cat > dist/windows/使用说明.txt << EOF
$APP_NAME - Windows 版本 v$APP_VERSION

文件说明：
- oh-my-rime-cli.exe: 命令行版本
- oh-my-rime-gui.exe: 图形界面版本（如果存在）
- start-cli.bat: 启动命令行版本
- start-gui.bat: 启动图形界面版本（如果存在）

使用方法：
1. 双击 start-gui.bat 启动图形界面（推荐）
2. 双击 start-cli.bat 启动命令行版本
3. 或直接运行对应的 .exe 文件

注意事项：
- 首次运行可能被杀毒软件拦截，请添加信任
- 程序需要网络权限来下载更新
- 支持薄荷输入法和万象模型更新

项目地址：$APP_OPENSOURCE
作者：$APP_AUTHOR
EOF

echo ""
echo "🎉 Windows 版本构建完成！"
echo "📁 输出目录: dist/windows/"
echo ""
echo "📋 构建结果："
ls -la dist/windows/

echo ""
echo "💡 使用提示："
echo "1. 将 dist/windows/ 目录复制到 Windows 机器"
echo "2. 双击 start-gui.bat 启动图形界面"
echo "3. 或双击 start-cli.bat 启动命令行版本"