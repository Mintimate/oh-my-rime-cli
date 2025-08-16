#!/bin/bash

# 创建macOS 15风格的应用图标
echo "🎨 创建macOS 15风格图标..."

# 创建现代化的SVG图标
cat > create_modern_icon.py << 'EOF'
#!/usr/bin/env python3
import os

# 创建macOS 15风格的SVG图标 - 圆角方形
svg_content = '''<?xml version="1.0" encoding="UTF-8"?>
<svg width="1024" height="1024" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <!-- 现代渐变背景 -->
    <linearGradient id="bgGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#007AFF;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#5856D6;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#AF52DE;stop-opacity:1" />
    </linearGradient>
    
    <!-- 内部元素渐变 -->
    <linearGradient id="keyGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#FFFFFF;stop-opacity:0.95" />
      <stop offset="100%" style="stop-color:#F2F2F7;stop-opacity:0.9" />
    </linearGradient>
    
    <!-- 阴影滤镜 -->
    <filter id="shadow" x="-50%" y="-50%" width="200%" height="200%">
      <feDropShadow dx="0" dy="8" stdDeviation="12" flood-color="#000000" flood-opacity="0.15"/>
    </filter>
    
    <!-- 内部阴影 -->
    <filter id="innerShadow" x="-50%" y="-50%" width="200%" height="200%">
      <feDropShadow dx="0" dy="2" stdDeviation="4" flood-color="#000000" flood-opacity="0.1"/>
    </filter>
  </defs>
  
  <!-- macOS 15风格的圆角方形背景 -->
  <rect x="64" y="64" width="896" height="896" rx="200" ry="200" 
        fill="url(#bgGrad)" filter="url(#shadow)"/>
  
  <!-- 内部容器 -->
  <rect x="144" y="224" width="736" height="576" rx="48" ry="48" 
        fill="url(#keyGrad)" filter="url(#innerShadow)" opacity="0.95"/>
  
  <!-- 现代化键盘布局 -->
  <!-- 第一行按键 -->
  <rect x="184" y="284" width="64" height="64" rx="12" ry="12" fill="#34C759" opacity="0.9"/>
  <rect x="268" y="284" width="64" height="64" rx="12" ry="12" fill="#007AFF" opacity="0.9"/>
  <rect x="352" y="284" width="64" height="64" rx="12" ry="12" fill="#FF9500" opacity="0.9"/>
  <rect x="436" y="284" width="64" height="64" rx="12" ry="12" fill="#FF3B30" opacity="0.9"/>
  <rect x="520" y="284" width="64" height="64" rx="12" ry="12" fill="#AF52DE" opacity="0.9"/>
  <rect x="604" y="284" width="64" height="64" rx="12" ry="12" fill="#FF2D92" opacity="0.9"/>
  <rect x="688" y="284" width="64" height="64" rx="12" ry="12" fill="#5856D6" opacity="0.9"/>
  <rect x="772" y="284" width="64" height="64" rx="12" ry="12" fill="#32D74B" opacity="0.9"/>
  
  <!-- 第二行按键 -->
  <rect x="204" y="368" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="288" y="368" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="372" y="368" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="456" y="368" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="540" y="368" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="624" y="368" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="708" y="368" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  
  <!-- 第三行按键 -->
  <rect x="224" y="452" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="308" y="452" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="392" y="452" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="476" y="452" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="560" y="452" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  <rect x="644" y="452" width="64" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  
  <!-- 空格键 -->
  <rect x="284" y="536" width="456" height="64" rx="12" ry="12" fill="#8E8E93" opacity="0.7"/>
  
  <!-- 中文字符 "输" -->
  <text x="512" y="720" font-family="-apple-system, SF Pro Display, PingFang SC, Helvetica Neue, Arial" 
        font-size="180" font-weight="600" text-anchor="middle" fill="#FFFFFF" opacity="0.95">薄荷</text>
  
  <!-- 底部标识 -->
  <text x="512" y="780" font-family="-apple-system, SF Pro Display, Helvetica Neue, Arial" 
        font-size="48" font-weight="500" text-anchor="middle" fill="#FFFFFF" opacity="0.8">Rime</text>
</svg>'''

with open('icon_macos15.svg', 'w') as f:
    f.write(svg_content)

print("✅ macOS 15风格SVG图标创建成功: icon_macos15.svg")
EOF

python3 create_modern_icon.py

rm -rf create_modern_icon.py

# 转换SVG为PNG
if command -v qlmanage &> /dev/null; then
    echo "📸 转换SVG为高分辨率PNG..."
    qlmanage -t -s 1024 -o . icon_macos15.svg
    if [ -f "icon_macos15.svg.png" ]; then
        mv "icon_macos15.svg.png" "icon_macos15.png"
        echo "✅ 高分辨率PNG图标创建成功: icon_macos15.png"
        
        # 替换旧图标
        cp "icon_macos15.png" "icon.png"
        echo "✅ 已更新主图标文件"
    fi
fi

# 创建完整的ICNS图标集
if [ -f "icon.png" ] && command -v sips &> /dev/null; then
    echo "🔄 创建macOS 15风格ICNS图标..."
    
    # 清理旧的iconset
    rm -rf icon.iconset
    mkdir -p icon.iconset
    
    # 生成所有需要的尺寸，使用高质量缩放
    echo "  生成 16x16..."
    sips -z 16 16 icon.png --out icon.iconset/icon_16x16.png > /dev/null 2>&1
    echo "  生成 32x32..."
    sips -z 32 32 icon.png --out icon.iconset/icon_16x16@2x.png > /dev/null 2>&1
    sips -z 32 32 icon.png --out icon.iconset/icon_32x32.png > /dev/null 2>&1
    echo "  生成 64x64..."
    sips -z 64 64 icon.png --out icon.iconset/icon_32x32@2x.png > /dev/null 2>&1
    echo "  生成 128x128..."
    sips -z 128 128 icon.png --out icon.iconset/icon_128x128.png > /dev/null 2>&1
    echo "  生成 256x256..."
    sips -z 256 256 icon.png --out icon.iconset/icon_128x128@2x.png > /dev/null 2>&1
    sips -z 256 256 icon.png --out icon.iconset/icon_256x256.png > /dev/null 2>&1
    echo "  生成 512x512..."
    sips -z 512 512 icon.png --out icon.iconset/icon_256x256@2x.png > /dev/null 2>&1
    sips -z 512 512 icon.png --out icon.iconset/icon_512x512.png > /dev/null 2>&1
    echo "  生成 1024x1024..."
    sips -z 1024 1024 icon.png --out icon.iconset/icon_512x512@2x.png > /dev/null 2>&1
    
    # 创建icns文件
    iconutil -c icns icon.iconset
    
    if [ -f "icon.icns" ]; then
        echo "✅ macOS 15风格ICNS图标创建成功: icon.icns"
        rm -rf icon.iconset
        echo "🎨 图标特色："
        echo "  • 现代圆角方形设计（macOS 15风格）"
        echo "  • 渐变色彩背景（蓝紫色系）"
        echo "  • 彩色键盘布局"
        echo "  • 中文'输'字标识"
        echo "  • 支持Retina显示器"
    else
        echo "❌ 创建ICNS失败"
        exit 1
    fi
else
    echo "❌ 无法创建图标，缺少必要工具"
    exit 1
fi

echo ""
echo "🎉 macOS 15风格图标创建完成！"