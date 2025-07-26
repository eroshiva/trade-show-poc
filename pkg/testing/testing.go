// Package testing contains helper functions to run unit tests within this repository.
package testing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	_ "github.com/lib/pq" // SQL driver, necessary for DB interaction
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const (
	component     = "component"
	componentName = "testing"
	// DefaultTestTimeout defines default test timeout for a testing package
	DefaultTestTimeout = time.Second * 1
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// Setup function sets up testing environment by uploading schema to the DB.
func Setup() (*ent.Client, error) {
	zlog.Info().Msgf("Opening connection to PostreSQL...")
	client, err := ent.Open(dialect.Postgres, "host=localhost port=5432 user=admin dbname=postgres password=pass sslmode=disable")
	if err != nil {
		zlog.Error().Err(err).Msgf("failed opening connection to postgres")
		return nil, err
	}

	zlog.Info().Msgf("Migrating database schema...")
	// Run the auto migration tool.
	if err = client.Schema.Create(context.Background()); err != nil {
		zlog.Error().Err(err).Msgf("failed creating schema resources")
		// gracefully closing client
		newErr := client.Close()
		if newErr != nil {
			zlog.Error().Err(newErr).Msgf("failed closing connection to postgres")
			return nil, newErr
		}
		return nil, err
	}

	return client, nil
}

// GracefullyCloseEntClient gracefully closes connection with the DB.
func GracefullyCloseEntClient(client *ent.Client) error {
	err := client.Close()
	if err != nil {
		zlog.Error().Err(err).Msgf("failed closing connection to postgres")
		return err
	}
	return nil
}

// AssertEqualVersion runs assertions on the fields of Version resource
func AssertEqualVersion(t *testing.T, expected, actual *ent.Version) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Version, actual.Version)
	assert.Equal(t, expected.Checksum, actual.Checksum)
}

// AssertDeviceStatus runs assertions on the fields of Version resource
func AssertDeviceStatus(t *testing.T, expected, actual *ent.DeviceStatus) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Status.String(), actual.Status.String())
	assert.Equal(t, expected.LastSeen, actual.LastSeen)
}
