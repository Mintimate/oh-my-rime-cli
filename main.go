package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// 薄荷仓库地址
	OhMyRimeRepo = "https://cnb.cool/Mintimate/rime/oh-my-rime/-/releases/download/latest/oh-my-rime.zip"
	//万象模型镜像
	WanXiangGRA = "https://cnb.cool/Mintimate/rime/oh-my-rime/-/releases/download/latest/wanxiang-lts-zh-hans.gram"
)

// 进度条读取器
type ProgressReader struct {
	io.Reader
	Total      int64
	Downloaded int64
	StartTime  time.Time
	LastUpdate time.Time
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Downloaded += int64(n)

	// 每100ms更新一次进度条
	now := time.Now()
	if now.Sub(pr.LastUpdate) > 100*time.Millisecond || err == io.EOF {
		pr.LastUpdate = now
		pr.printProgress()
	}

	return n, err
}

func (pr *ProgressReader) printProgress() {
	if pr.Total <= 0 {
		fmt.Printf("\r下载中... %s", formatBytes(pr.Downloaded))
		return
	}

	percentage := float64(pr.Downloaded) / float64(pr.Total) * 100
	elapsed := time.Since(pr.StartTime)

	// 计算下载速度
	speed := float64(pr.Downloaded) / elapsed.Seconds()

	// 计算剩余时间
	var eta string
	if speed > 0 && pr.Downloaded < pr.Total {
		remaining := float64(pr.Total-pr.Downloaded) / speed
		eta = fmt.Sprintf(" ETA: %s", formatDuration(time.Duration(remaining)*time.Second))
	}

	// 绘制进度条
	barWidth := 40
	filled := int(percentage * float64(barWidth) / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\r[%s] %.1f%% %s/%s %s/s%s",
		bar,
		percentage,
		formatBytes(pr.Downloaded),
		formatBytes(pr.Total),
		formatBytes(int64(speed)),
		eta)
}

// 格式化字节大小
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// 格式化时间
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), d.Seconds()-d.Truncate(time.Minute).Seconds())
	}
	return fmt.Sprintf("%.0fh%.0fm", d.Hours(), d.Minutes()-d.Truncate(time.Hour).Minutes())
}

// 私有方法: 下载（带进度条）
func download(url string) []byte {
	fmt.Printf("正在下载: %s\n", url)

	// 创建HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("\n请求失败: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("\nHTTP错误: %s\n", resp.Status)
		return nil
	}

	// 获取文件大小
	contentLength := resp.Header.Get("Content-Length")
	var totalSize int64
	if contentLength != "" {
		totalSize, _ = strconv.ParseInt(contentLength, 10, 64)
	}

	// 创建进度读取器
	progressReader := &ProgressReader{
		Reader:     resp.Body,
		Total:      totalSize,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
	}

	// 读取响应内容到内存
	data, err := io.ReadAll(progressReader)
	if err != nil {
		fmt.Printf("\n读取内容失败: %v\n", err)
		return nil
	}

	fmt.Printf("\n✅ 下载完成! 总大小: %s\n", formatBytes(int64(len(data))))
	return data
}

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
	case "Linux":
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
		if choice == "1" || choice == "" {
			fmt.Println("目标地址: ", filepath.Join(os.Getenv("HOME"), ".config", "rime"))
			return filepath.Join(os.Getenv("HOME"), ".config", "rime")
		} else if choice == "2" {
			fmt.Println("目标地址: ", filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime"))
			return filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime")
		} else if choice == "3" {
			fmt.Println("目标地址: ", filepath.Join(os.Getenv("HOME"), ".var", "app", "org.fcitx.Fcitx5", "data", "fcitx5", "rime"))
			return filepath.Join(os.Getenv("HOME"), ".var", "app", "org.fcitx.Fcitx5", "data", "fcitx5", "rime")
		} else {
			fmt.Println("无效选择，程序已退出。")
			os.Exit(1)
		}
	case "Darwin":
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
		if choice == "1" || choice == "" {
			fmt.Println("目标地址: ", filepath.Join(os.Getenv("HOME"), "Library", "Rime"))
			return filepath.Join(os.Getenv("HOME"), "Library", "Rime")
		} else if choice == "2" {
			fmt.Println("目标地址: ", filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime"))
			return filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime")
		} else {
			fmt.Println("无效选择，程序已退出。")
			os.Exit(1)
		}
	}
	return "Unknown"
}

