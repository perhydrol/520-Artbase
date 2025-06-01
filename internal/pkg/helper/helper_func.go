package helper

import (
	"demo520/internal/pkg/log"
	"fmt"
	"io"
	"os"
)

func WriteFile(filePath string, srcFile io.Reader) error {
	dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return fmt.Errorf("create destination file failed: %w", err)
	}
	defer dstFile.Close()
	// 使用io.Copy缓冲写入
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		log.Errorw("Failed to write file",
			"err", err,
			"filePath", filePath,
		)
		// 尝试清理失败的文件
		os.Remove(filePath)
		return fmt.Errorf("write file failed: %w", err)
	}
	return err
}
