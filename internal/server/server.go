// Package server implements main server logic
package server

import "time"

// StartServer starts the server
func StartServer(_ chan bool, _ chan bool) {
	time.Sleep(1 * time.Minute)
	panic("Implement me!")
}
