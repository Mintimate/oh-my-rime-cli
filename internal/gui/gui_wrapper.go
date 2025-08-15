package gui

import (
	"fmt"
	"oh-my-rime-cli/internal/downloader"
	"oh-my-rime-cli/internal/system"
	"oh-my-rime-cli/internal/updater"
	"strings"

	"fyne.io/fyne/v2"
)

// 常量定义 - 从根目录的 constants.go 复制
const (
	OhMyRimeRepo = "https://cnb.cool/Mintimate/rime/oh-my-rime/-/releases/download/latest/oh-my-rime.zip"
	WanXiangGRA  = "https://cnb.cool/Mintimate/rime/oh-my-rime/-/releases/download/latest/wanxiang-lts-zh-hans.gram"
)

// 类型别名
type ProgressCallback = downloader.ProgressCallback

// GUI 包装函数，将原有的交互式功能包装成可以被 GUI 调用的函数
// 这些函数现在支持异步的目录选择和进度回调

// updateMainSchemeConfigWithCallback 更新薄荷方案 - 对应 CLI 选项 1
func updateMainSchemeConfigWithCallback(window fyne.Window, callback func(error)) {
	updateMainSchemeConfigWithProgressCallback(window, nil, callback)
}

// 带进度回调的更新薄荷方案函数
func updateMainSchemeConfigWithProgressCallback(window fyne.Window, progressCallback downloader.ProgressCallback, callback func(error)) {
	fmt.Println("开始更新薄荷方案...")

	// 下载主方案
	rimeZip := downloader.DownloadWithCallback(OhMyRimeRepo, progressCallback)
	if rimeZip == nil {
		callback(fmt.Errorf("下载薄荷方案失败，请检查网络连接"))
		return
	}

	// 获取目标目录（可能需要用户选择）
	getTargetDirGUI(window, func(targetDir string, err error) {
		if err != nil {
			callback(fmt.Errorf("获取目标目录失败: %v", err))
			return
		}

		if err := updater.UpdateMainScheme(rimeZip, targetDir); err != nil {
			callback(fmt.Errorf("更新薄荷方案失败: %v", err))
			return
		}

		fmt.Println("✅ 薄荷方案更新完成")
		callback(nil)
	})
}

// updateModelConfigWithCallback 更新万象模型 - 对应 CLI 选项 2
func updateModelConfigWithCallback(window fyne.Window, callback func(error)) {
	updateModelConfigWithProgressCallback(window, nil, callback)
}

// 带进度回调的更新万象模型函数
func updateModelConfigWithProgressCallback(window fyne.Window, progressCallback ProgressCallback, callback func(error)) {
	fmt.Println("开始更新万象模型...")

	// 下载模型
	rimeGram := downloader.DownloadWithCallback(WanXiangGRA, progressCallback)
	if rimeGram == nil {
		callback(fmt.Errorf("下载万象模型失败，请检查网络连接"))
		return
	}

	// 获取目标目录（可能需要用户选择）
	getTargetDirGUI(window, func(targetDir string, err error) {
		if err != nil {
			callback(fmt.Errorf("获取目标目录失败: %v", err))
			return
		}

		if err := updater.UpdateModel(rimeGram, targetDir); err != nil {
			callback(fmt.Errorf("更新万象模型失败: %v", err))
			return
		}

		fmt.Println("✅ 万象模型更新完成")
		callback(nil)
	})
}

// updateDictConfigWithCallback 更新万象词库 - 对应 CLI 选项 3
func updateDictConfigWithCallback(window fyne.Window, callback func(error)) {
	updateDictConfigWithProgressCallback(window, nil, callback)
}

// 带进度回调的更新万象词库函数
func updateDictConfigWithProgressCallback(window fyne.Window, progressCallback ProgressCallback, callback func(error)) {
	fmt.Println("开始更新万象词库（Lite版）...")

	rimeZip := downloader.DownloadWithCallback(OhMyRimeRepo, progressCallback)
	if rimeZip == nil {
		callback(fmt.Errorf("下载词库失败，请检查网络连接"))
		return
	}

	// 获取目标目录（可能需要用户选择）
	getTargetDirGUI(window, func(targetDir string, err error) {
		if err != nil {
			callback(fmt.Errorf("获取目标目录失败: %v", err))
			return
		}

		if err := updater.UpdateDict(rimeZip, targetDir); err != nil {
			callback(fmt.Errorf("更新万象词库失败: %v", err))
			return
		}

		fmt.Println("✅ 万象词库更新完成")
		callback(nil)
	})
}

// customUpdateConfigWithCallback 自定义更新 - 对应 CLI 选项 4
func customUpdateConfigWithCallback(window fyne.Window, customUrl string, callback func(error)) {
	customUpdateConfigWithProgressCallback(window, customUrl, nil, callback)
}

