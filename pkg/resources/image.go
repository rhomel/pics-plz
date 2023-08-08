package resources

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/rhomel/pics-plz/pkg/config"
	"github.com/rhomel/pics-plz/pkg/file"
)

var (
	NotFound   error = errors.New("not found")
	NotAllowed error = errors.New("not allowed")
)

func shouldConvert(extension string) bool {
	return conversionMap(extension) != ""
}

func conversionMap(extension string) string {
	switch extension {
	case ".HEIC":
		return "JPEG"
	default:
		return ""
	}
}

type Converter interface {
	Convert(sourcePath string, newTargetExtension string) error
}

type Image struct {
	requestedPath    string
	requestedAbsPath string
	convertedPath    string
	isConverted      bool
}

func NewImage(requestedPath string, converter Converter, config *config.Config) (*Image, error) {
	i := &Image{
		requestedPath:    requestedPath,
		requestedAbsPath: filepath.Join(config.ImageRoot, requestedPath),
		isConverted:      false,
	}
	if !file.Exists(i.requestedAbsPath) {
		return nil, NotFound
	}
	if !IsAllowed(i.requestedAbsPath) {
		return nil, NotAllowed
	}
	if shouldConvert(i.RequestedExtension()) {
		if err := i.prepareConvertedPath(config); err != nil {
			return nil, err
		}
		if !i.isConverted {
			if err := i.convert(converter); err != nil {
				return nil, err
			}
		}
	}
	return i, nil
}

func (i *Image) prepareConvertedPath(config *config.Config) error {
	conversionExtension := conversionMap(i.RequestedExtension())
	// replace the extension
	filename := strings.TrimSuffix(i.requestedPath, filepath.Ext(i.requestedPath)) + "." + conversionExtension
	filename = filepath.Join(config.ConvertedImageCachePath, filename)
	if file.Exists(filename) && !file.IsDirectory(filename) {
		i.convertedPath = filename
		i.isConverted = true
		return nil
	}
	dirname := filepath.Dir(filename)
	if file.Exists(dirname) && !file.IsDirectory(dirname) {
		// the cache is invalid? so remove the directory and create a new one
		if err := os.Remove(dirname); err != nil {
			return err
		}
	}
	if !file.Exists(dirname) {
		if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
			return err
		}
		// cache directory should now exist
	}
	i.convertedPath = filename
	return nil
}

func (i *Image) convert(converter Converter) error {
	err := converter.Convert(i.requestedAbsPath, i.convertedPath)
	if err != nil {
		return err
	}
	i.isConverted = true
	return nil
}

func (i *Image) RequestedExtension() string {
	return file.Extension(i.requestedPath)
}

func (i *Image) IsConverted() bool {
	return i.isConverted
}

func (i *Image) GetServableImage() *ServableImage {
	si := &ServableImage{
		Path: i.getServableResourcePath(),
	}
	si.ContentType = ContentType(si.Path)
	return si
}

func (i *Image) getServableResourcePath() string {
	if i.IsConverted() {
		return i.convertedPath
	}
	return i.requestedAbsPath
}

func (i *Image) GetRequestedPath() string {
	return i.requestedPath
}

func (i *Image) GetConvertedPath() string {
	return i.convertedPath
}

type ServableImage struct {
	Path        string
	ContentType string
}

func IsAllowed(requestedPath string) bool {
	switch file.Extension(requestedPath) {
	case ".GIF":
		return true
	case ".PNG":
		return true
	case ".JPEG":
		return true
	case ".JPG":
		return true
	case ".WEBP":
		return true
	case ".HEIC":
		return true
	default:
		return false
	}
}

func ContentType(target string) string {
	switch file.Extension(target) {
	case ".GIF":
		return "image/gif"
	case ".PNG":
		return "image/png"
	case ".JPEG":
		fallthrough
	case ".JPG":
		return "image/jpeg"
	case ".WEBP":
		return "image/webp"
	default:
		return ""
	}
}
