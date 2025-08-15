#!/bin/bash

# Windows 构建脚本 - 适用于 macOS/Linux 交叉编译
# 支持 CLI 和 GUI 模式，包含图标

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
APP_NAME="Oh My Rime"
VERSION="1.1.1"
BUILD_DIR="build/windows"
DIST_DIR="dist/windows"

echo -e "${BLUE}开始构建 Windows 版本 (交叉编译)...${NC}"

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

# 检查 Fyne 交叉编译工具
echo -e "${YELLOW}检查 Fyne 交叉编译工具...${NC}"
if command -v fyne &> /dev/null; then
    echo -e "${GREEN}✓ 找到 fyne 命令行工具${NC}"
    USE_FYNE_PACKAGE=true
else
    echo -e "${YELLOW}! 未找到 fyne 命令行工具，尝试直接编译${NC}"
    echo -e "${YELLOW}  如需更好的交叉编译支持，请安装: go install fyne.io/fyne/v2/cmd/fyne@latest${NC}"
    USE_FYNE_PACKAGE=false
fi

# 设置交叉编译环境变量
echo -e "${YELLOW}设置交叉编译环境...${NC}"
export GOOS=windows
export GOARCH=amd64

# 对于 Fyne GUI 应用，在 macOS 上交叉编译到 Windows 需要特殊处理
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo -e "${YELLOW}检测到 macOS 环境，配置 Windows 交叉编译...${NC}"
    
    # 检查是否安装了 mingw-w64
    if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        echo -e "${GREEN}✓ 找到 mingw-w64 交叉编译器${NC}"
        export CGO_ENABLED=1
        export CC=x86_64-w64-mingw32-gcc
        export CXX=x86_64-w64-mingw32-g++
    else
        echo -e "${YELLOW}! 未找到 mingw-w64，尝试禁用 CGO${NC}"
        echo -e "${YELLOW}  安装方法: brew install mingw-w64${NC}"
        export CGO_ENABLED=0
    fi
else
    # Linux 环境
    echo -e "${YELLOW}检测到 Linux 环境${NC}"
    if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        export CGO_ENABLED=1
        export CC=x86_64-w64-mingw32-gcc
        export CXX=x86_64-w64-mingw32-g++
    else
        echo -e "${YELLOW}! 未找到 mingw-w64，尝试禁用 CGO${NC}"
        echo -e "${YELLOW}  安装方法: sudo apt-get install gcc-mingw-w64${NC}"
        export CGO_ENABLED=0
    fi
fi

# 构建 CLI 版本（控制台应用）
echo -e "${YELLOW}构建 CLI 版本...${NC}"
if $USE_FYNE_PACKAGE; then
    # 使用 fyne 工具打包
    fyne package -os windows -name "oh-my-rime-cli" -exe "$BUILD_DIR/oh-my-rime-cli.exe"
else
    # 直接使用 go build
    go build -ldflags="-s -w" -o "$BUILD_DIR/oh-my-rime-cli.exe" .
fi

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ CLI 版本构建成功${NC}"
else
    echo -e "${RED}✗ CLI 版本构建失败${NC}"
    exit 1
fi

# 构建 GUI 版本（无控制台窗口）
echo -e "${YELLOW}构建 GUI 版本...${NC}"
if $USE_FYNE_PACKAGE; then
    # 使用 fyne 工具打包 GUI 版本
    fyne package -os windows -name "oh-my-rime-gui" -exe "$BUILD_DIR/oh-my-rime-gui.exe"
else
    # 使用 go build 并隐藏控制台窗口
    go build -ldflags="-s -w -H windowsgui" -o "$BUILD_DIR/oh-my-rime-gui.exe" .
fi

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ GUI 版本构建成功${NC}"
else
    echo -e "${RED}✗ GUI 版本构建失败${NC}"
    # GUI 构建失败不退出，因为 CLI 版本可能已经成功
fi

# 创建启动脚本
echo -e "${YELLOW}创建启动脚本...${NC}"

# CLI 启动脚本
cat > "$BUILD_DIR/oh-my-rime-cli.bat" << 'EOF'
@echo off
chcp 65001 > nul
"%~dp0oh-my-rime-cli.exe" %*
pause
EOF

# GUI 启动脚本
cat > "$BUILD_DIR/oh-my-rime-gui.bat" << 'EOF'
@echo off
start "" "%~dp0oh-my-rime-gui.exe"
EOF

# 创建安装说明
cat > "$BUILD_DIR/README.txt" << 'EOF'
Oh My Rime - Windows 版本

文件说明：
- oh-my-rime-cli.exe: 命令行版本
- oh-my-rime-gui.exe: 图形界面版本
- oh-my-rime-cli.bat: CLI 启动脚本（带暂停）
- oh-my-rime-gui.bat: GUI 启动脚本

使用方法：
1. 双击 oh-my-rime-gui.bat 启动图形界面
2. 双击 oh-my-rime-cli.bat 启动命令行界面
3. 或者直接运行对应的 .exe 文件

注意事项：
- 首次运行可能被 Windows Defender 拦截，请添加信任
- 建议将程序放置在非系统目录中
- 如需卸载，直接删除整个文件夹即可
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
zip -r "../$DIST_DIR/oh-my-rime-windows-amd64.zip" .
cd - > /dev/null

# 显示构建结果
echo -e "${GREEN}构建完成！${NC}"
echo -e "${BLUE}构建文件位置:${NC}"
echo -e "  - 构建目录: $BUILD_DIR"
echo -e "  - 发布包: $DIST_DIR/oh-my-rime-windows-amd64.zip"
echo ""
echo -e "${BLUE}文件列表:${NC}"
ls -la "$BUILD_DIR"
echo ""
echo -e "${GREEN}Windows 版本构建成功！${NC}"

# 提供使用建议
echo -e "${YELLOW}使用建议:${NC}"
echo -e "1. 将发布包发送到 Windows 机器上解压"
echo -e "2. 双击 oh-my-rime-gui.bat 启动图形界面"
echo -e "3. 或使用命令行运行 oh-my-rime-cli.exe"
echo -e "4. 首次运行可能需要在 Windows Defender 中添加信任"