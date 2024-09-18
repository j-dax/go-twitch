package main

import (
	"context"
	"fmt"
	"helix/auth"
	"helix/dotenv"
	"helix/routes"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func openLinkInBrowser(link string) error {
	var err error = nil

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", link).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", link).Start()
	case "darwin":
		err = exec.Command("open", link).Start()
	default:
		err = fmt.Errorf("Unrecognized operating system: open this link in your browser\n\t%s\n", link)
	}

	return err
}

func NewAuthServer(
	logger *log.Logger,
	config dotenv.ServerConfig,
) http.Handler {
	mux := http.NewServeMux()
	routes.AddRoutes(mux, logger)
	var handler http.Handler = mux
	return handler
}

func runServer(logging *log.Logger, ctx context.Context) error {
	conf := dotenv.GetServerConfig()
	srv := NewAuthServer(logging, conf)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(conf.Host, conf.Port),
		Handler: srv,
	}

	// listen for close
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// have user pop browser to initiate authorization process
	if err := openLinkInBrowser(fmt.Sprintf("http://%s:%s/login", conf.Host, conf.Port)); err != nil {
		return err
	}

	// start the server
	go func() {
		logging.Printf("Listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Printf("Server error: %s\n", err)
		}
	}()

	// Wait for interrupt
	<-quit
	logging.Println("Shutdown signal received...")

	// Timeout to shutdown
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Do shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("Server force closed. %s", err)
	}

	return nil
}

func run(ctx context.Context) error {
	logging := log.Default()
	if err := dotenv.Load(".env"); err != nil {
		return err
	}

	if err := auth.ValidateAccess(); err != nil {
		dotenv.DefaultServerConfig()
		// run the server first to be ready for the callback
		if err := runServer(logging, ctx); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
