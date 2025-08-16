package updater

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UpdateMainScheme 更新主方案
func UpdateMainScheme(rimeZip []byte, targetDir string) error {
	fmt.Println("正在更新主方案...")

	// 检查zip数据是否有效
	if rimeZip == nil || len(rimeZip) == 0 {
		return fmt.Errorf("zip数据无效")
	}

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
		targetPath := filepath.Join(targetDir, file.Name)

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
			}

			// 解压文件
			if err := extractFile(file, targetPath); err != nil {
				fmt.Printf("解压文件失败 %s: %v\n", targetPath, err)
				continue
			}
			fmt.Printf("解压文件: %s\n", targetPath)
		}
	}

	fmt.Println("✅ 主方案更新完成！")
	return nil
}

// UpdateModel 更新模型文件
func UpdateModel(rimeGram []byte, targetDir string) error {
	fmt.Println("正在更新模型...")

	// 检查模型数据是否有效
	if rimeGram == nil || len(rimeGram) == 0 {
		return fmt.Errorf("模型数据无效")
	}

	// 覆盖保存目录内的模型文件
	modelPath := filepath.Join(targetDir, "wanxiang-lts-zh-hans.gram")
	if err := os.WriteFile(modelPath, rimeGram, 0644); err != nil {
		return fmt.Errorf("更新模型失败: %v", err)
	}

	fmt.Println("✅ 模型更新完成！")
	return nil
}

// UpdateDict 更新词库
func UpdateDict(rimeZip []byte, targetDir string) error {
	fmt.Println("正在更新词库...")

	// 检查zip数据是否有效
	if rimeZip == nil || len(rimeZip) == 0 {
		return fmt.Errorf("zip数据无效")
	}

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
		targetPath := filepath.Join(dictsTargetDir, relativePath)

		if file.FileInfo().IsDir() {
			// 创建子目录
			if err := os.MkdirAll(targetPath, file.FileInfo().Mode()); err != nil {
				fmt.Printf("创建词库子目录失败 %s: %v\n", targetPath, err)
				continue
			}
			fmt.Printf("创建词库目录: %s\n", targetPath)
		} else {
			// 创建父目录
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				fmt.Printf("创建父目录失败 %s: %v\n", filepath.Dir(targetPath), err)
				continue
			}

			// 解压文件
			if err := extractFile(file, targetPath); err != nil {
				fmt.Printf("解压词库文件失败 %s: %v\n", targetPath, err)
				continue
			}
			fmt.Printf("更新词库文件: %s\n", targetPath)
		}
	}

	fmt.Println("✅ 词库更新完成！")
	return nil
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
