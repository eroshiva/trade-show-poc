// Package main is the main entry point to the service
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/manager"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/checksum"
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
	reverseProxyTermChan := make(chan bool)
	managerTermChan := make(chan bool)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	readyChan := make(chan bool, 1)
	reverseProxyReadyChan := make(chan bool, 1)

	dbClient, err := db.RunSchemaMigration()
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to instantiate connection with PostgreSQL DB")
	}

	// creating waitgroup so main will wait for servers to exit cleanly
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		<-sigChan
		close(termChan)
		close(reverseProxyTermChan)
		close(managerTermChan)

		// gracefully closing client at the very end of execution
		_ = db.GracefullyCloseDBClient(dbClient)
		wg.Done()
	}()

	// starting NB API server (user interactions and creation of resource).
	wg.Add(1)
	go func() {
		wg.Add(1) //nolint:staticcheck
		server.StartServer(dbClient, wg, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan)
		wg.Done()
	}()

	checksumGen := checksum.NewMockGenerator() // right now, invoking mock generator for smooth testing
	// starting SB handler (updates device status and other monitoring information)
	manager.StartManager(dbClient, checksumGen, managerTermChan)

	wg.Wait()
}
