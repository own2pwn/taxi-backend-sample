package api

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

// Server - структура для сервера, который будет обслуживать запросы
type Server struct {
	server *http.Server
}

// NewServer инициализация севрера с настройками и зависимостями
func NewServer(addr string, h http.Handler) *Server {
	return &Server{server: &http.Server{Addr: addr, Handler: h}}
}

// Start taxa сервера
func (s *Server) Start() error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.server.ListenAndServe()
	}()
	if err := <-errChan; err != nil {
		if err != http.ErrServerClosed {
			return errors.Wrap(err, "Taxa server error")
		}
	}
	return nil
}

// Stop taxa сервера c плавным завершением всех входящих запросов
func (s *Server) Stop() error {

	if err := s.server.Shutdown(context.Background()); err != nil {
		return errors.Wrap(err, "could not shutdown taxa server")
	}
	return nil
}
