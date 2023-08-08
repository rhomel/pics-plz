package resources

import (
	"errors"

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
	Convert(sourcePath string, newTargetExtension string) string
}

type Image struct {
	requestedPath string
	convertedPath string
	isConverted   bool
}

func NewImage(requestedPath string, converter Converter) (*Image, error) {
	if !file.Exists(requestedPath) {
		return nil, NotFound
	}
	if !IsAllowed(requestedPath) {
		return nil, NotAllowed
	}
	i := &Image{
		requestedPath: requestedPath,
		isConverted:   false,
	}
	if shouldConvert(i.RequestedExtension()) {
		i.convert(converter)
	}
	return i, nil
}

func (i *Image) convert(converter Converter) {
	conversionExtension := conversionMap(i.RequestedExtension())
	i.convertedPath = converter.Convert(i.requestedPath, conversionExtension)
	i.isConverted = true
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
	return i.requestedPath
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
