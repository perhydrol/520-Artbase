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
	pathDir, err := genPathDir(hash)
	if err != nil {
		return fmt.Errorf("generate pathDir failed: %w", err)
	}
	// 创建目标目录（带权限控制）
	fullDirPath := filepath.Join(i.baseDir, pathDir)
	if err := os.MkdirAll(fullDirPath, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open uploaded file failed: %w", err)
	}
	defer srcFile.Close()
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
