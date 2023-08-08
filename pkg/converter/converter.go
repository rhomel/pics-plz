package converter

import (
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/gographics/imagick.v2/imagick"

	"github.com/rhomel/pics-plz/pkg/file"
)

type Logger interface {
	Println(v ...any)
}

type Converter struct {
	logger Logger

	mut *sync.Mutex
}

func NewConverter(logger Logger) *Converter {
	return &Converter{
		logger: logger,
		mut:    &sync.Mutex{},
	}
}

func (c *Converter) log(v ...any) {
	c.logger.Println(v...)
}

func (c *Converter) Convert(sourcePath string, newTargetExtension string) string {
	c.mut.Lock()
	defer c.mut.Unlock()
	filename := strings.TrimSuffix(sourcePath, filepath.Ext(sourcePath)) + "." + newTargetExtension
	if file.Exists(filename) {
		c.log("[converter:already converted]", filename)
		return filename
	}
	imagick.Initialize()
	defer imagick.Terminate()
	args := []string{
		"convert", sourcePath, filename,
	}
	c.log("[converter:start]", strings.Join(args, " "))
	_, err := imagick.ConvertImageCommand(args)
	if err != nil {
		panic(err)
	}
	c.log("[converter:finished]", strings.Join(args, " "))
	return filename
}
