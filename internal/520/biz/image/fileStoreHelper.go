package image

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func isHexString(s string) bool {
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F') {
			return false
		}
	}
	return true
}

func genPathDir(hash string) (string, error) {
	// 常见哈希长度校验
	switch len(hash) {
	case 40: // SHA-1
	case 64: // SHA-256
	case 32: // MD5 (虽然Git不用)
	default:
		return "", fmt.Errorf("unsupported hash length: %d", len(hash))
	}
	// 哈希字符有效性检查
	if !isHexString(hash) {
		return "", fmt.Errorf("invalid hash format")
	}
	return filepath.Join(hash[:2], hash[2:]), nil
}

func saveFile(srcFile multipart.File, hash, fullPath string) error {
	// 原子性写入流程：
	// 1. 先写入临时文件
	// 2. 重命名为目标文件
	tempPath := fullPath + ".tmp"
	defer os.Remove(tempPath) // 确保临时文件最终被清理
	// 打开上传文件
	// 创建目标文件（使用独占模式防止并发写入）
	dstFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("create target file failed: %w", err)
	}
	// 复制文件内容（带缓冲区）
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return fmt.Errorf("write file content failed: %w", err)
	}
	// 确保数据刷到磁盘
	if err := dstFile.Sync(); err != nil {
		dstFile.Close()
		return fmt.Errorf("sync file failed: %w", err)
	}
	dstFile.Close()
	// 原子性重命名（跨卷移动需要特殊处理）
	if err := os.Rename(tempPath, fullPath); err != nil {
		// 处理跨卷移动的情况
		if linkErr := os.Link(tempPath, fullPath); linkErr != nil {
			return fmt.Errorf("rename file failed: %w (link also failed: %v)", err, linkErr)
		}
		os.Remove(tempPath)
	}
	return nil
}
