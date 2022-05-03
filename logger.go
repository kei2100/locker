package locker

import (
	"log"
	"os"
)

var DefaultLogger Logger = log.New(os.Stderr, "[mysql] ", log.Ldate|log.Ltime|log.Lshortfile)

type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}
