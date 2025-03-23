package server

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewHTTP2(port int, handler http.Handler) *http.Server {
	h2s := &http2.Server{}

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           h2c.NewHandler(handler, h2s),
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
}
