#!/bin/bash

# 获取版本信息
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"

APP_VERSION=$(bash "$SCRIPT_DIR/get_version.sh" version)
APP_NAME_CONST=$(bash "$SCRIPT_DIR/get_version.sh" name)
APP_AUTHOR=$(bash "$SCRIPT_DIR/get_version.sh" author)
APP_OPENSOURCE=$(bash "$SCRIPT_DIR/get_version.sh" opensource)

echo "🚀 构建macOS GUI版本 v$APP_VERSION..."

# 构建Go应用 (Wails)
echo "🔨 使用 Wails 编译 Universal macOS 应用..."
wails build -platform darwin/universal -ldflags "-X 'oh-my-rime-cli/internal/constants.AppVersion=$APP_VERSION'" -clean

if [ $? -ne 0 ]; then
    echo "❌ Wails 编译失败"
    exit 1
fi

echo "✅ 编译成功！"

mv "build/bin/oh-my-rime-cli.app" "build/bin/Oh My Rime.app"
APP_NAME="build/bin/Oh My Rime.app"

# 签名应用包
echo "🔐 签名应用包..."
codesign --force --deep --sign - "$APP_NAME"

if [ $? -eq 0 ]; then
    echo "✅ 应用签名成功"
    
    # 验证签名
    codesign --verify --verbose "$APP_NAME"
    
    # 移除扩展属性（隔离标记）
    xattr -cr "$APP_NAME"
    echo "✅ 已移除隔离标记"
else
    echo "⚠️  应用签名失败，但继续构建..."
fi

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
    echo "  • $APP_NAME (原始应用包，在 build/bin/ 下)"
    echo "  • $DMG_NAME (分发安装包)"
    echo ""
    echo "使用方法："
    echo "  1. 双击 '$DMG_NAME' 打开安装包"
    echo "  2. 将应用拖拽到Applications文件夹"
    echo "  3. 双击应用启动GUI模式（无控制台窗口）"
    echo "  4. 或执行 dist/macos/ 下的 CLI 版本"
    echo ""
    echo "📝 注意事项："
    echo "  • 应用已使用 ad-hoc 签名，应该不会显示'已损坏'错误"
    echo "  • 如果仍有问题，请在终端运行: xattr -cr '$APP_NAME'"
else
    echo "❌ 创建DMG失败"
    rm -rf "$TEMP_DIR"
    exit 1
fi