// 更新方案
func updateMainScheme(rime_zip []byte, targetDir string) error {
	fmt.Println("正在更新主方案...")

	// 检查zip数据是否有效
	if rime_zip == nil || len(rime_zip) == 0 {
		return fmt.Errorf("zip数据无效")
	}

	// 创建目标目录（如果不存在）
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 从字节数组创建zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(rime_zip), int64(len(rime_zip)))
	if err != nil {
		return fmt.Errorf("读取zip文件失败: %v", err)
	}

	// 遍历zip文件中的每个文件
	for _, file := range zipReader.File {
		// 构建目标文件路径
		targetPath := filepath.Join(targetDir, file.Name)

		if file.FileInfo().IsDir() {
			// 创建目录
			if err := os.MkdirAll(targetPath, file.FileInfo().Mode()); err != nil {
				fmt.Printf("创建目录失败 %s: %v\n", targetPath, err)
				return err
			}
			fmt.Printf("创建目录: %s\n", targetPath)
		} else {
			// 创建父目录
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				fmt.Printf("创建父目录失败 %s: %v\n", filepath.Dir(targetPath), err)
			}

			// 解压文件
			if err := extractFile(file, targetPath); err != nil {
				fmt.Printf("解压文件失败 %s: %v\n", targetPath, err)
				continue
			}
			fmt.Printf("解压文件: %s\n", targetPath)
		}
	}

	fmt.Println("✅ 主方案更新完成！")
	return nil
}

// 辅助函数：解压单个文件
func extractFile(file *zip.File, targetPath string) error {
	// 打开zip文件中的文件
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// 创建目标文件（覆盖同名文件）
	outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 复制文件内容
	_, err = io.Copy(outFile, rc)
	return err
}

func updateModel(rime_gram []byte, targetDir string) {
	fmt.Println("正在更新模型...")
	// 检查模型数据是否有效
	if rime_gram == nil || len(rime_gram) == 0 {
		fmt.Println("模型数据无效，更新失败")
		return
	}
	// 覆盖保存目录内的模型文件
	modelPath := filepath.Join(targetDir, "wanxiang-lts-zh-hans.gram")
	if err := os.WriteFile(modelPath, rime_gram, 0644); err != nil {
		fmt.Printf("更新模型失败: %v\n", err)
		return
	}
	fmt.Println("✅ 模型更新完成！")
}

