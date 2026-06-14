#!/bin/bash

# Windows 构建脚本
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
VERSION="1.0.0"
BUILD_DIR="build/windows"
DIST_DIR="dist/windows"

echo -e "${BLUE}开始构建 Windows 版本...${NC}"

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
if ! go env GOOS=windows GOARCH=amd64 go version &> /dev/null; then
    echo -e "${YELLOW}警告: 可能不支持 Windows 交叉编译${NC}"
fi

# 创建 Windows 图标资源文件
echo -e "${YELLOW}准备 Windows 图标资源...${NC}"

# 创建资源文件目录
mkdir -p "$BUILD_DIR/resources"

# 创建 .rc 资源文件（Windows 资源脚本）
cat > "$BUILD_DIR/resources/app.rc" << 'EOF'
#include <windows.h>

// 应用程序图标
IDI_ICON1 ICON "app.ico"

// 版本信息
VS_VERSION_INFO VERSIONINFO
FILEVERSION 1,0,0,0
PRODUCTVERSION 1,0,0,0
FILEFLAGSMASK VS_FFI_FILEFLAGSMASK
FILEFLAGS 0x0L
FILEOS VOS__WINDOWS32
FILETYPE VFT_APP
FILESUBTYPE VFT2_UNKNOWN
BEGIN
    BLOCK "StringFileInfo"
    BEGIN
        BLOCK "040904b0"
        BEGIN
            VALUE "CompanyName", "Oh My Rime"
            VALUE "FileDescription", "Oh My Rime - Rime 输入法配置工具"
            VALUE "FileVersion", "1.0.0.0"
            VALUE "InternalName", "oh-my-rime"
            VALUE "LegalCopyright", "Copyright (C) 2024"
            VALUE "OriginalFilename", "oh-my-rime.exe"
            VALUE "ProductName", "Oh My Rime"
            VALUE "ProductVersion", "1.0.0.0"
        END
    END
    BLOCK "VarFileInfo"
    BEGIN
        VALUE "Translation", 0x409, 1200
    END
END
EOF

# 创建一个简单的 ICO 图标文件（如果不存在的话）
if [ ! -f "$BUILD_DIR/resources/app.ico" ]; then
    echo -e "${YELLOW}创建默认图标...${NC}"
    # 这里创建一个简单的 16x16 像素的 ICO 文件
    # 实际使用时，你应该替换为真正的图标文件
    cat > "$BUILD_DIR/resources/app.ico" << 'EOF'
EOF
    # 注意：这里应该是二进制 ICO 文件内容
    # 为了简化，我们先创建一个占位符
    echo "请将真正的 app.ico 图标文件放置到 $BUILD_DIR/resources/ 目录中"
fi

# 编译资源文件（如果有 windres 工具）
if command -v windres &> /dev/null; then
    echo -e "${YELLOW}编译 Windows 资源文件...${NC}"
    windres -i "$BUILD_DIR/resources/app.rc" -o "$BUILD_DIR/resources/app.syso"
    # 将编译后的资源文件复制到项目根目录，Go 会自动包含它
    cp "$BUILD_DIR/resources/app.syso" ./
else
    echo -e "${YELLOW}警告: windres 未找到，跳过资源文件编译${NC}"
    echo -e "${YELLOW}图标可能不会显示在最终的 exe 文件中${NC}"
fi

# 设置 Go 构建环境变量
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1

# 构建 CLI 版本
echo -e "${YELLOW}构建 CLI 版本...${NC}"
go build -ldflags="-s -w -H windowsgui" -o "$BUILD_DIR/oh-my-rime-cli.exe" .

# 构建 GUI 版本（隐藏控制台窗口）
echo -e "${YELLOW}构建 GUI 版本...${NC}"
go build -ldflags="-s -w -H windowsgui" -tags gui -o "$BUILD_DIR/oh-my-rime-gui.exe" .

# 清理临时资源文件
if [ -f "./app.syso" ]; then
    rm -f "./app.syso"
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