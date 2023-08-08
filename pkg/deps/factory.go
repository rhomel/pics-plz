package deps

import (
	"log"

	"github.com/rhomel/pics-plz/pkg/config"
	"github.com/rhomel/pics-plz/pkg/converter"
)

type Deps interface {
	LoggerProvider
	ConverterProvider
	ConfigProvider
}

func Defaults() (*DefaultProvider, error) {
	var err error
	p := &DefaultProvider{}
	p.config, err = config.NewConfig()
	if err != nil {
		return nil, err
	}
	p.logger = log.Default()
	p.converter = converter.NewConverter(p.logger, p.config)
	return p, nil
}

type DefaultProvider struct {
	config    *config.Config
	logger    Logger
	converter Converter
}

func (p *DefaultProvider) Config() *config.Config {
	return p.config
}

func (p *DefaultProvider) Logger() Logger {
	return p.logger
}

func (p *DefaultProvider) Converter() Converter {
	return p.converter
}

var _ Deps = (*DefaultProvider)(nil)

type LoggerProvider interface {
	Logger() Logger
}

type Logger interface {
	Print(v ...any)
	Println(v ...any)
	Printf(format string, v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
	Fatalln(v ...any)
}

type ConverterProvider interface {
	Converter() Converter
}

type Converter interface {
	Convert(sourcePath string, newPath string) error
}

type ConfigProvider interface {
	Config() *config.Config
}
