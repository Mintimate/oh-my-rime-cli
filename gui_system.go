package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// GUI 版本的目标目录获取函数
func getTargetDirGUI(window fyne.Window, callback func(string, error)) {
	switch detectOS() {
	case "Windows_NT":
		// Windows 系统直接返回目录，无需用户选择
		targetDir := getWindowsTargetDir()
		callback(targetDir, nil)
	case "Linux":
		// Linux 系统需要用户选择
		showLinuxDirSelectionGUI(window, callback)
	case "Darwin":
		// macOS 系统需要用户选择
		showDarwinDirSelectionGUI(window, callback)
	default:
		callback("", fmt.Errorf("不支持的操作系统"))
	}
}

// Linux 系统目录选择 GUI
func showLinuxDirSelectionGUI(window fyne.Window, callback func(string, error)) {
	var selectedDir string

	// 默认选择第一个选项
	selectedDir = filepath.Join(os.Getenv("HOME"), ".config", "rime")

	// 创建预览标签
	previewLabel := widget.NewLabel(selectedDir)

	// 创建单选组
	radioGroup := widget.NewRadioGroup([]string{"iBus", "Fcitx5", "Fcitx5-Flatpak"}, func(value string) {
		switch value {
		case "iBus":
			selectedDir = filepath.Join(os.Getenv("HOME"), ".config", "rime")
		case "Fcitx5":
			selectedDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime")
		case "Fcitx5-Flatpak":
			selectedDir = filepath.Join(os.Getenv("HOME"), ".var", "app", "org.fcitx.Fcitx5", "data", "fcitx5", "rime")
		}
		previewLabel.SetText(selectedDir)
	})
	radioGroup.SetSelected("iBus") // 默认选择

	// 创建内容
	content := container.NewVBox(
		widget.NewLabelWithStyle("请选择 Rime 配置目录类型", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		radioGroup,
		widget.NewSeparator(),
		widget.NewLabel("目标地址预览:"),
		previewLabel,
	)

	// 创建对话框
	dialog.ShowCustomConfirm("选择配置目录", "确定", "取消", content,
		func(confirmed bool) {
			if confirmed {
				fmt.Printf("Linux 目标地址: %s\n", selectedDir)
				callback(selectedDir, nil)
			} else {
				callback("", fmt.Errorf("用户取消了目录选择"))
			}
		}, window)
}

// macOS 系统目录选择 GUI
func showDarwinDirSelectionGUI(window fyne.Window, callback func(string, error)) {
	var selectedDir string

	// 默认选择第一个选项
	selectedDir = filepath.Join(os.Getenv("HOME"), "Library", "Rime")

	// 创建预览标签
	previewLabel := widget.NewLabel(selectedDir)

	// 创建单选组
	radioGroup := widget.NewRadioGroup([]string{"鼠须管", "小企鹅"}, func(value string) {
		switch value {
		case "鼠须管":
			selectedDir = filepath.Join(os.Getenv("HOME"), "Library", "Rime")
		case "小企鹅":
			selectedDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "fcitx5", "rime")
		}
		previewLabel.SetText(selectedDir)
	})
	radioGroup.SetSelected("鼠须管") // 默认选择

	// 创建内容
	content := container.NewVBox(
		widget.NewLabelWithStyle("请选择 Rime 配置目录类型", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		radioGroup,
		widget.NewSeparator(),
		widget.NewLabel("目标地址预览:"),
		previewLabel,
	)

	// 创建对话框
	dialog.ShowCustomConfirm("选择配置目录", "确定", "取消", content,
		func(confirmed bool) {
			if confirmed {
				fmt.Printf("macOS 目标地址: %s\n", selectedDir)
				callback(selectedDir, nil)
			} else {
				callback("", fmt.Errorf("用户取消了目录选择"))
			}
		}, window)
}

// 检查当前系统是否需要用户选择目录
func needsDirectorySelection() bool {
	os := runtime.GOOS
	return os == "linux" || os == "darwin"
}
