package internalhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cronnoss/avitotech/internal/model"
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

func (s *Server) helperDecode(stream io.ReadCloser, w http.ResponseWriter, data interface{}) error {
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&data); err != nil {
		s.log.Errorf("Can't decode json:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't decode json:%v\"}\n", err)))
		return err
	}
	return nil
}

func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {
	var balance model.Balance
	if err := s.helperDecode(r.Body, w, &balance); err != nil {
		return
	}
	bal, err := s.app.GetBalance(r.Context(), &balance)
	if err != nil {
		s.log.Errorf("Can't get balance:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't get balance:%v\"}\n", err)))
		return
	}
	w.Write([]byte(fmt.Sprintf("{\"balance\": \"%s\"}\n", bal)))
}

func (s *Server) TopUp(w http.ResponseWriter, r *http.Request) {
	var balance model.Balance
	if err := s.helperDecode(r.Body, w, &balance); err != nil {
		return
	}
	err := s.app.TopUp(r.Context(), &balance)
	if err != nil {
		s.log.Errorf("Can't Top up:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't Top up:%v\"}\n", err)))
		return
	}
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

	mux.Handle("/balance", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.GetBalance))))
	/*mux.Handle("/transaction", midLogger.setCommonHeadersMiddleware(
	midLogger.loggingMiddleware(http.HandlerFunc(s.GetTransactions))))*/
	mux.Handle("/top-up", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.TopUp))))
	/*mux.Handle("/debit", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.Debit))))
	mux.Handle("/transfer", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.Transfer))))*/

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
