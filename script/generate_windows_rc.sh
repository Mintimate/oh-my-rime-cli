#!/bin/bash

# 动态生成 Windows 资源文件

# 获取版本信息
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

APP_VERSION=$(bash "$SCRIPT_DIR/get_version.sh" version)
APP_NAME=$(bash "$SCRIPT_DIR/get_version.sh" name)
APP_AUTHOR=$(bash "$SCRIPT_DIR/get_version.sh" author)
APP_OPENSOURCE=$(bash "$SCRIPT_DIR/get_version.sh" opensource)
APP_DESC=$(bash "$SCRIPT_DIR/get_version.sh" desc)

# 将版本号转换为逗号分隔的格式 (例如: 1.1.1 -> 1,1,1,0)
VERSION_COMMA=$(echo "$APP_VERSION" | sed 's/\./,/g'),0

# 生成资源文件
cat > script/build/windows/resources/app.rc << EOF
#include <windows.h>

// 应用程序图标
IDI_ICON1 ICON "app.ico"

// 版本信息
VS_VERSION_INFO VERSIONINFO
FILEVERSION $VERSION_COMMA
PRODUCTVERSION $VERSION_COMMA
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
            VALUE "CompanyName", "$APP_AUTHOR"
            VALUE "FileDescription", "$APP_DESC"
            VALUE "FileVersion", "$APP_VERSION.0"
            VALUE "InternalName", "oh-my-rime"
            VALUE "LegalCopyright", "Copyright (C) 2024 $APP_AUTHOR"
            VALUE "OriginalFilename", "oh-my-rime.exe"
            VALUE "ProductName", "$APP_NAME"
            VALUE "ProductVersion", "$APP_VERSION.0"
        END
    END
    BLOCK "VarFileInfo"
    BEGIN
        VALUE "Translation", 0x409, 1200
    END
END
EOF

echo "✅ Windows 资源文件已生成: script/build/windows/resources/app.rc"