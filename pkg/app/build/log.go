package build

import (
	"os"

	"github.com/pantheon-systems/secrets-searcher/pkg/dev"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var textFormatter = prefixed.TextFormatter{ForceFormatting: true}

func buildInitLog(logLevel string) (result *logg.LogrusLogg, err error) {
	var level logrus.Level
	if level, err = logrus.ParseLevel(logLevel); err != nil {
		err = errors.Wrapv(err, "invalid value for `log-level`: ", logLevel)
		return
	}

	// Usable before log file exists
	initLogger := newLogger(level)
	result = logg.NewLogrusLogg(initLogger)

	return
}

func buildAppLog(initLog *logg.LogrusLogg, logFile string) (result logg.Logg, err error) {

	// Log writer for app
	writer := logg.New(logFile)

	// Log for app
	appLogger := initLog.Spawn()
	appLogger.SetOutput(writer)
	result = logg.NewLogrusLogg(appLogger)

	if appLogger.Level() == logg.Trace {
		result = result.WithField("goroutine", dev.GoroutineID)
	}

	return
}

func newLogger(logLevel logrus.Level) (result *logg.LogrusLogg) {
	logrusLogger := logrus.New()
	logrusLogger.SetOutput(os.Stdout)
	logrusLogger.SetFormatter(&textFormatter)
	logrusLogger.SetLevel(logLevel)

	return logg.NewLogrusLogg(logrusLogger)
}
