package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func Extension(path string) string {
	return strings.ToUpper(filepath.Ext(path))
}