func updateDict(rime_zip []byte, targetDir string) error {
	fmt.Println("正在更新词库...")

	// 检查zip数据是否有效
	if rime_zip == nil || len(rime_zip) == 0 {
		return fmt.Errorf("zip数据无效")
	}

	// 创建目标词库目录
	dictsTargetDir := filepath.Join(targetDir, "dicts")
	if err := os.MkdirAll(dictsTargetDir, 0755); err != nil {
		return fmt.Errorf("创建词库目录失败: %v", err)
	}

	// 从字节数组创建zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(rime_zip), int64(len(rime_zip)))
	if err != nil {
		return fmt.Errorf("读取zip文件失败: %v", err)
	}

	// 遍历zip文件中的每个文件，只处理dicts目录下的文件
	for _, file := range zipReader.File {
		// 检查文件是否在dicts目录下
		if !strings.HasPrefix(file.Name, "dicts/") {
			continue
		}

		// 计算相对于dicts目录的路径
		relativePath := strings.TrimPrefix(file.Name, "dicts/")
		if relativePath == "" {
			// 这是dicts目录本身，跳过
			continue
		}

		// 构建目标文件路径
		targetPath := filepath.Join(dictsTargetDir, relativePath)

		if file.FileInfo().IsDir() {
			// 创建子目录
			if err := os.MkdirAll(targetPath, file.FileInfo().Mode()); err != nil {
				fmt.Printf("创建词库子目录失败 %s: %v\n", targetPath, err)
				continue
			}
			fmt.Printf("创建词库目录: %s\n", targetPath)
		} else {
			// 创建父目录
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				fmt.Printf("创建父目录失败 %s: %v\n", filepath.Dir(targetPath), err)
				continue
			}

			// 解压文件
			if err := extractFile(file, targetPath); err != nil {
				fmt.Printf("解压词库文件失败 %s: %v\n", targetPath, err)
				continue
			}
			fmt.Printf("更新词库文件: %s\n", targetPath)
		}
	}

	fmt.Println("✅ 词库更新完成！")
	return nil
}

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
	customZip := download(customUrl)
	if customZip == nil {
		fmt.Println("下载自定义方案失败，请检查 URL 或网络连接")
		return
	}
	targetDir := getTargetDir()
	// 判断文件类型
	if strings.HasSuffix(customUrl, ".zip") {
		// 如果是 zip 文件，更新主方案
		if err := updateMainScheme(customZip, targetDir); err != nil {
			fmt.Printf("更新自定义方案失败: %v\n", err)
			return
		}
	} else if strings.HasSuffix(customUrl, ".gram") {
		// 如果是 gram 文件，更新模型
		updateModel(customZip, targetDir)
		return
	} else {
		fmt.Println("不支持的文件类型，请提供 zip 或 gram 文件的 URL")
		return
	}
}

func showMenu() {
	fmt.Println("\n==============================")
	fmt.Println("   Rime 配置更新工具菜单   ")
	fmt.Println("==============================")
	fmt.Println("1️⃣  更新薄荷方案")
	fmt.Println("2️⃣  更新万象模型")
	fmt.Println("3️⃣  更新万象词库（薄荷使用的 Lite 版本）")
	fmt.Println("4️⃣  自定义更新（适用于其他方案）")
	fmt.Println("q  退出")
	fmt.Println("------------------------------")
	fmt.Print("请输入选项（1/2/3/4/q）：")
}

func main() {
	fmt.Println("欢迎使用 Rime oh-my-rime 配置更新工具")
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

		switch input {
		case "1":
			// 更新主方案
			rimeZip := download(OhMyRimeRepo)
			if rimeZip == nil {
				fmt.Println("下载主方案失败，请检查网络连接或稍后重试")
				continue
			}
			targetDir := getTargetDir()
			if err := updateMainScheme(rimeZip, targetDir); err != nil {
				fmt.Printf("更新主方案失败: %v\n", err)
			}
		case "2":
			// 更新模型
			rimeGram := download(WanXiangGRA)
			if rimeGram == nil {
				fmt.Println("下载模型失败，请检查网络连接或稍后重试")
				continue
			}
			targetDir := getTargetDir()
			// 更新模型
			updateModel(rimeGram, targetDir)
		case "3":
			targetDir := getTargetDir()
			rimeZip := download(OhMyRimeRepo)
			if rimeZip == nil {
				fmt.Println("下载词库失败，请检查网络连接或稍后重试")
				continue
			}
			// 更新词库
			if err := updateDict(rimeZip, targetDir); err != nil {
				fmt.Printf("更新词库失败: %v\n", err)
			}
		case "4":
			// 自定义更新
			customUpdate()
		case "q":
			fmt.Println("感谢使用记得更新后，重新部署方案以使更改生效")
			return
		default:
			fmt.Println("无效选项，请重新输入")
		}
	}
}
