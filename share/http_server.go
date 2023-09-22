package chshare

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/riportdev/riport/share/logger"
)

const readHeaderTimeout = 5 * time.Second

type ServerOption func(*HTTPServer)

func WithTLS(certFile string, keyFile string, tlsConfig *tls.Config) ServerOption {
	return func(s *HTTPServer) {
		s.certFile = certFile
		s.keyFile = keyFile
		s.TLSConfig = tlsConfig
	}
}

// HTTPServer extends net/http Server and
// adds graceful shutdowns
type HTTPServer struct {
	*http.Server
	listener  net.Listener
	ctx       context.Context
	running   chan error
	isRunning bool
	certFile  string
	keyFile   string
	logger    *logger.Logger
}

// NewHTTPServer creates a new HTTPServer
func NewHTTPServer(maxHeaderBytes int, l *logger.Logger, options ...ServerOption) *HTTPServer {
	var httpLogger *logger.Logger
	if l != nil {
		httpLogger = l.Fork("http-server")
	}
	s := &HTTPServer{
		Server: &http.Server{
			MaxHeaderBytes:    maxHeaderBytes,
			ReadHeaderTimeout: readHeaderTimeout,
		},
		listener: nil,
		running:  make(chan error, 1),
		logger:   httpLogger,
	}

	for _, o := range options {
		o(s)
	}

	return s
}

func (h *HTTPServer) GoListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	h.isRunning = true
	h.ctx = ctx
	h.Handler = handler
	h.listener = l
	h.BaseContext = func(l net.Listener) context.Context {
		return h.ctx
	}

	go func() {
		if h.TLSConfig != nil {
			h.logger.Debugf("serving HTTPS")
			h.closeWith(h.ServeTLS(l, h.certFile, h.keyFile))
		} else {
			h.logger.Debugf("serving HTTP")
			h.closeWith(h.Serve(l))
		}
	}()
	return nil
}

func (h *HTTPServer) closeWith(err error) {
	if !h.isRunning {
		return
	}
	h.isRunning = false
	h.running <- err
}

func (h *HTTPServer) Close() error {
	h.closeWith(nil)
	if h.listener == nil {
		return nil
	}
	return h.listener.Close()
}

func (h *HTTPServer) Wait() error {
	if !h.isRunning {
		return errors.New("Already closed")
	}
	select {
	case <-h.running:
		return nil
	case <-h.ctx.Done():
		h.logger.Debugf("context canceled")
		return h.ctx.Err()
	}
}
