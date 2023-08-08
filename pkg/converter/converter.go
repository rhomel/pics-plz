package converter

import (
	"strings"
	"sync"

	"gopkg.in/gographics/imagick.v2/imagick"

	"github.com/rhomel/pics-plz/pkg/config"
	"github.com/rhomel/pics-plz/pkg/file"
)

type Logger interface {
	Println(v ...any)
}

type Converter struct {
	logger Logger
	config *config.Config

	mut *sync.Mutex
}

func NewConverter(logger Logger, config *config.Config) *Converter {
	return &Converter{
		logger: logger,
		config: config,
		mut:    &sync.Mutex{},
	}
}

func (c *Converter) log(v ...any) {
	c.logger.Println(v...)
}

func (c *Converter) Convert(sourcePath string, targetPath string) error {
	c.mut.Lock()
	defer c.mut.Unlock()
	if file.Exists(targetPath) && !file.IsDirectory(targetPath) {
		c.log("[converter:already converted]", targetPath)
		return nil
	}
	imagick.Initialize()
	defer imagick.Terminate()
	args := []string{
		"convert", sourcePath, targetPath,
	}
	c.log("[converter:start]", strings.Join(args, " "))
	_, err := imagick.ConvertImageCommand(args)
	if err != nil {
		return err
	}
	c.log("[converter:finished]", strings.Join(args, " "))
	return nil
}
