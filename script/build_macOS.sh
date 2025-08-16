#!/bin/bash

# 获取版本信息
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

APP_VERSION=$(bash "$SCRIPT_DIR/get_version.sh" version)
APP_NAME_CONST=$(bash "$SCRIPT_DIR/get_version.sh" name)
APP_AUTHOR=$(bash "$SCRIPT_DIR/get_version.sh" author)
APP_OPENSOURCE=$(bash "$SCRIPT_DIR/get_version.sh" opensource)

echo "🚀 构建macOS GUI版本 v$APP_VERSION..."

# 构建Go应用
echo "🔨 编译Go应用..."
go build -o oh-my-rime-gui .

if [ $? -ne 0 ]; then
    echo "❌ Go编译失败"
    exit 1
fi

echo "✅ 编译成功！"

# 创建macOS应用包结构
APP_NAME="Oh My Rime.app"
echo "📦 创建应用包: $APP_NAME"

# 清理旧的应用包
rm -rf "$APP_NAME"

mkdir -p "$APP_NAME/Contents/MacOS"
mkdir -p "$APP_NAME/Contents/Resources"

# 移动可执行文件到应用包
mv oh-my-rime-gui "$APP_NAME/Contents/MacOS/oh-my-rime-gui"

# 创建简单的图标（使用系统默认图标）
# 如果有icon.icns就使用，否则跳过
if [ -f "script/build/macOS/icon.icns" ]; then
    cp "script/build/macOS/icon.icns" "$APP_NAME/Contents/Resources/"
    ICON_KEY="    <key>CFBundleIconFile</key>
    <string>icon</string>"
else
    ICON_KEY=""
fi

# 创建Info.plist文件
cat > "$APP_NAME/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>oh-my-rime-gui</string>
    <key>CFBundleIdentifier</key>
    <string>com.rime.oh-my-rime</string>
    <key>CFBundleName</key>
    <string>Oh My Rime</string>
    <key>CFBundleDisplayName</key>
    <string>$APP_NAME_CONST</string>
    <key>CFBundleVersion</key>
    <string>$APP_VERSION</string>
    <key>CFBundleShortVersionString</key>
    <string>$APP_VERSION</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
$ICON_KEY
    <key>LSUIElement</key>
    <false/>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSAllowsArbitraryLoads</key>
        <true/>
    </dict>
</dict>
</plist>
EOF

echo "✅ 应用包创建完成"

# 创建简单的DMG
echo "💿 创建DMG安装包..."

DMG_NAME="Oh-My-Rime-v$APP_VERSION.dmg"

# 清理旧文件
rm -f "$DMG_NAME"

# 创建临时目录
TEMP_DIR="dmg_temp"
rm -rf "$TEMP_DIR"
mkdir "$TEMP_DIR"

# 复制应用到临时目录
cp -R "$APP_NAME" "$TEMP_DIR/"

# 创建Applications链接
ln -s /Applications "$TEMP_DIR/Applications"

# 创建安装说明
cat > "$TEMP_DIR/安装说明.txt" << EOF
$APP_NAME_CONST - Rime输入法配置管理工具

安装方法：
将 "Oh My Rime.app" 拖拽到 "Applications" 文件夹

使用方法：
• 双击应用图标 -> GUI模式（无控制台窗口）
• 终端运行可执行文件 -> CLI模式

项目地址：$APP_OPENSOURCE
作者：$APP_AUTHOR
版本：v$APP_VERSION
EOF

# 创建DMG（简化版，无复杂布局）
hdiutil create -srcfolder "$TEMP_DIR" -volname "Oh My Rime" -format UDZO \
    -imagekey zlib-level=9 "$DMG_NAME"

if [ $? -eq 0 ]; then
    echo "✅ DMG创建成功: $DMG_NAME"
    
    # 清理临时文件
    rm -rf "$TEMP_DIR"
    
    echo ""
    echo "🎉 构建完成！"
    echo "📁 生成文件："
    echo "  • $APP_NAME (应用包)"
    echo "  • $DMG_NAME (安装包)"
    echo ""
    echo "使用方法："
    echo "  1. 双击 '$DMG_NAME' 打开安装包"
    echo "  2. 将应用拖拽到Applications文件夹"
    echo "  3. 双击应用启动GUI模式（无控制台窗口）"
    echo "  4. 或使用 ./oh-my-rime-test 启动CLI模式"
else
    echo "❌ 创建DMG失败"
    rm -rf "$TEMP_DIR"
    exit 1
fi