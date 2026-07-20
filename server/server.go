package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Mux struct {
	*http.ServeMux
	db *sql.DB
}

func NewMux(db *sql.DB) *Mux {
	mux := &Mux{
		ServeMux: http.NewServeMux(),
		db:       db,
	}

	return mux
}

func NewServer(handler http.Handler) *http.Server {
	server := &http.Server{
		Addr:         ":3000",
		Handler:      handler,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
	}

	return server
}

func StartServer(ctx context.Context, server *http.Server) error {
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Starting server at http://localhost%s...", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(
			err, http.ErrServerClosed,
		) {
			errChan <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case err := <-errChan:
		return err
	case <-sigCtx.Done():
		log.Println("Received termination signal, shutting down.")
	}

	shutDownCtx, cancel := context.WithTimeout(
		ctx,
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutDownCtx); err != nil {
		if closeErr := server.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}

	log.Println("Server exited gracefully.")
	return nil
}
