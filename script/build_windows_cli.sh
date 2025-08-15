#!/bin/bash

# Windows CLI 构建脚本 - 纯命令行版本，无 GUI 依赖
# 适用于 macOS/Linux 交叉编译

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
APP_NAME="Oh My Rime CLI"
VERSION="1.1.1"
BUILD_DIR="build/windows-cli"
DIST_DIR="dist/windows-cli"

echo -e "${BLUE}开始构建 Windows CLI 版本 (交叉编译)...${NC}"

# 清理之前的构建
echo -e "${YELLOW}清理构建目录...${NC}"
rm -rf "$BUILD_DIR"
rm -rf "$DIST_DIR"
mkdir -p "$BUILD_DIR"
mkdir -p "$DIST_DIR"

# 检查必要的工具
echo -e "${YELLOW}检查构建工具...${NC}"

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: Go 未安装${NC}"
    exit 1
fi

# 检查是否支持 Windows 交叉编译
echo -e "${YELLOW}检查 Windows 交叉编译支持...${NC}"
if go tool dist list | grep -q "windows/amd64"; then
    echo -e "${GREEN}✓ 支持 Windows amd64 交叉编译${NC}"
else
    echo -e "${RED}✗ 不支持 Windows 交叉编译${NC}"
    exit 1
fi

# 设置交叉编译环境变量（纯 CLI 版本不需要 CGO）
echo -e "${YELLOW}设置交叉编译环境...${NC}"
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=0

# 构建 CLI 版本
echo -e "${YELLOW}构建纯 CLI 版本...${NC}"
go build -ldflags="-s -w" -o "$BUILD_DIR/oh-my-rime-cli.exe" ./cmd/cli

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ CLI 版本构建成功${NC}"
else
    echo -e "${RED}✗ CLI 版本构建失败${NC}"
    exit 1
fi

# 构建带 GUI 的版本（如果可能的话）
echo -e "${YELLOW}尝试构建完整版本（包含 GUI）...${NC}"
export CGO_ENABLED=1

# 检查是否有交叉编译工具
if [[ "$OSTYPE" == "darwin"* ]]; then
    if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        echo -e "${GREEN}✓ 找到 mingw-w64 交叉编译器${NC}"
        export CC=x86_64-w64-mingw32-gcc
        export CXX=x86_64-w64-mingw32-g++
        
        go build -ldflags="-s -w -H windowsgui" -o "$BUILD_DIR/oh-my-rime-gui.exe" .
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ GUI 版本构建成功${NC}"
        else
            echo -e "${YELLOW}! GUI 版本构建失败，但 CLI 版本可用${NC}"
        fi
    else
        echo -e "${YELLOW}! 未找到 mingw-w64，跳过 GUI 版本构建${NC}"
        echo -e "${YELLOW}  安装方法: brew install mingw-w64${NC}"
    fi
fi

# 创建启动脚本
echo -e "${YELLOW}创建启动脚本...${NC}"

# CLI 启动脚本
cat > "$BUILD_DIR/oh-my-rime-cli.bat" << 'EOF'
@echo off
chcp 65001 > nul
echo 启动 Oh My Rime CLI 版本...
"%~dp0oh-my-rime-cli.exe"
pause
EOF

# 如果 GUI 版本存在，创建 GUI 启动脚本
if [ -f "$BUILD_DIR/oh-my-rime-gui.exe" ]; then
    cat > "$BUILD_DIR/oh-my-rime-gui.bat" << 'EOF'
@echo off
echo 启动 Oh My Rime GUI 版本...
start "" "%~dp0oh-my-rime-gui.exe"
EOF
fi

# 创建使用说明
cat > "$BUILD_DIR/README.txt" << 'EOF'
Oh My Rime - Windows CLI 版本

文件说明：
- oh-my-rime-cli.exe: 纯命令行版本（推荐）
- oh-my-rime-gui.exe: 图形界面版本（如果存在）
- oh-my-rime-cli.bat: CLI 启动脚本
- oh-my-rime-gui.bat: GUI 启动脚本（如果存在）

使用方法：
1. 双击 oh-my-rime-cli.bat 启动命令行版本
2. 或者在命令提示符中运行 oh-my-rime-cli.exe
3. 如果有 GUI 版本，可以双击 oh-my-rime-gui.bat

功能特性：
- 更新薄荷输入法方案
- 更新万象模型和词库
- 自定义方案和模型更新
- 跨平台支持

注意事项：
- 首次运行可能被 Windows Defender 拦截，请添加信任
- 建议将程序放置在非系统目录中
- 如需卸载，直接删除整个文件夹即可

项目地址：https://cnb.cool/Mintimate/rime/oh-my-rime-cli
作者：Mintimate
版本：1.1.1
EOF

# 复制必要文件
echo -e "${YELLOW}复制项目文件...${NC}"
if [ -f "README.md" ]; then
    cp README.md "$BUILD_DIR/"
fi
if [ -f "LICENSE" ]; then
    cp LICENSE "$BUILD_DIR/"
fi

# 创建发布包
echo -e "${YELLOW}创建发布包...${NC}"
cd "$BUILD_DIR"
zip -r "../$DIST_DIR/oh-my-rime-windows-cli-amd64.zip" .
cd - > /dev/null

# 显示构建结果
echo -e "${GREEN}构建完成！${NC}"
echo -e "${BLUE}构建文件位置:${NC}"
echo -e "  - 构建目录: $BUILD_DIR"
echo -e "  - 发布包: $DIST_DIR/oh-my-rime-windows-cli-amd64.zip"
echo ""
echo -e "${BLUE}文件列表:${NC}"
ls -la "$BUILD_DIR"
echo ""

# 显示文件大小
echo -e "${BLUE}文件大小:${NC}"
if [ -f "$BUILD_DIR/oh-my-rime-cli.exe" ]; then
    CLI_SIZE=$(ls -lh "$BUILD_DIR/oh-my-rime-cli.exe" | awk '{print $5}')
    echo -e "  - CLI 版本: $CLI_SIZE"
fi
if [ -f "$BUILD_DIR/oh-my-rime-gui.exe" ]; then
    GUI_SIZE=$(ls -lh "$BUILD_DIR/oh-my-rime-gui.exe" | awk '{print $5}')
    echo -e "  - GUI 版本: $GUI_SIZE"
fi

echo -e "${GREEN}Windows CLI 版本构建成功！${NC}"

# 提供使用建议
echo -e "${YELLOW}使用建议:${NC}"
echo -e "1. 将发布包发送到 Windows 机器上解压"
echo -e "2. 双击 oh-my-rime-cli.bat 启动命令行版本"
echo -e "3. 或在 PowerShell/CMD 中直接运行 oh-my-rime-cli.exe"
echo -e "4. 首次运行可能需要在 Windows Defender 中添加信任"
echo -e "5. CLI 版本功能完整，推荐使用"