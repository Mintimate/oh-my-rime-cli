package downloader

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ProgressCallback 进度回调函数类型
type ProgressCallback func(downloaded, total int64, percentage float64, speed float64)

// 进度条读取器
type ProgressReader struct {
	io.Reader
	Total      int64
	Downloaded int64
	StartTime  time.Time
	LastUpdate time.Time
	Callback   ProgressCallback
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Downloaded += int64(n)

	// 每100ms更新一次进度条
	now := time.Now()
	if now.Sub(pr.LastUpdate) > 100*time.Millisecond || err == io.EOF {
		pr.LastUpdate = now
		pr.updateProgress()
	}

	return n, err
}

func (pr *ProgressReader) updateProgress() {
	if pr.Total <= 0 {
		fmt.Printf("\r下载中... %s", FormatBytes(pr.Downloaded))
		if pr.Callback != nil {
			pr.Callback(pr.Downloaded, pr.Total, 0, 0)
		}
		return
	}

	percentage := float64(pr.Downloaded) / float64(pr.Total) * 100
	elapsed := time.Since(pr.StartTime)

	// 计算下载速度
	speed := float64(pr.Downloaded) / elapsed.Seconds()

	// 调用回调函数
	if pr.Callback != nil {
		pr.Callback(pr.Downloaded, pr.Total, percentage, speed)
	}

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
		FormatBytes(pr.Downloaded),
		FormatBytes(pr.Total),
		FormatBytes(int64(speed)),
		eta)
}

// Download 下载文件并返回字节数据
func Download(url string) []byte {
	return DownloadWithCallback(url, nil)
}

// DownloadWithCallback 带进度回调的下载函数
func DownloadWithCallback(url string, callback ProgressCallback) []byte {
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
		Callback:   callback,
	}

	// 读取响应内容到内存
	data, err := io.ReadAll(progressReader)
	if err != nil {
		fmt.Printf("\n读取内容失败: %v\n", err)
		return nil
	}

	fmt.Printf("\n✅ 下载完成! 总大小: %s\n", FormatBytes(int64(len(data))))
	return data
}

// FormatBytes 格式化字节大小
func FormatBytes(bytes int64) string {
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
