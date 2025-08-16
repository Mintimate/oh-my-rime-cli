package gui

import (
	"fmt"
	"io"
	"log"
	"oh-my-rime-cli/internal/downloader"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// LogWriter 实现 io.Writer 接口，用于重定向控制台输出到GUI
type LogWriter struct {
	gui *GUI
	mu  sync.Mutex
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	message := strings.TrimSpace(string(p))
	if message != "" {
		// 在主线程中更新UI
		go func() {
			lw.gui.appendLogSafe(message)
		}()
	}
	return len(p), nil
}

type GUI struct {
	app    fyne.App
	window fyne.Window

	// UI 组件
	statusLabel  *widget.Label
	progressBar  *widget.ProgressBar
	logText      *widget.Entry
	logScroll    *container.Scroll
	downloadBtn  *widget.Button
	updateBtn    *widget.Button
	installBtn   *widget.Button
	uninstallBtn *widget.Button

	// 日志管理
	logMutex    sync.Mutex
	logWriter   *LogWriter
	originalOut io.Writer
	originalErr io.Writer

	// 进度管理
	progressMutex sync.Mutex
	isRunning     bool
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

	// 添加新的按钮 - 打开链接
	bilibiliBtn := widget.NewButton("📺 关注作者 Bilibili", func() {
		OpenAuthorBilibili()
		g.appendLogSafe("已打开作者 Bilibili 页面")
	})
	docBtn := widget.NewButton("📖 打开薄荷文档", func() {
		OpenMintimateDoc()
		g.appendLogSafe("已打开薄荷输入法文档")
	})

	// 主功能按钮容器
	mainButtonContainer := container.NewGridWithColumns(2,
		g.downloadBtn,
		g.updateBtn,
		g.installBtn,
		g.uninstallBtn,
	)

	// 链接按钮容器
	linkButtonContainer := container.NewGridWithColumns(2,
		bilibiliBtn,
		docBtn,
	)

	// 组合所有按钮
	buttonContainer := container.NewVBox(
		mainButtonContainer,
		widget.NewSeparator(),
		linkButtonContainer,
	)

	// 日志区域
	logLabel := widget.NewLabelWithStyle("操作日志:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	g.logText = widget.NewMultiLineEntry()
	g.logText.SetPlaceHolder("操作日志将在这里显示...")
	g.logText.Wrapping = fyne.TextWrapWord
	g.logText.Disable()
	// 设置日志文本框的最小行数，增加显示高度
	g.logText.Resize(fyne.NewSize(0, 300))

	g.logScroll = container.NewScroll(g.logText)
	logContainer := container.NewBorder(logLabel, nil, nil, nil, g.logScroll)
	// 大幅增加日志容器的高度到 450 像素，提供更多日志显示空间
	logContainer.Resize(fyne.NewSize(0, 450))

	// 状态栏
	statusContainer := container.NewBorder(nil, nil,
		widget.NewLabel("状态:"), nil, g.statusLabel)

	// 顶部区域（标题、状态、按钮）
	topContent := container.NewVBox(
		container.NewVBox(title, subtitle),
		widget.NewSeparator(),
		statusContainer,
		g.progressBar,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
	)

	// 主布局 - 使用 Border 布局让日志区域占据剩余空间
	content := container.NewBorder(
		topContent,   // 顶部
		nil,          // 底部
		nil,          // 左侧
		nil,          // 右侧
		logContainer, // 中心区域（日志）
	)

	g.window.SetContent(container.NewPadded(content))
}

func (g *GUI) onUpdateMainScheme() {
	// 在后台线程执行，避免阻塞UI
	go func() {
		g.startOperation("更新薄荷方案")

		// 创建进度回调函数
		progressCallback := func(downloaded, total int64, percentage float64, speed float64) {
			g.updateProgressSafe(percentage / 100.0)
			speedStr := downloader.FormatBytes(int64(speed))
			g.appendLogSafe(fmt.Sprintf("下载进度: %.1f%% (%s/%s) 速度: %s/s",
				percentage, downloader.FormatBytes(downloaded), downloader.FormatBytes(total), speedStr))
		}

		updateMainSchemeConfigWithProgressCallback(g.window, progressCallback, func(err error) {
			g.finishOperation("薄荷方案更新", err)

			if err != nil {
				dialog.ShowError(err, g.window)
			} else {
				dialog.ShowInformation("成功", "薄荷方案更新完成！", g.window)
			}
		})
	}()
}

func (g *GUI) onUpdateModel() {
	// 在后台线程执行，避免阻塞UI
	go func() {
		g.startOperation("更新万象模型")

		// 创建进度回调函数
		progressCallback := func(downloaded, total int64, percentage float64, speed float64) {
			g.updateProgressSafe(percentage / 100.0)
			speedStr := downloader.FormatBytes(int64(speed))
			g.appendLogSafe(fmt.Sprintf("下载进度: %.1f%% (%s/%s) 速度: %s/s",
				percentage, downloader.FormatBytes(downloaded), downloader.FormatBytes(total), speedStr))
		}

		updateModelConfigWithProgressCallback(g.window, progressCallback, func(err error) {
			g.finishOperation("万象模型更新", err)

			if err != nil {
				dialog.ShowError(err, g.window)
			} else {
				dialog.ShowInformation("成功", "万象模型更新完成！", g.window)
			}
		})
	}()
}

func (g *GUI) onUpdateDict() {
	// 在后台线程执行，避免阻塞UI
	go func() {
		g.startOperation("更新万象词库")

		// 创建进度回调函数
		progressCallback := func(downloaded, total int64, percentage float64, speed float64) {
			g.updateProgressSafe(percentage / 100.0)
			speedStr := downloader.FormatBytes(int64(speed))
			g.appendLogSafe(fmt.Sprintf("下载进度: %.1f%% (%s/%s) 速度: %s/s",
				percentage, downloader.FormatBytes(downloaded), downloader.FormatBytes(total), speedStr))
		}

		updateDictConfigWithProgressCallback(g.window, progressCallback, func(err error) {
			g.finishOperation("万象词库更新", err)

			if err != nil {
				dialog.ShowError(err, g.window)
			} else {
				dialog.ShowInformation("成功", "万象词库（Lite版）更新完成！", g.window)
			}
		})
	}()
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

			// 在后台线程执行，避免阻塞UI
			go func() {
				g.startOperation("自定义更新")
				g.appendLogSafe(fmt.Sprintf("URL: %s", customUrl))

				// 创建进度回调函数
				progressCallback := func(downloaded, total int64, percentage float64, speed float64) {
					g.updateProgressSafe(percentage / 100.0)
					speedStr := downloader.FormatBytes(int64(speed))
					g.appendLogSafe(fmt.Sprintf("下载进度: %.1f%% (%s/%s) 速度: %s/s",
						percentage, downloader.FormatBytes(downloaded), downloader.FormatBytes(total), speedStr))
				}

				customUpdateConfigWithProgressCallback(g.window, customUrl, progressCallback, func(err error) {
					g.finishOperation("自定义更新", err)

					if err != nil {
						dialog.ShowError(err, g.window)
					} else {
						dialog.ShowInformation("成功", "自定义更新完成！", g.window)
					}
				})
			}()
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
