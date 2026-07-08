package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dkoshenkov/packages-go/logx"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "api server stopped: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	application, cleanup, err := initializeApplication(ctx)
	if err != nil {
		return err
	}
	defer cleanup()
	ctx = logx.WithContext(ctx, application.Logger)

	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		logx.Info(ctx).Str("addr", application.HTTPServer.Addr).Msg("api server listening")
		errc <- application.HTTPServer.ListenAndServe()
	}()

	select {
	case <-runCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownCtx = logx.WithContext(shutdownCtx, application.Logger)
		return application.HTTPServer.Shutdown(shutdownCtx)
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
