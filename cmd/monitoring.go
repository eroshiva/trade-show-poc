// Package main is the main entry point to the service
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"entgo.io/ent/dialect"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
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
		return filepath.Dir(fmt.Sprintf("%s", i)) + filepath.Base(fmt.Sprintf("%s", i))
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

	client := runSchemaMigration()
	// gracefully close client
	defer func() {
		err := client.Close()
		if err != nil {
			zlog.Error().Err(err).Msg("failed closing postgres client")
		}
	}()

	// creating three endpoints - two for first device, one for the second device
	ep1, err := db.CreateEndpoint(context.Background(), client, "192.168.0.1", "532", endpoint.ProtocolPROTOCOL_SNMP)
	if err != nil {
		return
	}
	zlog.Info().Msgf("Endpoint has been created in the DB: %v", ep1)

	ep2, err := db.CreateEndpoint(context.Background(), client, "192.168.0.2", "1084", endpoint.ProtocolPROTOCOL_NETCONF)
	if err != nil {
		return
	}
	zlog.Info().Msgf("Endpoint has been created in the DB: %v", ep2)

	ep3, err := db.CreateEndpoint(context.Background(), client, "192.168.0.3", "532", endpoint.ProtocolPROTOCOL_SNMP)
	if err != nil {
		return
	}
	zlog.Info().Msgf("Endpoint has been created in the DB: %v", ep3)

	// creating Network device
	nd, err := db.CreateNetworkDevice(context.Background(), client, uuid.New().String(), "MODEL-XYZ", "HW-XYZ", networkdevice.VendorVENDOR_UBIQUITI, []*ent.Endpoint{ep1, ep2})
	if err != nil {
		return
	}
	zlog.Info().Msgf("Network device has been created in the DB: %v", nd)

	// creating another Network device
	nd2, err := db.CreateNetworkDevice(context.Background(), client, uuid.New().String(), "MODEL-YZX", "HW-YZX", networkdevice.VendorVENDOR_JUNIPER, []*ent.Endpoint{ep3})
	if err != nil {
		return
	}
	zlog.Info().Msgf("Another Network device has been created in the DB: %v", nd2)

	// retrieving other network device
	ndBack, err := db.GetNetworkDeviceByID(context.Background(), client, nd2.ID)
	if err != nil {
		return
	}
	zlog.Info().Msgf("Returned network device is: %v", ndBack)

	ndBack1, err := db.GetNetworkDeviceByEndpoint(context.Background(), client, ep2.Host, ep2.Port)
	if err != nil {
		return
	}
	zlog.Info().Msgf("Returned network device is: %v", ndBack1)

	ndBack2, err := db.GetNetworkDeviceByEndpoint(context.Background(), client, ep3.Host, ep3.Port)
	if err != nil {
		return
	}
	zlog.Info().Msgf("Returned network device is: %v", ndBack2)

	// this must be an error
	ndBackErr, err := db.GetNetworkDeviceByEndpoint(context.Background(), client, ep1.Host, ep2.Port)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to return a network device")
	}
	zlog.Info().Msgf("%v", ndBackErr)

	zlog.Info().Msgf("starting server")
	server.StartServer(termChan, readyChan)
}

func runSchemaMigration() *ent.Client {
	zlog.Debug().Msgf("Opening connection to PostreSQL...")
	client, err := ent.Open(dialect.Postgres, "host=localhost port=5432 user=admin dbname=postgres password=pass")
	if err != nil {
		zlog.Fatal().Msgf("failed opening connection to postgres: %v", err)
	}

	zlog.Debug().Msgf("Migrating database schema...")
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		zlog.Fatal().Msgf("failed creating schema resources: %v", err)
	}

	return client
}
