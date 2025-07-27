// Package checksum implements an abstraction (i.e., interface) for checksum generation check.
// It also implements a mock to enable smooth testing.
package checksum

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

const (
	componentNameMock = "mock-checksum-generator"
)

var zlogMock = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentNameMock).Logger()

// MockGenerator is an implementation of checksum generator interface for the sake of easy testing.
type MockGenerator struct{}

// NewMockGenerator function creates a new instance of the mock generator.
func NewMockGenerator() *MockGenerator {
	return &MockGenerator{}
}

// Generate function generates SHA256 checksum based on binary data provided at input of the function.
func (g *MockGenerator) Generate(data []byte) (string, error) {
	zlogMock.Info().Msg("Mock Generate checksum from provided data")
	// simple implementation for unit tests.
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
