package internalhttp

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/cronnoss/avitotech/internal/server"
)

type ctxKeyID int

const (
	KeyLoggerID ctxKeyID = iota
)

type Server struct {
	srv  http.Server
	app  server.Application
	log  Logger
	host string
	port string
}

type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

func NewServer(log Logger, app server.Application, host, port string) *Server {
	return &Server{log: log, app: app, host: host, port: port}
}

func (s *Server) Start(ctx context.Context) error {
	addr := net.JoinHostPort(s.host, s.port)
	midLogger := NewMiddlewareLogger()
	mux := http.NewServeMux()

	mux.Handle("/healthz", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK healthz\n"))
		}))))

	mux.Handle("/readiness", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK readiness\n"))
		}))))

	s.srv = http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			bCtx := context.WithValue(ctx, KeyLoggerID, s.log)
			return bCtx
		},
	}

	s.log.Infof("http server started on %s:%s\n", s.host, s.port)
	err := s.srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}
	s.log.Infof("http server shutdown\n")
	return nil
}
