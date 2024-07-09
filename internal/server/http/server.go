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

func writeResponse(w http.ResponseWriter, _ error, ans *model.Balance, s *Server) {
	responseBytes, err := json.Marshal(map[string]interface{}{"balance": ans})
	if err != nil {
		s.log.Errorf("Can't marshal response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error\": \"Can't process response\"}\n"))
		return
	}
	w.Write(responseBytes)
}

func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {
	var balance model.Balance
	if err := s.helperDecode(r.Body, w, &balance); err != nil {
		return
	}

	currency := r.URL.Query().Get("currency")

	ans, err := s.app.GetBalance(r.Context(), &balance)
	if err != nil {
		s.log.Errorf("Can't get balance:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't get balance:%v\"}\n", err)))
		return
	}

	if currency != "" {
		ans, err = s.app.ConvertBalance(r.Context(), ans, currency)
		if err != nil {
			s.log.Errorf("Can't convert balance:%v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't convert balance:%v\"}\n", err)))
			return
		}
	}
	writeResponse(w, err, ans, s)
}

func (s *Server) GetTransactions(w http.ResponseWriter, r *http.Request) {
	var balance model.Balance
	if err := s.helperDecode(r.Body, w, &balance); err != nil {
		return
	}

	sort := r.URL.Query().Get("sort")

	ans, err := s.app.GetTransactions(r.Context(), balance.UserID, sort)
	if err != nil {
		s.log.Errorf("Can't get transactions:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't get transactions:%v\"}\n", err)))
		return
	}
	responseBytes, err := json.Marshal(map[string]interface{}{"transactions": ans})
	if err != nil {
		s.log.Errorf("Can't marshal response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error\": \"Can't process response\"}\n"))
		return
	}
	w.Write(responseBytes)
}

func (s *Server) TopUp(w http.ResponseWriter, r *http.Request) {
	var balance model.Balance
	if err := s.helperDecode(r.Body, w, &balance); err != nil {
		return
	}
	ans, err := s.app.TopUp(r.Context(), balance.UserID, balance.Amount)
	if err != nil {
		s.log.Errorf("Can't Top up:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't Top up:%v\"}\n", err)))
		return
	}
	writeResponse(w, err, ans, s)
}

func (s *Server) Debit(w http.ResponseWriter, r *http.Request) {
	var balance model.Balance
	if err := s.helperDecode(r.Body, w, &balance); err != nil {
		return
	}
	ans, err := s.app.Debit(r.Context(), balance.UserID, balance.Amount)
	if err != nil {
		s.log.Errorf("Can't Debit:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't Debit:%v\"}\n", err)))
		return
	}
	writeResponse(w, err, ans, s)
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
	mux.Handle("/top-up", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.TopUp))))
	mux.Handle("/debit", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.Debit))))
	mux.Handle("/transaction", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.GetTransactions))))
	/*mux.Handle("/transfer", midLogger.setCommonHeadersMiddleware(
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
