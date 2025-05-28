package image

import (
	"crypto/sha256"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/viper"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
)

var (
	imageFileStoreOnce sync.Once
	this               *imageFileStore
)

func isHexString(s string) bool {
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F') {
			return false
		}
	}
	return true
}

func genPath(hash string) (string, error) {
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

type ImageFileStore interface {
	Save(fileHeader *multipart.FileHeader, hash string) error
	Validate(fileHeader *multipart.FileHeader) (bool, error)
	IsContainerImage(hash string) (bool, error)
	Remove(fileHeader *multipart.FileHeader) error
	Hash(fileHeader *multipart.FileHeader) (string, error)
}

type imageFileStore struct {
	baseDir string
}

var _ ImageFileStore = (*imageFileStore)(nil)

func NewImageFileStore() ImageFileStore {
	imageFileStoreOnce.Do(func() {
		this = &imageFileStore{baseDir: viper.GetString("image_dir")}
	})
	return this
}

func (i *imageFileStore) Save(fileHeader *multipart.FileHeader, hash string) error {
	if fileHeader == nil || hash == "" {
		return fmt.Errorf("fileHeader or hash is nil")
	}
	ok, err := i.IsContainerImage(hash)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	// 生成存储路径
	path, err := genPath(hash)
	if err != nil {
		return fmt.Errorf("generate path failed: %w", err)
	}
	// 创建目标目录（带权限控制）
	fullPath := filepath.Join(i.baseDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}
	// 原子性写入流程：
	// 1. 先写入临时文件
	// 2. 重命名为目标文件
	tempPath := fullPath + ".tmp"
	defer os.Remove(tempPath) // 确保临时文件最终被清理
	// 打开上传文件
	srcFile, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open uploaded file failed: %w", err)
	}
	defer srcFile.Close()
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

func (i *imageFileStore) Validate(fileHeader *multipart.FileHeader) (bool, error) {
	if fileHeader == nil {
		return false, fmt.Errorf("fileHeader is nil")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return false, err
	}
	defer file.Close()

	mtype, err := mimetype.DetectReader(file)
	if err != nil {
		return false, err
	}
	isImage := mtype.Is("image/jpeg") ||
		mtype.Is("image/png") ||
		mtype.Is("image/gif") ||
		mtype.Is("image/webp")
	return isImage, nil
}

func (i *imageFileStore) IsContainerImage(hash string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (i *imageFileStore) Remove(fileHeader *multipart.FileHeader) error {
	//TODO implement me
	panic("implement me")
}

func (i *imageFileStore) Hash(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
