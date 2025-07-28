// Package manager implements SB control loop, that fetches data from the devices and stores it to the DB.
package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/pkg/checksum"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/eroshiva/trade-show-poc/pkg/connectors"
	"github.com/rs/zerolog"
)

const (
	component     = "component"
	componentName = "control-loop"

	defaultControlLoopPerioud = 30 * time.Second
	// EnvControlLoopPeriod defines a control loop period in seconds.
	EnvControlLoopPeriod = "CONTROL_LOOP_PERIOD" // in seconds.
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// Manager structure holds the dependencies for main control loop.
type Manager struct {
	dbClient          *ent.Client
	checksumGenerator checksum.Generator
	closeChan         chan bool
}

// NewManager function creates Manager structure.
func NewManager(dbClient *ent.Client, checksumGen checksum.Generator) *Manager {
	return &Manager{
		dbClient:          dbClient,
		checksumGenerator: checksumGen,
		closeChan:         make(chan bool),
	}
}

// StopManager sends signal to stop main control loop.
func (m *Manager) StopManager() {
	close(m.closeChan)
	zlog.Info().Msg("Stopping manager...")
}

// StartManager function starts main control loop that periodically fetches data from the network devices.
func (m *Manager) StartManager() {
	zlog.Info().Msg("Starting manager...")
	controlLoopTick := defaultControlLoopPerioud
	// read env variable, where Control Loop Period is specified
	envControlLoopTick := os.Getenv(EnvControlLoopPeriod)
	if envControlLoopTick == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default address: %s",
			EnvControlLoopPeriod, defaultControlLoopPerioud)
	} else {
		// control loop tick is specified
		duration, err := strconv.Atoi(envControlLoopTick)
		if err != nil {
			zlog.Fatal().Err(err).Msgf("Failed to convert \"%s\"variable to number", EnvControlLoopPeriod)
		}
		controlLoopTick = time.Duration(duration) * time.Second
	}

	// executing control loop
	go m.controlLoop(controlLoopTick)
}

func (m *Manager) controlLoop(controlLoopTick time.Duration) {
	zlog.Info().Msgf("Starting periodical (%s seconds) execution of main control loop", controlLoopTick)
	// creating ticker
	ticker := time.NewTicker(controlLoopTick)

	// performing control loop routine at the very beginning
	m.PerformControlLoopRoutine(controlLoopTick)

	// starting infinite control loop
	for {
		select {
		case <-ticker.C:
			// performing control loop routine
			m.PerformControlLoopRoutine(controlLoopTick)
		case <-m.closeChan:
			// shutting down this routine
			zlog.Debug().Msgf("Stopping main control loop")
			return
		}
	}
}

// PerformControlLoopRoutine runs main control loop routine, i.e., fetches all devices and updates theirs status.
func (m *Manager) PerformControlLoopRoutine(controlLoopTick time.Duration) {
	zlog.Debug().Msgf("Executing main control loop routine")
	ctx, cancel := context.WithTimeout(context.Background(), controlLoopTick)
	defer cancel()

	// fetching all devices from the DB
	ndList, err := db.ListNetworkDevices(ctx, m.dbClient)
	if err != nil {
		// error is already logged in in the inner function
		return
	}

	// checking if we retrieved more than 0 network devices
	if len(ndList) == 0 {
		// no devices were retrieved, nothing to do
		zlog.Warn().Msgf("No network devices found in the DB")
		return
	}

	// updating network devices concurrently
	wg := &sync.WaitGroup{}
	for _, nd := range ndList {
		wg.Add(1)
		go func() {
			m.processNetworkDevice(ctx, nd)
			wg.Done()
		}()
	}
	// waiting for all goroutines to finish their iteration (otherwise context gets cancelled before network device processing will be done)
	wg.Wait()
}

// processNetworkDevice runs routine to get network device status, SW, FW, and HW versions from the device and update them in the DB.
func (m *Manager) processNetworkDevice(ctx context.Context, networkDevice *ent.NetworkDevice) {
	zlog.Debug().Msgf("Processing network device (%s)", networkDevice.ID)
	// iterating over endpoints and checking if any of them is alive.
	// it is enough to find one alive Endpoint.
	hwV := ""
	swV := &ent.Version{}
	fwV := &ent.Version{}
	status := devicestatus.StatusSTATUS_DEVICE_DOWN
	aliveConnectionFound := false
	for _, ep := range networkDevice.Edges.Endpoints {
		// obtain connection based on the protocol.
		connector, err := connectors.NewConnector(ep)
		if err != nil {
			// we've hit an unsupported protocol case, skipping the rest of the iteration
			continue
		}
		// retrieve device status.
		s, err := connector.GetStatus(ctx)
		if err != nil {
			// failed to retrieve status, proceeding with other the endpoint.
			// assuming that error is already logged in within the function.
			continue
		}
		// device status was retrieved, break the loop and perform an update.
		aliveConnectionFound = true
		status = s // assuming that any live connection is different from down

		// checking HW version
		hwV, _ = connector.GetHWVersion(ctx)
		// in case of error, assuming that error is already logged in within the function.
		// even if this call fails, continue to retrieve the other versions.
		// DB client will do sanity check and skip default values.

		swV, _ = connector.GetSWVersion(ctx)
		// in case of error, assuming that error is already logged in within the function.
		// even if this call fails, continue to retrieve the other version.
		// DB client will do sanity check and skip default values.

		fwV, _ = connector.GetFWVersion(ctx)
		// in case of error, assuming that error is already logged in within the function.
		// even if this call fails, continue to retrieve the other versions.
		// DB client will do sanity check and skip default values.

		// no need in further sniffing of other endpoints
		break
	}

	lastSeen := time.Now().String()
	if !aliveConnectionFound {
		// no alive endpoint was found, updating device status to down state and resetting timestamp.
		lastSeen = ""
	}
	// alive connection was found and status was fetched (and already fixed, updating device status
	_, _ = db.UpdateDeviceStatusByNetworkDeviceID(ctx, m.dbClient, networkDevice.ID, status, lastSeen)
	// error is already logged in in the internal function

	// conducting checksum verifications
	err := m.verifyChecksum(swV)
	if err != nil {
		// resetting SW version, do not updating it in the DB
		swV = &ent.Version{}
	}
	err = m.verifyChecksum(fwV)
	if err != nil {
		// resetting FW version, do not updating it in the DB
		fwV = &ent.Version{}
	}

	// updating HW, SW, and FW versions
	_, _ = db.UpdateNetworkDeviceVersions(ctx, m.dbClient, networkDevice.ID, hwV, swV, fwV)
	// error is already logged in in the internal function
}

// verifyChecksum runs checksum verification against checksum generator binary.
func (m *Manager) verifyChecksum(version *ent.Version) error {
	checksumGen, err := m.checksumGenerator.Generate([]byte(version.Version))
	if err != nil {
		// failed generating checksum, assuming that error is logged in internally in function
		return err
	}
	// comparing checksums, they should be identical
	if checksumGen != version.Checksum {
		// checksums are different, reporting error
		newErr := fmt.Errorf("checksum verification failed - invalid checksum")
		zlog.Error().Err(newErr).Msgf("Checksum verification failed")
		return newErr
	}
	// all good
	return nil
}
