package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

func (s *APIServer) serverError(w http.ResponseWriter, err error, r *http.Request) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	s.logWarn(http.StatusInternalServerError, r, err, trace)
	s.Logger.Error(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (s *APIServer) clientError(w http.ResponseWriter, r *http.Request, status int, err error, info string) {
	s.logWarn(status, r, err, info)
	http.Error(w, http.StatusText(status), status)
}

func (s *APIServer) notFound(w http.ResponseWriter, err error, info string, r *http.Request) {
	s.clientError(w, r, http.StatusNotFound, err, info)
}

func (s *APIServer) badRequest(w http.ResponseWriter, err error, info string, r *http.Request) {
	s.clientError(w, r, http.StatusBadRequest, err, info)
}

// Функция заносит в лог все запросы полученные сервером используя http.Request
// для получения большей части информации.
func (s *APIServer) logWarn(status int, r *http.Request, err error, info string) {
	msg := ""
	if err != nil {
		msg = err.Error()
	} else {
		msg = ""
	}
	s.Logger.Info(msg,
		zap.Int("status_code", status),
		zap.String("info", info),
		zap.String("request_method", r.Method),
		zap.String("request_uri", r.RequestURI),
		zap.String("source_ip", r.RemoteAddr),
	)
}

func (s *APIServer) logRequest(r *http.Request) {
	s.Logger.Info("success",
		zap.Int("status_code", http.StatusOK),
		zap.String("request_method", r.Method),
		zap.String("request_uri", r.RequestURI),
		zap.String("source_ip", r.RemoteAddr),
	)
}
