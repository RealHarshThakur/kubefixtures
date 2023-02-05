package pkg

import (
	"io"

	"github.com/sirupsen/logrus"
)

// SetupLogging sets up the logging for the router daemon
func SetupLogging(o io.Writer) *logrus.Logger {
	// Logging create logging object
	log := logrus.New()
	log.SetOutput(o)
	log.SetLevel(logrus.DebugLevel)

	return log
}
