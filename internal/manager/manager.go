// Package manager implements SB control loop, that fetches data from the devices and stores it to the DB.
package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/rs/zerolog"
)

const (
	component     = "component"
	componentName = "control-loop"

	defaultControlLoopPerioud = 30 * time.Second
	envControlLoopPeriod      = "CONTROL_LOOP_PERIOD" // in seconds.
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// StartManager function starts main control loop that periodically fetches data from the network devices.
func StartManager(dbClient *ent.Client, termChan chan bool) {
	controlLoopTick := defaultControlLoopPerioud
	// read env variable, where Control Loop Period is specified
	envControlLoopTick := os.Getenv(envControlLoopPeriod)
	if envControlLoopTick == "" {
		zlog.Warn().Msgf("environment variable \"%s\" is not set, using default address: %s",
			envControlLoopPeriod, defaultControlLoopPerioud)
	} else {
		// control loop tick is specified
		duration, err := strconv.Atoi(envControlLoopTick)
		if err != nil {
			zlog.Fatal().Err(err).Msgf("failed to convert \"%s\"variable to number", envControlLoopPeriod)
		}
		controlLoopTick = time.Duration(duration) * time.Second
	}

	// creating ticker
	ticker := time.NewTicker(controlLoopTick)

	// starting infinite control loop
	for {
		select {
		case <-ticker.C:
			// performing control loop routine
			performControlLoopRoutine(dbClient, controlLoopTick)
		case <-termChan:
			// shutting down this routine
			return
		}
	}
}

// performControlLoopRoutine runs main control loop routine, i.e., fetches all devices and updates theirs status.
func performControlLoopRoutine(dbClient *ent.Client, controlLoopTick time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), controlLoopTick)
	defer cancel()

	// fetching all devices from the DB
	ndList, err := db.ListNetworkDevices(ctx, dbClient)
	if err != nil {
		// error is already logged in in the inner function
		return
	}

	// checking if we retrieved more than 0 network devices
	if len(ndList) == 0 {
		// no devices were retrieved, nothing to do
		zlog.Warn().Msgf("no network devices found in the DB")
		return
	}

	// updating network devices concurrently
	for _, nd := range ndList {
		go processNetworkDevice(ctx, dbClient, nd)
	}
	// finished iteration
	return
}

// processNetworkDevice runs routine to get network device status and update it in the DB.
func processNetworkDevice(ctx context.Context, dbClient *ent.Client, networkDevice *ent.NetworkDevice) {
	// iterating over endpoints and checking if any of them is alive.
	// it is enough to find one alive Endpoint.
	status := devicestatus.StatusSTATUS_DEVICE_DOWN
	aliveConnectionFound := false
	for _, ep := range networkDevice.Edges.Endpoints {
		// obtain connection based on the protocol.
		// retrieve device status.
		// if device status has failed to retrieve, proceed with other endpoint.
		// if device status was retrieved, break the loop and perform an update.
		// aliveConnectionFound = true
		// status = devicestatus.StatusSTATUS_DEVICE_UP
	}

	lastSeen := time.Now().String()
	if !aliveConnectionFound {
		// no alive endpoint was found, updating device status to down state.
		lastSeen = ""
	}
	// alive connection was found, status was fetched (and already fixed, updating device status
	_, _ = db.UpdateDeviceStatusByNetworkDeviceID(ctx, dbClient, networkDevice.ID, status, lastSeen)
	// error is already logged in in the internal function
	return
}
