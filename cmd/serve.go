package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve [directory]",
	Short: "Start the labeling server without opening a browser",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		return nil
	},
}
