package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rhomel/pics-plz/pkg/deps"
	"github.com/rhomel/pics-plz/pkg/file"
	"github.com/rhomel/pics-plz/pkg/resources"
)

type browser struct {
	logger           deps.Logger
	imageAbsRoot     string
	imagePathPrefix  string
	browsePathPrefix string
	header           []byte
	footer           []byte
}

func NewBrowser(logger deps.Logger, imageAbsRoot, imagePathPrefix, browsePathPrefix string) *browser {
	return &browser{
		logger:           logger,
		imageAbsRoot:     imageAbsRoot,
		imagePathPrefix:  imagePathPrefix,
		browsePathPrefix: browsePathPrefix,
		header: []byte(`<!doctype html>
<html lang=en>
<head>
<meta charset=utf-8>
<script src="https://cdn.tailwindcss.com"></script>
<title>Picture Browser</title>
</head>
<body style="background: black;">
<div class="m-4">
<div class="flex flex-wrap w-full gap-4">
		`),
		footer: []byte(`</div></div></body></html>`),
	}
}

func (b *browser) Browse(response http.ResponseWriter, path string) {
	dir := filepath.Join(b.imageAbsRoot, path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		b.errorf(response, http.StatusInternalServerError, "internal server error")
		return
	}
	buffer := &bytes.Buffer{}
	for _, entry := range entries {
		href := filepath.Join(path, entry.Name())
		extension := file.Extension(entry.Name())
		if entry.IsDir() {
			b.renderDirectory(buffer, href, entry.Name())
		}
		if !resources.IsAllowed(extension) {
			continue
		}
		b.renderImage(buffer, href)
	}
	if _, err := response.Write(b.header); err != nil {
		b.logger.Printf("failed writing to response: %v", err)
	}
	if _, err := response.Write(buffer.Bytes()); err != nil {
		b.logger.Printf("failed writing to response: %v", err)
	}
	if _, err := response.Write(b.footer); err != nil {
		b.logger.Printf("failed writing to response: %v", err)
	}
}

func (b *browser) renderDirectory(w io.Writer, path string, name string) {
	href := filepath.Join(b.browsePathPrefix, path)
	//path = url.QueryEscape(href) // need to ignore slashes?
	fmt.Fprintf(w, "<a href=\"%s\"><div class=\"flex-initial flex items-center justify-center w-80 h-60 border border-blue-600 text-slate-100\">%s</div></a>", href, name)
}

func (b *browser) renderImage(w io.Writer, path string) {
	href := filepath.Join(b.imagePathPrefix, path)
	//href = url.QueryEscape(href)
	fmt.Fprintf(w, "<img class=\"flex-initial items-center justify-center w-80 h-60\" style=\"object-fit:contain\" src=\"%s\" />", href)
}

func (b *browser) errorf(response http.ResponseWriter, code int, message string) {
	response.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(response, "%d: %s", code, message)
}
