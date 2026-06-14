package updater

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const backupKeepCount = 3

// UpdateMainScheme 更新主方案
func UpdateMainScheme(rimeZip []byte, targetDir string) error {
	fmt.Println("正在更新主方案...")

	// 检查zip数据是否有效
	if rimeZip == nil || len(rimeZip) == 0 {
		return fmt.Errorf("zip数据无效")
	}

	return runWithBackup("主方案更新", targetDir, func() error {
		// 创建目标目录（如果不存在）
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("创建目标目录失败: %v", err)
		}

		// 从字节数组创建zip reader
		zipReader, err := zip.NewReader(bytes.NewReader(rimeZip), int64(len(rimeZip)))
		if err != nil {
			return fmt.Errorf("读取zip文件失败: %v", err)
		}

		// 遍历zip文件中的每个文件
		for _, file := range zipReader.File {
			// 构建目标文件路径
			targetPath, err := safeJoin(targetDir, file.Name)
			if err != nil {
				return err
			}

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
					return err
				}

				// 解压文件
				if err := extractFile(file, targetPath); err != nil {
					fmt.Printf("解压文件失败 %s: %v\n", targetPath, err)
					return err
				}
				fmt.Printf("解压文件: %s\n", targetPath)
			}
		}

		fmt.Println("✅ 主方案更新完成！")
		return nil
	})
}

// UpdateModel 更新模型文件
func UpdateModel(rimeGram []byte, targetDir string) error {
	fmt.Println("正在更新模型...")

	// 检查模型数据是否有效
	if rimeGram == nil || len(rimeGram) == 0 {
		return fmt.Errorf("模型数据无效")
	}

	return runWithBackup("模型更新", targetDir, func() error {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("创建目标目录失败: %v", err)
		}

		// 覆盖保存目录内的模型文件
		modelPath := filepath.Join(targetDir, "wanxiang-lts-zh-hans.gram")
		if err := os.WriteFile(modelPath, rimeGram, 0644); err != nil {
			return fmt.Errorf("更新模型失败: %v", err)
		}

		fmt.Println("✅ 模型更新完成！")
		return nil
	})
}

// UpdateDict 更新词库
func UpdateDict(rimeZip []byte, targetDir string) error {
	fmt.Println("正在更新词库...")

	// 检查zip数据是否有效
	if rimeZip == nil || len(rimeZip) == 0 {
		return fmt.Errorf("zip数据无效")
	}

	return runWithBackup("词库更新", targetDir, func() error {
		// 创建目标词库目录
		dictsTargetDir := filepath.Join(targetDir, "dicts")
		if err := os.MkdirAll(dictsTargetDir, 0755); err != nil {
			return fmt.Errorf("创建词库目录失败: %v", err)
		}

		// 从字节数组创建zip reader
		zipReader, err := zip.NewReader(bytes.NewReader(rimeZip), int64(len(rimeZip)))
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
			targetPath, err := safeJoin(dictsTargetDir, relativePath)
			if err != nil {
				return err
			}

			if file.FileInfo().IsDir() {
				// 创建子目录
				if err := os.MkdirAll(targetPath, file.FileInfo().Mode()); err != nil {
					fmt.Printf("创建词库子目录失败 %s: %v\n", targetPath, err)
					return err
				}
				fmt.Printf("创建词库目录: %s\n", targetPath)
			} else {
				// 创建父目录
				if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
					fmt.Printf("创建父目录失败 %s: %v\n", filepath.Dir(targetPath), err)
					return err
				}

				// 解压文件
				if err := extractFile(file, targetPath); err != nil {
					fmt.Printf("解压词库文件失败 %s: %v\n", targetPath, err)
					return err
				}
				fmt.Printf("更新词库文件: %s\n", targetPath)
			}
		}

		fmt.Println("✅ 词库更新完成！")
		return nil
	})
}

func runWithBackup(operationName, targetDir string, update func() error) error {
	backupDir, hasBackup, err := createBackup(targetDir)
	if err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}
	if hasBackup {
		fmt.Printf("已创建备份: %s\n", backupDir)
	}

	if err := update(); err != nil {
		if hasBackup {
			fmt.Printf("%s失败，正在恢复备份...\n", operationName)
			if restoreErr := restoreBackup(targetDir, backupDir); restoreErr != nil {
				return fmt.Errorf("%v；备份恢复失败: %v", err, restoreErr)
			}
			fmt.Println("已恢复到更新前状态")
		}
		return err
	}

	if err := pruneBackups(targetDir, backupKeepCount); err != nil {
		fmt.Printf("清理旧备份失败: %v\n", err)
	}
	return nil
}

func createBackup(targetDir string) (string, bool, error) {
	info, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("目标目录不存在，将跳过备份并创建新目录")
			return "", false, nil
		}
		return "", false, err
	}
	if !info.IsDir() {
		return "", false, fmt.Errorf("目标路径不是目录: %s", targetDir)
	}

	backupRoot := backupRootDir(targetDir)
	if err := os.MkdirAll(backupRoot, 0755); err != nil {
		return "", false, err
	}

	backupDir := filepath.Join(backupRoot, time.Now().Format("20060102-150405"))
	if err := copyDir(targetDir, backupDir); err != nil {
		return "", false, err
	}
	return backupDir, true, nil
}

func restoreBackup(targetDir, backupDir string) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	return copyDir(backupDir, targetDir)
}

func backupRootDir(targetDir string) string {
	cleanTarget := filepath.Clean(targetDir)
	return filepath.Join(filepath.Dir(cleanTarget), filepath.Base(cleanTarget)+".backups")
}

func pruneBackups(targetDir string, keep int) error {
	backupRoot := backupRootDir(targetDir)
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var dirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		}
	}
	if len(dirs) <= keep {
		return nil
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name() < dirs[j].Name()
	})

	for _, entry := range dirs[:len(dirs)-keep] {
		if err := os.RemoveAll(filepath.Join(backupRoot, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		return copyFile(path, targetPath, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func safeJoin(baseDir, name string) (string, error) {
	cleanName := filepath.Clean(name)
	if filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("zip 包含不安全路径: %s", name)
	}

	targetPath := filepath.Join(baseDir, cleanName)
	relPath, err := filepath.Rel(baseDir, targetPath)
	if err != nil {
		return "", err
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("zip 包含不安全路径: %s", name)
	}
	return targetPath, nil
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
