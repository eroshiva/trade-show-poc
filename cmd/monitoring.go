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
	_ "github.com/lib/pq" // SQL driver, necessary for DB interaction
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
	go func() {
		<-sigChan
		close(termChan)
	}()

	readyChan := make(chan bool, 1)

	dbClient, err := db.RunSchemaMigration()
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to instantiate connection with PostgreSQL DB")
	}
	// gracefully closing client at the very end of execution
	defer func() {
		err = db.GracefullyCloseDBClient(dbClient)
		if err == nil {
			zlog.Info().Msg("Connection to the DB was gracefully closed.")
		}
		// in the opposite case, error is already logged in within function GracefullyCloseDBClient.
	}()

	// starting NB API server (user interactions and creation of resource).
	server.StartServer(dbClient, termChan, readyChan)

	// starting SB handler (updates device status and other monitoring information)
	// controller.StartController(dbClient)
}
