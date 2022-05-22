package locker

import (
	"log"
	"os"
)

// DefaultLogger is the default logger used by the Locker implementations
var DefaultLogger Logger = log.New(os.Stderr, "", log.LstdFlags)

// Logger interface
type Logger interface {
	// Printf calls Output to print to the standard logger.
	// Arguments are handled in the manner of fmt.Printf.
	Printf(format string, v ...interface{})

	// Println calls Output to print to the standard logger.
	// Arguments are handled in the manner of fmt.Println.
	Println(v ...interface{})
}
