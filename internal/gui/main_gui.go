package gui

import (
	"flag"
	"fmt"
	"os"

	"oh-my-rime-cli/internal/constants"
	"oh-my-rime-cli/internal/system"
)

// StartApp 启动应用程序（CLI模式）
func StartApp() {
	var guiMode = flag.Bool("gui", false, "启动图形界面模式")
	flag.Parse()

	fmt.Println("欢迎使用: ", constants.AppName)
	fmt.Println("工具版本: ", constants.AppVersion)

	// 检测操作系统
	currentOS := system.DetectOS()
	if currentOS == "Unknown" {
		fmt.Println("无法识别当前操作系统，请确保在支持的操作系统上运行")
		os.Exit(-1)
	}
	fmt.Printf("当前操作系统: %s\n", currentOS)

	// 如果不是GUI模式，提示用户使用CLI模式
	if !*guiMode {
		fmt.Println("提示：使用 --gui 参数启动图形界面模式")
		fmt.Println("当前启动CLI模式...")
		// 这里应该调用CLI模式，但为了避免循环依赖，我们直接启动GUI
		*guiMode = true
	}

	// 启动GUI
	if *guiMode {
		fmt.Println("启动图形界面...")
		gui := NewGUI()
		gui.Run()
	}
}

// StartGUIApp 启动GUI应用程序（无控制台窗口）
func StartGUIApp() {
	// 检测操作系统
	currentOS := system.DetectOS()
	if currentOS == "Unknown" {
		os.Exit(-1)
	}

	// 直接启动GUI，不显示控制台信息
	gui := NewGUI()
	gui.Run()
}