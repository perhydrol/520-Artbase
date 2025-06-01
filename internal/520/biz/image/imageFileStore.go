package image

import (
	"crypto/sha256"
	"demo520/internal/pkg/convert"
	"demo520/internal/pkg/helper"
	"demo520/internal/pkg/log"
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
	baseDir        string
	imageConverter convert.ImageConverter
}

var _ ImageFileStore = (*imageFileStore)(nil)

func NewImageFileStore() ImageFileStore {
	imageFileStoreOnce.Do(func() {
		this = &imageFileStore{baseDir: viper.GetString("image_dir"), imageConverter: convert.InitImageConverter()}
	})
	return this
}

func (i *imageFileStore) Save(fileHeader *multipart.FileHeader, hash string) error {
	// 参数校验
	if fileHeader == nil {
		return fmt.Errorf("fileHeader cannot be nil")
	}
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}
	// 检查是否已存在
	exists, err := i.IsContainerImage(hash)
	if err != nil {
		return fmt.Errorf("check image existence failed: %w", err)
	}
	if exists {
		return nil
	}
	// 生成存储路径
	pathDir, err := genPathDir(hash)
	if err != nil {
		return fmt.Errorf("generate pathDir failed: %w", err)
	}
	// 创建目录（更严格的权限）
	fullDirPath := filepath.Join(i.baseDir, pathDir)
	if err := os.MkdirAll(fullDirPath, 0750); err != nil { // 750更安全
		return fmt.Errorf("create directory %s failed: %w", fullDirPath, err)
	}

	// 打开源文件
	srcFile, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open uploaded file failed: %w", err)
	}
	defer srcFile.Close()
	// 创建目标文件（使用更安全的文件名）
	fileName := fmt.Sprintf("%s%s", hash, filepath.Ext(fileHeader.Filename))
	filePath := filepath.Join(fullDirPath, fileName)
	saveSrcFileErr := helper.WriteFile(filePath, srcFile)
	if saveSrcFileErr != nil {
		return saveSrcFileErr
	}
	convertErr := i.imageConverter.ConvertImage(filePath)
	if convertErr != nil {
		log.Errorw("Convert image file failed", "filePath", filePath, "err", convertErr)
		return convertErr
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
	if hash == "" {
		return false, fmt.Errorf("hash cannot be empty")
	}
	fileDir, genFileDirErr := genPathDir(hash)
	if genFileDirErr != nil {
		return false, fmt.Errorf("generate pathDir failed: %w", genFileDirErr)
	}
	fileInfo, err := os.Stat(fileDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func (i *imageFileStore) Remove(fileHeader *multipart.FileHeader) error {
	//TODO implement me
	return nil
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
