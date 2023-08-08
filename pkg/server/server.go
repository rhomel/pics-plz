package server

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
	"time"

	"github.com/rhomel/pics-plz/pkg/deps"
	"github.com/rhomel/pics-plz/pkg/resources"
)

const (
	DefaultImagePathPrefix = "/images"
)

type server struct {
	deps       deps.Deps
	httpServer *http.Server

	imagePathPrefix           string
	imageConverterCachePrefix string
}

func New(imageRoot string) (*server, error) {
	root, err := filepath.Abs(imageRoot)
	if err != nil {
		log.Printf("%s failed to resolve to an absolute path: %v", imageRoot, err)
		return nil, err
	}
	defaults, err := deps.Defaults()
	if err != nil {
		log.Printf("failed to configure dependencies: %v", err)
		return nil, err
	}
	defaults.Config().ImageRoot = root
	s := &server{
		deps:            defaults,
		imagePathPrefix: DefaultImagePathPrefix,
	}
	s.httpServer = &http.Server{
		Handler:        s,
		Addr:           ":8080",
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	return s, nil
}

func (s *server) Serve() {
	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		if err := s.httpServer.Shutdown(context.Background()); err != nil {
			s.logf("shutdown failed: %v", err)
		}
		close(idleConnsClosed)
	}()
	s.logf("http://localhost%s\n", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil {
		s.logf(err.Error())
	}
	<-idleConnsClosed
	s.logf("shutdown")
}

func (s *server) getImagesSubPath(request *http.Request) (string, bool) {
	if request.Method != "GET" {
		return "", false
	}
	path, ok := strings.CutPrefix(request.URL.Path, s.imagePathPrefix)
	if !ok {
		return "", false
	}
	return path, true
}

func (s *server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	s.deps.Logger().Println(request.Method, request.URL.Path)
	if imagePath, ok := s.getImagesSubPath(request); ok {
		s.ServeImage(response, request, imagePath)
		return
	}
	response.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(response, "400 Bad Request")
}

func (s *server) ServeError(response http.ResponseWriter, code int, err error) {
	message := fmt.Sprintf("%d: %s", code, err.Error())
	if code == http.StatusInternalServerError {
		s.errorf("unexpected error: %v", err)
		message = "500 Internal Server Error"
	}
	response.WriteHeader(code)
	fmt.Fprintf(response, message)
	s.logf("error response: %s", message)
}

func (s *server) ServeImage(response http.ResponseWriter, request *http.Request, target string) {
	s.logf("requested image: %s", target)
	image, err := resources.NewImage(target, s.deps.Converter(), s.deps.Config())
	if errors.Is(err, resources.NotFound) {
		s.ServeError(response, http.StatusNotFound, err)
		return
	}
	if errors.Is(err, resources.NotAllowed) {
		s.ServeError(response, http.StatusForbidden, err)
		return
	}
	if err != nil {
		s.ServeError(response, http.StatusInternalServerError, err)
		return
	}
	servableImage := image.GetServableImage()
	fh, err := os.Open(servableImage.Path)
	if err != nil {
		s.ServeError(response, http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := fh.Close(); err != nil {
			s.errorf("fh.Close(): %v", err)
		}
	}()
	response.Header().Add("content-type", servableImage.ContentType)
	if _, err := io.Copy(response, fh); err != nil {
		s.errorf("io.Copy(): %v", err)
		return
	}
}

func (s *server) logf(fmt string, v ...any) {
	s.deps.Logger().Printf(fmt, v...)
}

func (s *server) errorf(fmt string, v ...any) {
	s.deps.Logger().Printf("[ERROR] "+fmt, v...)
}
