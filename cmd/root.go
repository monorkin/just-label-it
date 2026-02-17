package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/monorkin/just-label-it/internal/db"
	"github.com/monorkin/just-label-it/internal/scanner"
	"github.com/monorkin/just-label-it/internal/server"
	"github.com/spf13/cobra"
)

var (
	flagBind string
	flagPort int
)

var rootCmd = &cobra.Command{
	Use:   "jli [directory]",
	Short: "Label media files for LLM training sets",
	Long:  "Just Label It — a tool for classifying images, video, and audio files with tags and descriptions.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runOpen,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagBind, "bind", "127.0.0.1", "Address to bind the server to")
	rootCmd.PersistentFlags().IntVar(&flagPort, "port", 0, "Port to listen on (0 for auto)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// startServer initializes the database, scans for media files, and starts the HTTP server.
// It returns the listener address so callers can open a browser if desired.
func startServer(dir string) (net.Listener, *http.Server, error) {
	dbPath := filepath.Join(dir, "jli.db")
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("opening database: %w", err)
	}

	files, err := scanner.Scan(dir)
	if err != nil {
		database.Close()
		return nil, nil, fmt.Errorf("scanning directory: %w", err)
	}

	for _, f := range files {
		if err := database.UpsertMediaFile(f.Path, f.MediaType); err != nil {
			log.Printf("warning: skipping %s: %v", f.Path, err)
		}
	}

	count, _ := database.MediaFileCount()
	log.Printf("Found %d media files in %s", count, dir)

	handler, err := server.New(database, dir)
	if err != nil {
		database.Close()
		return nil, nil, fmt.Errorf("creating server: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", flagBind, flagPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		database.Close()
		return nil, nil, fmt.Errorf("listening on %s: %w", addr, err)
	}

	srv := &http.Server{Handler: handler}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	return listener, srv, nil
}
