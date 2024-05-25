package http

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type httpServer struct {
	addr             string
	server           *http.Server
	cleanupFns       []func() error
	failedCleanupFns []func() error
}

func NewHTTPServer(host string, port int, cleanupFns ...func() error) *httpServer {
	return &httpServer{
		addr:             host + ":" + strconv.Itoa(port),
		cleanupFns:       cleanupFns,
		failedCleanupFns: make([]func() error, 0),
	}
}

func (h *httpServer) start(handler http.Handler) error {
	h.server = &http.Server{
		Addr:    h.addr,
		Handler: handler,
	}

	return h.server.ListenAndServe()
}

func (h *httpServer) shutdown(shutdown chan<- error, cancelFn context.CancelFunc,
) {
	defer close(shutdown)

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	cancelFn()

	// Server resources cleanup.
	for _, cleanupFunction := range h.cleanupFns {
		if err := cleanupFunction(); err != nil {
			h.failedCleanupFns = append(h.failedCleanupFns, cleanupFunction)
		}
	}

	shutdownErr := h.shutdownHTTPServer()

	retryFailedCleanupFns(shutdown, h.failedCleanupFns)

	shutdown <- shutdownErr
}

func (h *httpServer) shutdownHTTPServer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if h.server != nil {
		if err := h.server.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (h *httpServer) Start(handler http.Handler, cancelFn context.CancelFunc) error {
	errChan := make(chan error, 1)

	go h.shutdown(errChan, cancelFn)

	if err := h.start(handler); errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return <-errChan
}

func retryFailedCleanupFns(shutdown chan<- error, failedCleanupFns []func() error) {
	// Retry failed cleanup functions
	errorQ := make([]error, 0)
	for _, cleanupFunction := range failedCleanupFns {
		if err := cleanupFunction(); err != nil {
			errorQ = append(errorQ, err)
		}
	}

	for _, err := range errorQ {
		shutdown <- err
	}
}
