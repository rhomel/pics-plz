package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/gographics/imagick.v2/imagick"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <source-path-to-serve>\n", os.Args[0])
		os.Exit(1)
	}
	sourcePath := os.Args[1]
	serve(sourcePath)
}

type server struct {
	root         string
	convertMutux *sync.Mutex
}

func (s *server) isAllowed(target string) bool {
	switch extension(target) {
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

func (s *server) contentType(target string) string {
	switch extension(target) {
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

func (s *server) shouldConvert(target string) bool {
	return extension(target) == ".HEIC"
}

func (s *server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	target := filepath.Join(s.root, request.URL.Path)
	log.Println(request.Method, request.URL.Path, target)
	if request.Method == "GET" {
		s.ServeImage(response, request, target)
		return
	}
	response.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(response, "400 Bad Request")
}

func (s *server) ServeImage(response http.ResponseWriter, request *http.Request, target string) {
	if !exists(target) {
		response.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(response, "404 Not Found")
		return
	}
	if !s.isAllowed(target) {
		response.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(response, "403 Forbidden")
		return
	}
	if s.shouldConvert(target) {
		// FIXME: slow for now but prevents crashes with multiple imagemagick
		// commands
		// Ideally we only run convert once per an image but if we receive
		// multiple requests for different images, then they should be allowed
		// to run in parallel up to a maximum (where max == number of physical
		// cpu cores).
		s.convertMutux.Lock()
		target = convert(target, "JPEG")
		s.convertMutux.Unlock()
	}
	log.Println("target:", target, "content-type:", s.contentType(target))
	fh, err := os.Open(target)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(response, "500 Internal Server Error")
		log.Printf("os.Open('%s') ERROR: %v\n", target, err)
		return
	}
	defer func() {
		if err := fh.Close(); err != nil {
			log.Println("fh.Close() ERROR:", err)
		}
	}()
	response.Header().Add("content-type", s.contentType(target))
	if _, err := io.Copy(response, fh); err != nil {
		log.Printf("io.Copy() ERROR: %v\n", err)
		return
	}
}

func serve(localRoot string) {
	root, err := filepath.Abs(localRoot)
	if err != nil {
		log.Fatalf("%s failed to resolve to an absolute path: %v", localRoot, err)
	}
	s := &server{
		root:         root,
		convertMutux: &sync.Mutex{},
	}
	httpServer := &http.Server{
		Handler:        s,
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Println("shutdown failed:", err)
		}
		close(idleConnsClosed)
	}()
	log.Printf("http://localhost%s\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Println(err)
	}
	<-idleConnsClosed
	log.Println("shutdown")
}

func convert(source string, targetExtension string) string {
	filename := strings.TrimSuffix(source, filepath.Ext(source)) + "." + targetExtension
	if exists(filename) {
		return filename
	}
	imagick.Initialize()
	defer imagick.Terminate()
	args := []string{
		"convert", source, filename,
	}
	log.Println("[converter:start]", strings.Join(args, " "))
	_, err := imagick.ConvertImageCommand(args)
	if err != nil {
		panic(err)
	}
	log.Println("[converter:finished]", strings.Join(args, " "))
	return filename
}

func exists(target string) bool {
	_, err := os.Stat(target)
	return !errors.Is(err, os.ErrNotExist)
}

func extension(target string) string {
	return strings.ToUpper(filepath.Ext(target))
}
