package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"fyne.io/fyne/v2"
)

// setupLogging 设置日志重定向
func (g *GUI) setupLogging() {
	// 重定向标准输出和错误输出到GUI
	multiWriter := io.MultiWriter(g.originalOut, g.logWriter)
	log.SetOutput(multiWriter)

	g.appendLogSafe("=== Oh My Rime CLI 启动 ===")
	g.appendLogSafe(fmt.Sprintf("当前时间: %s", time.Now().Format("2006-01-02 15:04:05")))
}

// appendLogSafe 线程安全的日志追加方法
func (g *GUI) appendLogSafe(message string) {
	// 添加时间戳
	timestamp := time.Now().Format("15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	// 输出到控制台
	log.Println(formattedMessage)

	// 在主线程中更新GUI日志
	if g.logText != nil {
		fyne.Do(func() {
			currentText := g.logText.Text
			if currentText != "" {
				currentText += "\n"
			}
			g.logText.SetText(currentText + formattedMessage)

			// 滚动到底部
			if g.logScroll != nil {
				g.logScroll.ScrollToBottom()
			}
		})
	}
}

// setStatusSafe 线程安全的状态设置方法
func (g *GUI) setStatusSafe(status string) {
	log.Printf("状态: %s", status)

	// 在主线程中更新状态标签
	if g.statusLabel != nil {
		fyne.Do(func() {
			g.statusLabel.SetText(status)
		})
	}
}

// showProgressSafe 线程安全的进度显示方法
func (g *GUI) showProgressSafe() {
	log.Println("显示进度条")

	// 在主线程中显示进度条
	if g.progressBar != nil {
		fyne.Do(func() {
			g.progressBar.Show()
			g.progressBar.SetValue(0)
		})
	}
}

// hideProgressSafe 线程安全的进度隐藏方法
func (g *GUI) hideProgressSafe() {
	log.Println("隐藏进度条")

	// 在主线程中隐藏进度条
	if g.progressBar != nil {
		fyne.Do(func() {
			g.progressBar.Hide()
		})
	}
}

// updateProgressSafe 线程安全的进度更新方法
func (g *GUI) updateProgressSafe(value float64) {
	log.Printf("进度更新: %.1f%%", value*100)

	// 在主线程中更新进度条
	if g.progressBar != nil {
		fyne.Do(func() {
			g.progressBar.SetValue(value)
		})
	}
}

// disableButtonsSafe 线程安全的按钮禁用方法
func (g *GUI) disableButtonsSafe() {
	log.Println("禁用按钮")

	// 在主线程中禁用所有按钮
	fyne.Do(func() {
		if g.downloadBtn != nil {
			g.downloadBtn.Disable()
		}
		if g.updateBtn != nil {
			g.updateBtn.Disable()
		}
		if g.installBtn != nil {
			g.installBtn.Disable()
		}
		if g.uninstallBtn != nil {
			g.uninstallBtn.Disable()
		}
	})
}

// enableButtonsSafe 线程安全的按钮启用方法
func (g *GUI) enableButtonsSafe() {
	log.Println("启用按钮")

	// 在主线程中启用所有按钮
	fyne.Do(func() {
		if g.downloadBtn != nil {
			g.downloadBtn.Enable()
		}
		if g.updateBtn != nil {
			g.updateBtn.Enable()
		}
		if g.installBtn != nil {
			g.installBtn.Enable()
		}
		if g.uninstallBtn != nil {
			g.uninstallBtn.Enable()
		}
	})
}

// logProgress 记录进度信息
func (g *GUI) logProgress(message string, progress float64) {
	g.appendLogSafe(message)
	g.updateProgressSafe(progress)
}

// resetUI 重置UI状态
func (g *GUI) resetUI() {
	g.enableButtonsSafe()
	g.hideProgressSafe()
	g.setStatusSafe("就绪")
}

// startOperation 开始操作的通用方法
func (g *GUI) startOperation(operationName string) {
	g.disableButtonsSafe()
	g.showProgressSafe()
	g.setStatusSafe(fmt.Sprintf("正在%s...", operationName))
	g.appendLogSafe(fmt.Sprintf("开始%s", operationName))
}

// finishOperation 完成操作的通用方法
func (g *GUI) finishOperation(operationName string, err error) {
	defer g.resetUI()

	if err != nil {
		g.setStatusSafe("操作失败")
		g.appendLogSafe(fmt.Sprintf("❌ %s失败: %v", operationName, err))
	} else {
		g.setStatusSafe("操作完成")
		g.appendLogSafe(fmt.Sprintf("✅ %s完成", operationName))
	}
}
