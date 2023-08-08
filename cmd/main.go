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
	"time"

	"github.com/rhomel/pics-plz/pkg/deps"
	"github.com/rhomel/pics-plz/pkg/resources"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <source-path-to-serve>\n", os.Args[0])
		os.Exit(1)
	}
	sourcePath := os.Args[1]
	s := NewServer(sourcePath)
	s.Serve()
}

type server struct {
	root       string
	deps       deps.Deps
	httpServer *http.Server
}

func NewServer(localRoot string) *server {
	root, err := filepath.Abs(localRoot)
	if err != nil {
		log.Fatalf("%s failed to resolve to an absolute path: %v", localRoot, err)
	}
	s := &server{
		root: root,
		deps: deps.Defaults(),
	}
	s.httpServer = &http.Server{
		Handler:        s,
		Addr:           ":8080",
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	return s
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

func (s *server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	target := filepath.Join(s.root, request.URL.Path)
	s.deps.Logger().Println(request.Method, request.URL.Path, target)
	if request.Method == "GET" {
		s.ServeImage(response, request, target)
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
	image, err := resources.NewImage(target, s.deps.Converter())
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
