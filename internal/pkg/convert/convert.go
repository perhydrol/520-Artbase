package convert

import (
	"bytes"
	"demo520/internal/pkg/helper"
	"demo520/internal/pkg/log"
	"errors"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ImageConverter interface {
	ConvertImage(filePath string) error
	Shutdown()
}

var (
	once      sync.Once
	converter imageConverter
	c         config
)

type imageConverter struct{}

type config struct {
	WebPQuality        int
	WebReductionEffort int
	AvifQuality        int
	AvifEffort         int
	Lossless           bool
}

var _ ImageConverter = (*imageConverter)(nil)

func InitImageConverter() ImageConverter {
	once.Do(func() {
		converter = imageConverter{}
		c = config{
			WebPQuality:        viper.GetInt("WebPQuality"),
			WebReductionEffort: viper.GetInt("WebReductionEffort"),
			AvifQuality:        viper.GetInt("AvifQuality"),
			AvifEffort:         viper.GetInt("AvifEffort"),
			Lossless:           viper.GetBool("ImageLossless"),
		}
		vips.Startup(nil)
	})
	return &converter
}

func (i *imageConverter) ConvertImage(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file path is empty")
		}
		return err
	}
	if fileInfo.IsDir() {
		return errors.New("file path is directory")
	}
	ext := filepath.Ext(filepath.Base(filePath))
	filePathWithoutExt := strings.TrimSuffix(filepath.Base(filePath), ext)
	image, err := vips.NewImageFromFile(filePath)
	if err != nil {
		return err
	}
	defer image.Close()

	webp, _, err := image.ExportWebp(&vips.WebpExportParams{
		Quality:         c.WebPQuality,
		Lossless:        c.Lossless,
		StripMetadata:   true,
		ReductionEffort: c.WebReductionEffort,
	})
	if err != nil {
		log.Errorw("Failed to export webp", "err", err, filePathWithoutExt)
		return err
	}
	webpErr := helper.WriteFile(filePathWithoutExt+".webp", bytes.NewReader(webp))
	if webpErr != nil {
		return webpErr
	}

	avif, _, err := image.ExportAvif(&vips.AvifExportParams{
		Quality:       c.AvifQuality,
		Lossless:      c.Lossless,
		StripMetadata: true,
		Effort:        c.AvifEffort,
	})
	if err != nil {
		log.Errorw("Failed to export avif", "err", err, filePathWithoutExt)
		return err
	}
	avifErr := helper.WriteFile(filePath+".avif", bytes.NewReader(avif))
	if avifErr != nil {
		return avifErr
	}
	return nil
}

func (i *imageConverter) Shutdown() {
	vips.Shutdown()
}
