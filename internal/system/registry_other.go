//go:build !windows

package system

import "fmt"

func getRimeUserDirFromRegistry() (string, error) {
	return "", fmt.Errorf("Windows注册表功能仅在Windows系统上可用")
}
