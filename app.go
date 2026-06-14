package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"oh-my-rime-cli/internal/constants"
	"oh-my-rime-cli/internal/downloader"
	"oh-my-rime-cli/internal/system"
	"oh-my-rime-cli/internal/updater"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	startStdoutCapture(ctx)
}

func startStdoutCapture(ctx context.Context) {
	originalStdout := os.Stdout
	r, wPipe, err := os.Pipe()
	if err != nil {
		return
	}
	os.Stdout = wPipe

	// 重定向标准日志输出到管道
	log.SetOutput(wPipe)

	go func() {
		reader := bufio.NewReader(r)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			// 输出到原始终端
			originalStdout.Write(line)

			cleanLine := string(line)
			if strings.TrimSpace(cleanLine) != "" {
				runtime.EventsEmit(ctx, "log", cleanLine)
			}
		}
	}()
}

func (a *App) getProgressCallback() downloader.ProgressCallback {
	return func(downloaded, total int64, percentage float64, speed float64) {
		downloadedStr := downloader.FormatBytes(downloaded)
		speedStr := downloader.FormatBytes(int64(speed)) + "/s"
		var details string
		if total > 0 {
			totalStr := downloader.FormatBytes(total)
			details = fmt.Sprintf("%s / %s (%s)", downloadedStr, totalStr, speedStr)
		} else {
			details = fmt.Sprintf("%s (%s)", downloadedStr, speedStr)
		}
		runtime.EventsEmit(a.ctx, "progress", map[string]interface{}{
			"percentage": percentage,
			"details":    details,
		})
	}
}

// SelectDirectory opens the native OS directory picker
func (a *App) SelectDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Rime 配置目录",
	})
	if err != nil {
		runtime.LogErrorf(a.ctx, "选择目录失败: %v", err)
		return ""
	}
	return dir
}

// UpdateAction performs the update download and extract
func (a *App) UpdateAction(actionType string, targetDir string, customUrl string) map[string]interface{} {
	if targetDir == "" && system.DetectOS() == "Windows_NT" {
		targetDir = system.GetWindowsTargetDir()
	} else if targetDir == "" {
		return map[string]interface{}{"success": false, "error": "请选择目标目录"}
	}

	progressCallback := a.getProgressCallback()
	var err error

	switch actionType {
	case "main":
		zip := downloader.DownloadWithCallback(constants.OhMyRimeRepo, progressCallback)
		if zip != nil {
			err = updater.UpdateMainScheme(zip, targetDir)
		} else {
			err = fmt.Errorf("下载主方案失败")
		}
	case "model":
		gram := downloader.DownloadWithCallback(constants.WanXiangGRA, progressCallback)
		if gram != nil {
			err = updater.UpdateModel(gram, targetDir)
		} else {
			err = fmt.Errorf("下载万象模型失败")
		}
	case "dict":
		zip := downloader.DownloadWithCallback(constants.OhMyRimeRepo, progressCallback)
		if zip != nil {
			err = updater.UpdateDict(zip, targetDir)
		} else {
			err = fmt.Errorf("下载万象词库失败")
		}
	case "custom":
		data := downloader.DownloadWithCallback(customUrl, progressCallback)
		if data != nil {
			if strings.HasSuffix(strings.ToLower(customUrl), ".zip") {
				err = updater.UpdateMainScheme(data, targetDir)
			} else {
				err = updater.UpdateModel(data, targetDir)
			}
		} else {
			err = fmt.Errorf("下载自定义资源失败")
		}
	default:
		err = fmt.Errorf("未知的更新类型")
	}

	if err != nil {
		runtime.EventsEmit(a.ctx, "log", err.Error()+"\n")
		return map[string]interface{}{"success": false, "error": err.Error()}
	}

	return map[string]interface{}{"success": true}
}

func (a *App) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"os": system.DetectOS(),
	}
}

func (a *App) OpenUrlBrowser(url string) map[string]interface{} {
	if url != "" {
		system.OpenUrlBrowser(url)
	}
	return map[string]interface{}{"success": true}
}
