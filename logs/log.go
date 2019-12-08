package logs

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)
}

func Info(format string, args ...interface{}) {
	log.WithFields(log.Fields{}).Info(fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	log.Warn(fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	log.Warn(fmt.Sprintf(format, args...))
}
