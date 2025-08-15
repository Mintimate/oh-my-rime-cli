package main

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	app    fyne.App
	window fyne.Window

	// UI 组件
	statusLabel  *widget.Label
	progressBar  *widget.ProgressBar
	logText      *widget.Entry
	downloadBtn  *widget.Button
	updateBtn    *widget.Button
	installBtn   *widget.Button
	uninstallBtn *widget.Button
}

func NewGUI() *GUI {
	myApp := app.NewWithID("com.rime.cli")
	myApp.SetIcon(theme.ComputerIcon())

	window := myApp.NewWindow("Oh My Rime - 输入法配置管理工具")
	window.Resize(fyne.NewSize(800, 600))
	window.CenterOnScreen()

	gui := &GUI{
		app:    myApp,
		window: window,
	}

	gui.setupUI()
	return gui
}

func (g *GUI) setupUI() {
	// 标题
	title := widget.NewLabelWithStyle("Oh My Rime", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.TextStyle.Monospace = false

	subtitle := widget.NewLabelWithStyle("Rime 输入法配置管理工具", fyne.TextAlignCenter, fyne.TextStyle{})

	// 状态显示
	g.statusLabel = widget.NewLabel("就绪")
	g.progressBar = widget.NewProgressBar()
	g.progressBar.Hide()

	// 按钮区域 - 对齐 CLI 功能
	g.downloadBtn = widget.NewButton("📦 更新薄荷方案", g.onUpdateMainScheme)
	g.downloadBtn.Importance = widget.HighImportance

	g.updateBtn = widget.NewButton("🧠 更新万象模型", g.onUpdateModel)
	g.installBtn = widget.NewButton("📚 更新万象词库", g.onUpdateDict)
	g.uninstallBtn = widget.NewButton("🔗 自定义更新", g.onCustomUpdate)

	buttonContainer := container.NewGridWithColumns(2,
		g.downloadBtn,
		g.updateBtn,
		g.installBtn,
		g.uninstallBtn,
	)

	// 日志区域
	logLabel := widget.NewLabelWithStyle("操作日志:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	g.logText = widget.NewMultiLineEntry()
	g.logText.SetPlaceHolder("操作日志将在这里显示...")
	g.logText.Wrapping = fyne.TextWrapWord
	g.logText.Disable()

	logContainer := container.NewBorder(logLabel, nil, nil, nil,
		container.NewScroll(g.logText))
	logContainer.Resize(fyne.NewSize(0, 200))

	// 状态栏
	statusContainer := container.NewBorder(nil, nil,
		widget.NewLabel("状态:"), nil, g.statusLabel)

	// 主布局
	content := container.NewVBox(
		container.NewVBox(title, subtitle),
		widget.NewSeparator(),
		statusContainer,
		g.progressBar,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
		logContainer,
	)

	g.window.SetContent(container.NewPadded(content))
}

func (g *GUI) appendLog(message string) {
	current := g.logText.Text
	if current != "" {
		current += "\n"
	}
	current += message
	g.logText.SetText(current)
	g.logText.CursorRow = len(strings.Split(current, "\n"))
}

func (g *GUI) setStatus(status string) {
	g.statusLabel.SetText(status)
}

func (g *GUI) showProgress() {
	g.progressBar.Show()
}

func (g *GUI) hideProgress() {
	g.progressBar.Hide()
}

func (g *GUI) disableButtons() {
	g.downloadBtn.Disable()
	g.updateBtn.Disable()
	g.installBtn.Disable()
	g.uninstallBtn.Disable()
}

func (g *GUI) enableButtons() {
	g.downloadBtn.Enable()
	g.updateBtn.Enable()
	g.installBtn.Enable()
	g.uninstallBtn.Enable()
}

func (g *GUI) onUpdateMainScheme() {
	g.disableButtons()
	g.showProgress()
	g.setStatus("正在更新薄荷方案...")
	g.appendLog("开始更新薄荷方案...")

	updateMainSchemeConfigWithCallback(g.window, func(err error) {
		// 确保 UI 更新在主线程中执行
		defer func() {
			g.enableButtons()
			g.hideProgress()
		}()

		if err != nil {
			g.setStatus("更新失败")
			g.appendLog(fmt.Sprintf("❌ 薄荷方案更新失败: %v", err))
			dialog.ShowError(err, g.window)
		} else {
			g.setStatus("更新完成")
			g.appendLog("✅ 薄荷方案更新完成")
			dialog.ShowInformation("成功", "薄荷方案更新完成！", g.window)
		}
	})
}

func (g *GUI) onUpdateModel() {
	g.disableButtons()
	g.showProgress()
	g.setStatus("正在更新万象模型...")
	g.appendLog("开始更新万象模型...")

	updateModelConfigWithCallback(g.window, func(err error) {
		// 确保 UI 更新在主线程中执行
		defer func() {
			g.enableButtons()
			g.hideProgress()
		}()

		if err != nil {
			g.setStatus("更新失败")
			g.appendLog(fmt.Sprintf("❌ 万象模型更新失败: %v", err))
			dialog.ShowError(err, g.window)
		} else {
			g.setStatus("更新完成")
			g.appendLog("✅ 万象模型更新完成")
			dialog.ShowInformation("成功", "万象模型更新完成！", g.window)
		}
	})
}

func (g *GUI) onUpdateDict() {
	g.disableButtons()
	g.showProgress()
	g.setStatus("正在更新万象词库...")
	g.appendLog("开始更新万象词库（Lite版）...")

	updateDictConfigWithCallback(g.window, func(err error) {
		// 确保 UI 更新在主线程中执行
		defer func() {
			g.enableButtons()
			g.hideProgress()
		}()

		if err != nil {
			g.setStatus("更新失败")
			g.appendLog(fmt.Sprintf("❌ 万象词库更新失败: %v", err))
			dialog.ShowError(err, g.window)
		} else {
			g.setStatus("更新完成")
			g.appendLog("✅ 万象词库更新完成")
			dialog.ShowInformation("成功", "万象词库（Lite版）更新完成！", g.window)
		}
	})
}

func (g *GUI) onCustomUpdate() {
	// 创建自定义 URL 输入对话框
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("请输入 ZIP 或 GRAM 文件的 URL...")
	urlEntry.Resize(fyne.NewSize(400, 0))

	content := container.NewVBox(
		widget.NewLabel("自定义更新功能:"),
		widget.NewLabel("• ZIP 文件 => 更新方案包"),
		widget.NewLabel("• GRAM 文件 => 更新模型文件"),
		widget.NewSeparator(),
		widget.NewLabel("请输入文件 URL:"),
		urlEntry,
	)

	dialog.ShowCustomConfirm("自定义更新", "确定", "取消", content,
		func(confirmed bool) {
			if !confirmed {
				return
			}

			customUrl := strings.TrimSpace(urlEntry.Text)
			if customUrl == "" {
				dialog.ShowError(fmt.Errorf("URL 不能为空"), g.window)
				return
			}

			g.disableButtons()
			g.showProgress()
			g.setStatus("正在自定义更新...")
			g.appendLog(fmt.Sprintf("开始自定义更新: %s", customUrl))

			customUpdateConfigWithCallback(g.window, customUrl, func(err error) {
				// 确保 UI 更新在主线程中执行
				defer func() {
					g.enableButtons()
					g.hideProgress()
				}()

				if err != nil {
					g.setStatus("更新失败")
					g.appendLog(fmt.Sprintf("❌ 自定义更新失败: %v", err))
					dialog.ShowError(err, g.window)
				} else {
					g.setStatus("更新完成")
					g.appendLog("✅ 自定义更新完成")
					dialog.ShowInformation("成功", "自定义更新完成！", g.window)
				}
			})
		}, g.window)
}

func (g *GUI) Run() {
	g.window.ShowAndRun()
}

// 包装原有的功能函数
func downloadRimeConfig() error {
	log.Println("下载 Rime 配置...")
	return downloadConfig()
}

func updateRimeConfig() error {
	log.Println("更新 Rime 配置...")
	return updateConfig()
}

func installRimeConfig() error {
	log.Println("安装 Rime 配置...")
	return installConfig()
}

func uninstallRimeConfig() error {
	log.Println("卸载 Rime 配置...")
	return uninstallConfig()
}
