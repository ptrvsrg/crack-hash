package http

import (
	"fmt"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/di"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	cfg config.Config
}

func NewServer(c *di.Container) *http.Server {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Config.Server.Port),
		Handler:      SetupRouter(c),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
