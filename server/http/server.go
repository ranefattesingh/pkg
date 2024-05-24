package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ranefattesingh/pkg/log"
	"go.uber.org/zap"
)

type httpServer struct {
	addr       string
	server     *http.Server
	cleanupFns []func() error
}

func NewHTTPServer(host string, port int, cleanupFns ...func() error) *httpServer {
	return &httpServer{
		addr:       host + ":" + strconv.Itoa(port),
		cleanupFns: cleanupFns,
	}
}

func (h *httpServer) Start(handler http.Handler) error {
	h.server = &http.Server{
		Addr:    h.addr,
		Handler: handler,
	}

	log.Info("starting server on ", zap.String("addr", h.addr))

	return h.server.ListenAndServe()
}

func (h *httpServer) Shutdown(shutdown chan<- struct{}, cancelFn context.CancelFunc,
) {
	defer close(shutdown)

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	cancelFn()

	log.Info("started cleaning up server resources")

	// Server resources cleanup.
	for _, cleanupFunction := range h.cleanupFns {
		if err := cleanupFunction(); err != nil {
			log.Error("resource cleanhup", zap.Error(err))
		}
	}

	h.shutdownHTTPServer()
}

func (h *httpServer) shutdownHTTPServer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if h.server != nil {
		if err := h.server.Shutdown(ctx); err != nil {
			log.Error("server shutdown : %v", zap.Error(err))

			return
		}

		log.Info("server shutdown complete")
	}
}
