package cmd

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
    "github.com/sirupsen/logrus"
    prefixed "github.com/x-cray/logrus-prefixed-formatter"
    "os"
    "path/filepath"
)

var (
    logger    *logrus.Logger
    log       *logrus.Entry
    logWriter *logwriter.LogWriter
)

// Initialize logger (pre-config)
func initLogging() {
    logger = logrus.New()
    logger.SetOutput(os.Stdout)
    logger.SetFormatter(&prefixed.TextFormatter{ForceFormatting: true})
}

// Initialize logger (post-config)
func configureLogging() (err error) {
    // Log level
    var logLevel logrus.Level
    logLevel, err = logrus.ParseLevel(cfg.LogLevel)
    if err != nil {
        err = errors.Wrapv(err, "invalid value for log-level: ", cfg.LogLevel)
        return
    }
    logger.SetLevel(logLevel)

    // Log file setup
    logFilePath := filepath.Join(cfg.OutputDir, "run.log")
    if _, statErr := os.Stat(logFilePath); os.IsNotExist(statErr) {
        // Log file does not exist so create one
        // FIXME This step shouldn't be necessary
        var empty *os.File
        empty, err = os.Create(logFilePath)
        if err != nil {
            err = errors.Wrapv(err, "unable to create log file", logFilePath)
            return
        }
        empty.Close()
    } else if statErr == nil {
        // Log file exists so truncate it
        // If you delete it, `tail -f` needs to be restarted
        if err = os.Truncate(logFilePath, 0); err != nil {
            err = errors.Wrapv(err, "unable to truncate log file", logFilePath)
            return
        }
    }

    // Log writer
    logWriter, err = logwriter.New(logFilePath)
    if err != nil {
        err = errors.Wrap(err, "unable to build log writer")
        return
    }
    logger.SetOutput(logWriter)

    // Log entry
    log = logger.WithField("prefix", "root")

    return
}
