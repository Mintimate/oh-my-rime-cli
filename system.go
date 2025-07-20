package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// 判断当前系统，Linux? Windows? MacOS?
func detectOS() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows_NT"
	case "linux":
		return "Linux"
	case "darwin":
		return "Darwin"
	}
	return "Unknown"
}

// 根据系统判断目标目录
func getTargetDir() string {
	switch detectOS() {
	case "Windows_NT":
		return getWindowsTargetDir()
	case "Linux":
		return getLinuxTargetDir()
	case "Darwin":
		return getDarwinTargetDir()
	}
	return "Unknown"
}

// Windows系统目标目录获取
func getWindowsTargetDir() string {
	// 首先尝试从注册表读取Rime用户目录
	if rimeDir, err := getRimeUserDirFromRegistry(); err == nil {
		if rimeDir == "" {
			fmt.Println("从注册表获取到的Rime用户目录为空，使用默认目录:", filepath.Join(os.Getenv("APPDATA"), "Rime"))
			return filepath.Join(os.Getenv("APPDATA"), "Rime")
		}
		fmt.Println("从注册表获取到的Rime用户目录:", rimeDir)
		return rimeDir
	} else {
		fmt.Println("从注册表读取Rime目录失败: ", err)
		fmt.Println("使用默认目录...")
	}
	// 如果注册表读取失败，回退到默认目录
	return filepath.Join(os.Getenv("APPDATA"), "Rime")
}

// Linux系统目标目录获取
func getLinuxTargetDir() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n==============================")
	fmt.Println(" 请选择 Rime 配置目录类型 ")
	fmt.Println("==============================")
	fmt.Println("1️⃣  iBus")
	fmt.Println("2️⃣  Fcitx5")
	fmt.Println("3️⃣  Fcitx5-Flatpak")
	fmt.Println("------------------------------")
	fmt.Print("请输入选项（1/2/3）：")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1", "":
		targetDir := filepath.Join(os.Getenv("HOME"), ".config", "rime")
		fmt.Println("目标地址: ", targetDir)
		return targetDir
	case "2":
		targetDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime")
		fmt.Println("目标地址: ", targetDir)
		return targetDir
	case "3":
		targetDir := filepath.Join(os.Getenv("HOME"), ".var", "app", "org.fcitx.Fcitx5", "data", "fcitx5", "rime")
		fmt.Println("目标地址: ", targetDir)
		return targetDir
	default:
		fmt.Println("无效选择，程序已退出。")
		os.Exit(1)
	}
	return ""
}

// Darwin系统目标目录获取
func getDarwinTargetDir() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n==============================")
	fmt.Println(" 请选择 Rime 配置目录类型 ")
	fmt.Println("==============================")
	fmt.Println("1️⃣  鼠须管")
	fmt.Println("2️⃣  小企鹅")
	fmt.Println("------------------------------")
	fmt.Print("请输入选项（1/2）：")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1", "":
		targetDir := filepath.Join(os.Getenv("HOME"), "Library", "Rime")
		fmt.Println("目标地址: ", targetDir)
		return targetDir
	case "2":
		targetDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime")
		fmt.Println("目标地址: ", targetDir)
		return targetDir
	default:
		fmt.Println("无效选择，程序已退出。")
		os.Exit(1)
	}
	return ""
}
