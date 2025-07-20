package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// 自定义更新函数
func customUpdate() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n==============================")
	fmt.Println("自定义更新功能: ")
	fmt.Println("粘贴方案打包的 zip 文件 URL => 将下载并替换当前 Rime 配置目录下的文件")
	fmt.Println("粘贴模型的 gram 文件 URL => 将下载并替换当前 Rime 配置目录下的同名文件")
	fmt.Println("URL 下载失败不会更新任何文件，本质是同名文件覆盖")
	fmt.Println("==============================")
	fmt.Println("请粘贴 URL （例如：https://github.com/Mintimate/oh-my-rime/archive/refs/heads/main.zip）：")
	customUrl, _ := reader.ReadString('\n')
	customUrl = strings.TrimSpace(customUrl)

	customData := download(customUrl)
	if customData == nil {
		fmt.Println("下载自定义方案失败，请检查 URL 或网络连接")
		return
	}

	targetDir := getTargetDir()

	// 判断文件类型
	if strings.HasSuffix(customUrl, ".zip") {
		// 如果是 zip 文件，更新主方案
		if err := updateMainScheme(customData, targetDir); err != nil {
			fmt.Printf("更新自定义方案失败: %v\n", err)
		}
	} else if strings.HasSuffix(customUrl, ".gram") {
		// 如果是 gram 文件，更新模型
		if err := updateModel(customData, targetDir); err != nil {
			fmt.Printf("更新自定义模型失败: %v\n", err)
		}
	} else {
		fmt.Println("不支持的文件类型，请提供 zip 或 gram 文件的 URL")
	}
}

// 显示主菜单
func showMenu() {
	fmt.Println("\n==============================")
	fmt.Println("   Rime 配置更新工具菜单   ")
	fmt.Println("原理: ")
	fmt.Println(" 1. 下载最新的方案或模型文件 ")
	fmt.Println(" 2. 替换当前 Rime配置目录下的同名文件 ")
	fmt.Println("==============================")
	fmt.Println("1️⃣  更新薄荷方案")
	fmt.Println("2️⃣  更新万象模型")
	fmt.Println("3️⃣  更新万象词库（薄荷使用的 Lite 版本）")
	fmt.Println("4️⃣  自定义更新（适用于其他方案）")
	fmt.Println("b  打开薄荷输入法文档")
	fmt.Println("q  退出")
	fmt.Println("------------------------------")
	fmt.Print("请输入选项（1/2/3/4/q）：")
}

// 处理用户选择的操作
func handleUserChoice(choice string) bool {
	switch choice {
	case "1":
		return handleUpdateMainScheme()
	case "2":
		return handleUpdateModel()
	case "3":
		return handleUpdateDict()
	case "4":
		customUpdate()
		return true
	case "b":
		fmt.Println("打开薄荷输入法文档...")
		openUrlBrowser(AppURL)
		return true
	case "q":
		fmt.Println("感谢使用！记得更新后，重新部署方案以使更改生效")
		return false
	default:
		fmt.Println("无效选项，请重新输入")
		return true
	}
}

// 处理更新主方案
func handleUpdateMainScheme() bool {
	rimeZip := download(OhMyRimeRepo)
	if rimeZip == nil {
		fmt.Println("下载主方案失败，请检查网络连接或稍后重试")
		return true
	}

	targetDir := getTargetDir()
	if err := updateMainScheme(rimeZip, targetDir); err != nil {
		fmt.Printf("更新主方案失败: %v\n", err)
	}
	return true
}

// 处理更新模型
func handleUpdateModel() bool {
	rimeGram := download(WanXiangGRA)
	if rimeGram == nil {
		fmt.Println("下载模型失败，请检查网络连接或稍后重试")
		return true
	}

	targetDir := getTargetDir()
	if err := updateModel(rimeGram, targetDir); err != nil {
		fmt.Printf("更新模型失败: %v\n", err)
	}
	return true
}

// 处理更新词库
func handleUpdateDict() bool {
	targetDir := getTargetDir()
	rimeZip := download(OhMyRimeRepo)
	if rimeZip == nil {
		fmt.Println("下载词库失败，请检查网络连接或稍后重试")
		return true
	}

	if err := updateDict(rimeZip, targetDir); err != nil {
		fmt.Printf("更新词库失败: %v\n", err)
	}
	return true
}

func main() {
	fmt.Println("欢迎使用: ", AppName)
	fmt.Println("工具版本: ", AppVersion)

	// 检测操作系统
	currentOS := detectOS()
	if currentOS == "Unknown" {
		fmt.Println("无法识别当前操作系统，请确保在支持的操作系统上运行")
		os.Exit(-1)
	}
	fmt.Printf("当前操作系统: %s\n", currentOS)

	reader := bufio.NewReader(os.Stdin)

	for {
		showMenu()

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if !handleUserChoice(input) {
			break
		}
	}
}
