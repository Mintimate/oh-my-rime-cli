#!/bin/bash

set -e

echo "安装基础 GO 依赖..."
go mod tidy

echo "开始编译..."

echo "编译 Windows 版本..."
GOOS=windows GOARCH=amd64 go build -o oh-my-rime-cli.exe .

echo "编译 macOS 版本..."
GOOS=darwin GOARCH=amd64 go build -o oh-my-rime-cli-mac-amd64 .
GOOS=darwin GOARCH=arm64 go build -o oh-my-rime-cli-mac-arm64 .

echo "编译 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -o oh-my-rime-cli-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o oh-my-rime-cli-linux-arm64 .

echo "编译完成！"