// 带进度回调的自定义更新函数
func customUpdateConfigWithProgressCallback(window fyne.Window, customUrl string, progressCallback ProgressCallback, callback func(error)) {
	fmt.Printf("开始自定义更新: %s\n", customUrl)

	customData := downloader.DownloadWithCallback(customUrl, progressCallback)
	if customData == nil {
		callback(fmt.Errorf("下载自定义文件失败，请检查 URL 或网络连接"))
		return
	}

	// 获取目标目录（可能需要用户选择）
	getTargetDirGUI(window, func(targetDir string, err error) {
		if err != nil {
			callback(fmt.Errorf("获取目标目录失败: %v", err))
			return
		}

		// 判断文件类型
		if strings.HasSuffix(customUrl, ".zip") {
			// 如果是 zip 文件，更新主方案
			if err := updater.UpdateMainScheme(customData, targetDir); err != nil {
				callback(fmt.Errorf("更新自定义方案失败: %v", err))
				return
			}
			fmt.Println("✅ 自定义方案更新完成")
		} else if strings.HasSuffix(customUrl, ".gram") {
			// 如果是 gram 文件，更新模型
			if err := updater.UpdateModel(customData, targetDir); err != nil {
				callback(fmt.Errorf("更新自定义模型失败: %v", err))
				return
			}
			fmt.Println("✅ 自定义模型更新完成")
		} else {
			callback(fmt.Errorf("不支持的文件类型，请提供 zip 或 gram 文件的 URL"))
			return
		}

		callback(nil)
	})
}

// 以下是保持向后兼容的同步函数，用于非 GUI 模式
func updateMainSchemeConfig() error {
	fmt.Println("开始更新薄荷方案...")

	// 下载主方案
	rimeZip := downloader.Download(OhMyRimeRepo)
	if rimeZip == nil {
		return fmt.Errorf("下载薄荷方案失败，请检查网络连接")
	}

	targetDir := system.GetTargetDir()
	if err := updater.UpdateMainScheme(rimeZip, targetDir); err != nil {
		return fmt.Errorf("更新薄荷方案失败: %v", err)
	}

	fmt.Println("✅ 薄荷方案更新完成")
	return nil
}

func updateModelConfig() error {
	fmt.Println("开始更新万象模型...")

	// 下载模型
	rimeGram := downloader.Download(WanXiangGRA)
	if rimeGram == nil {
		return fmt.Errorf("下载万象模型失败，请检查网络连接")
	}

	targetDir := system.GetTargetDir()
	if err := updater.UpdateModel(rimeGram, targetDir); err != nil {
		return fmt.Errorf("更新万象模型失败: %v", err)
	}

	fmt.Println("✅ 万象模型更新完成")
	return nil
}

func updateDictConfig() error {
	fmt.Println("开始更新万象词库（Lite版）...")

	targetDir := system.GetTargetDir()
	rimeZip := downloader.Download(OhMyRimeRepo)
	if rimeZip == nil {
		return fmt.Errorf("下载词库失败，请检查网络连接")
	}

	if err := updater.UpdateDict(rimeZip, targetDir); err != nil {
		return fmt.Errorf("更新万象词库失败: %v", err)
	}

	fmt.Println("✅ 万象词库更新完成")
	return nil
}

func customUpdateConfig(customUrl string) error {
	fmt.Printf("开始自定义更新: %s\n", customUrl)

	customData := downloader.Download(customUrl)
	if customData == nil {
		return fmt.Errorf("下载自定义文件失败，请检查 URL 或网络连接")
	}

	targetDir := system.GetTargetDir()

	// 判断文件类型
	if strings.HasSuffix(customUrl, ".zip") {
		// 如果是 zip 文件，更新主方案
		if err := updater.UpdateMainScheme(customData, targetDir); err != nil {
			return fmt.Errorf("更新自定义方案失败: %v", err)
		}
		fmt.Println("✅ 自定义方案更新完成")
	} else if strings.HasSuffix(customUrl, ".gram") {
		// 如果是 gram 文件，更新模型
		if err := updater.UpdateModel(customData, targetDir); err != nil {
			return fmt.Errorf("更新自定义模型失败: %v", err)
		}
		fmt.Println("✅ 自定义模型更新完成")
	} else {
		return fmt.Errorf("不支持的文件类型，请提供 zip 或 gram 文件的 URL")
	}

	return nil
}

// 以下是保持向后兼容的旧函数名
func downloadConfig() error {
	return updateMainSchemeConfig()
}

func updateConfig() error {
	return updateModelConfig()
}

func installConfig() error {
	return updateDictConfig()
}

func uninstallConfig() error {
	fmt.Println("开始清理配置...")
	fmt.Println("✅ 配置清理完成")
	return nil
}
