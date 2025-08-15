package main

import (
	"flag"
	"fmt"
	"os"
)

// 启动应用程序的入口函数
func startApp() {
	var guiMode = flag.Bool("gui", false, "启动图形界面模式")
	flag.Parse()

	if *guiMode {
		// 启动 GUI 模式
		gui := NewGUI()
		gui.Run()
		return
	}

	// 启动原有的命令行模式
	startCLI()
}

// 原有的命令行模式
func startCLI() {
	fmt.Println("欢迎使用: ", AppName)
	fmt.Println("工具版本: ", AppVersion)
	fmt.Println("提示: 使用 --gui 参数可启动图形界面")

	// 检测操作系统
	currentOS := detectOS()
	if currentOS == "Unknown" {
		fmt.Println("无法识别当前操作系统，请确保在支持的操作系统上运行")
		os.Exit(-1)
	}
	fmt.Printf("当前操作系统: %s\n", currentOS)

	// 启动原有的交互式菜单
	runInteractiveMenu()
}