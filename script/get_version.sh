#!/bin/bash

# 从 constants.go 读取版本信息的通用脚本

CONSTANTS_FILE="internal/constants/constants.go"

if [ ! -f "$CONSTANTS_FILE" ]; then
    echo "错误: 找不到 $CONSTANTS_FILE" >&2
    exit 1
fi

# 提取版本号
get_app_version() {
    grep -o 'AppVersion.*=.*"[^"]*"' "$CONSTANTS_FILE" | sed 's/.*"\([^"]*\)".*/\1/'
}

# 提取应用名称
get_app_name() {
    grep -o 'AppName.*=.*"[^"]*"' "$CONSTANTS_FILE" | sed 's/.*"\([^"]*\)".*/\1/'
}

# 提取作者
get_app_author() {
    grep 'APPAuthor.*=' "$CONSTANTS_FILE" | grep -v 'APPAuthorBilibili' | grep -o '"[^"]*"' | sed 's/"//g'
}

# 提取开源地址
get_app_opensource() {
    grep -o 'APPOpenSource.*=.*"[^"]*"' "$CONSTANTS_FILE" | sed 's/.*"\([^"]*\)".*/\1/'
}

# 提取描述
get_app_desc() {
    grep -o 'AppDesc.*=.*"[^"]*"' "$CONSTANTS_FILE" | sed 's/.*"\([^"]*\)".*/\1/'
}

# 根据参数返回对应值
case "$1" in
    "version"|"--version"|"-v")
        get_app_version
        ;;
    "name"|"--name"|"-n")
        get_app_name
        ;;
    "author"|"--author"|"-a")
        get_app_author
        ;;
    "opensource"|"--opensource"|"-o")
        get_app_opensource
        ;;
    "desc"|"--desc"|"-d")
        get_app_desc
        ;;
    "all"|"--all")
        echo "AppName: $(get_app_name)"
        echo "AppVersion: $(get_app_version)"
        echo "AppAuthor: $(get_app_author)"
        echo "AppOpenSource: $(get_app_opensource)"
        echo "AppDesc: $(get_app_desc)"
        ;;
    *)
        echo "用法: $0 [version|name|author|opensource|desc|all]"
        echo "示例:"
        echo "  $0 version    # 获取版本号"
        echo "  $0 name       # 获取应用名称"
        echo "  $0 all        # 获取所有信息"
        exit 1
        ;;
esac