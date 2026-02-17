package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/monorkin/just-label-it/internal/browser"
	"github.com/spf13/cobra"
)

func runOpen(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	listener, _, err := startServer(dir)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s", listener.Addr().String())
	log.Printf("Server running at %s", url)

	if err := browser.Open(url); err != nil {
		log.Printf("Could not open browser: %v (visit %s manually)", err, url)
	}

	// Wait for interrupt.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Shutting down...")
	return nil
}
