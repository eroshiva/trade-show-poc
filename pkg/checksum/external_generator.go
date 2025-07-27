// Package checksum implements an abstraction (i.e., interface) for checksum generation check.
// It also implements a mock to enable smooth testing.
package checksum

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	component     = "component"
	componentName = "checksum-generator"
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// ExternalGenerator implements checksum generator interface and defines interaction with production
// checksum generator (provided binary).
type ExternalGenerator struct {
	BinaryPath string
}

// NewExternalGenerator creates a new instance of the external generator.
func NewExternalGenerator(binaryPath string) (*ExternalGenerator, error) {
	if _, err := exec.LookPath(binaryPath); err != nil {
		zlog.Error().Err(err).Msg("external generator was not found")
		return nil, fmt.Errorf("checksum binary not found at path %q: %w", binaryPath, err)
	}
	return &ExternalGenerator{BinaryPath: binaryPath}, nil
}

// Generate function shows theoretical implementation on command line execution on top of a binary installed into system.
// This is NOT tested, just a sample implementation.
func (g *ExternalGenerator) Generate(data []byte) (string, error) {
	// creating the command and passing the context to allow for timeouts/cancellation.
	cmd := exec.CommandContext(context.Background(), g.BinaryPath)

	// obtaining a pipe to the command's standard input.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		newErr := fmt.Errorf("failed to get stdin pipe: %w", err)
		zlog.Error().Err(newErr).Msg("Failed to get stdin pipe")
		return "", newErr
	}

	// obtaining a pipe to the command's standard output.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		newErr := fmt.Errorf("failed to get stdout pipe: %w", err)
		zlog.Error().Err(newErr).Msg("Failed to get stdout pipe")
		return "", newErr
	}

	// creating a buffer to capture any error output from the command.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// firing off the command
	if err := cmd.Start(); err != nil {
		newErr := fmt.Errorf("failed to start checksum command: %w", err)
		zlog.Error().Err(newErr).Msg("Failed to start checksum command")
		return "", newErr
	}

	// writing input data to the command's stdin in a separate goroutine
	// and then closing the pipe to signal that we are done writing.
	go func() {
		defer func() {
			err := stdout.Close()
			if err != nil {
				zlog.Error().Err(err).Msg("Failed to close stdin pipe")
			}
		}()
		numWrtBytes, err := io.Copy(stdin, bytes.NewReader(data))
		if err != nil {
			zlog.Error().Err(err).Msg("Failed to copy data to stdin pipe")
		}
		zlog.Debug().Msgf("%d bytes were copied from stdin pipe", numWrtBytes)
	}()

	// reading output from the command's stdout.
	output, err := io.ReadAll(stdout)
	if err != nil {
		newErr := fmt.Errorf("failed to read checksum output: %w", err)
		zlog.Error().Err(newErr).Msg("Failed to read checksum output")
		return "", newErr
	}

	// waiting for the command to exit and release its resources.
	if err := cmd.Wait(); err != nil {
		newErr := fmt.Errorf("checksum command failed: %w | stderr: %s", err, stderr.String())
		zlog.Error().Err(newErr).Msg("Checksum command failed")
		return "", newErr
	}

	// cleaning up the output and returning it.
	return strings.TrimSpace(string(output)), nil
}
