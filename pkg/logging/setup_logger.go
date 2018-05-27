package logging

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func Setup(config Config) {
	switch config.Level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	switch config.Format {
	case "json":
		log.SetFormatter(NewLogrusJSONFormatter())

	default:
		log.SetFormatter(&log.TextFormatter{})

	}

	log.SetOutput(os.Stderr)
}
