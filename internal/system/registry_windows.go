//go:build windows

package system

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// 从Windows注册表读取Rime用户目录
func getRimeUserDirFromRegistry() (string, error) {
	// 打开注册表键 HKEY_CURRENT_USER\Software\Rime\Weasel
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Rime\Weasel`, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("无法打开注册表键: %v", err)
	}
	defer key.Close()

	// 读取 RimeUserDir 值
	rimeUserDir, _, err := key.GetStringValue("RimeUserDir")
	if err != nil {
		return "", fmt.Errorf("无法读取RimeUserDir值: %v", err)
	}

	return rimeUserDir, nil
}
