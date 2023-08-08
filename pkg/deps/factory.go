package deps

import (
	"log"

	"github.com/rhomel/pics-plz/pkg/converter"
)

type Deps interface {
	LoggerProvider
	ConverterProvider
}

func Defaults() *DefaultProvider {
	p := &DefaultProvider{}
	p.logger = log.Default()
	p.converter = converter.NewConverter(p.logger)
	return p
}

type DefaultProvider struct {
	logger    Logger
	converter Converter
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
	Convert(sourcePath string, newTargetExtension string) string
}
