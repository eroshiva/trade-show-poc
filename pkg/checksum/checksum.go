// Package checksum implements an abstraction (i.e., interface) for checksum generation check.
// It also implements a mock to enable smooth testing.
package checksum

// Generator interface defines main functions for checksum generator.
type Generator interface {
	Generate([]byte) (string, error)
}
