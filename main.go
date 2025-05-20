package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/polarisjrex0406/static-proxy/config"
	"github.com/polarisjrex0406/static-proxy/pkg"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	lc := zap.NewProductionConfig()
	if cfg.Debug {
		lc = zap.NewDevelopmentConfig()
		lc.Development = true
	}

	logger, err := lc.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() //nolint:errcheck

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Server.Port),
	}
	httpServer.Handler = http.HandlerFunc(pkg.HandlerHTTP)

	if cfg.Server.Port > 0 {
		logger.Info(fmt.Sprintf("Proxy: HTTP Starting :%d", cfg.Server.Port))
		defer logger.Info("Proxy: HTTP Stopped")

		err = pkg.ListenHTTP(ctx, httpServer)
		if err != http.ErrServerClosed {
			logger.Error("HTTP Proxy failed to listen", zap.Error(err))
		}
	}
}
