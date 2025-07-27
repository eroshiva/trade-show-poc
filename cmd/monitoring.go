// Package main is the main entry point to the service
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/rs/zerolog"
)

const (
	component     = "component"
	componentName = "main"
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

func main() {
	// channels to handle termination and capture signals
	termChan := make(chan bool)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	readyChan := make(chan bool, 1)

	dbClient, err := db.RunSchemaMigration()
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to instantiate connection with PostgreSQL DB")
	}

	go func() {
		<-sigChan
		close(termChan)

		// gracefully closing client at the very end of execution
		_ = db.GracefullyCloseDBClient(dbClient)
	}()

	// starting NB API server (user interactions and creation of resource).
	server.StartServer(dbClient, termChan, readyChan)

	// starting SB handler (updates device status and other monitoring information)
	// controller.StartController(dbClient)
}
