package cmd

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/sirupsen/logrus"
    prefixed "github.com/x-cray/logrus-prefixed-formatter"
    "os"
)

// Initialize logger (pre-config)
func initLogging() (logger *logrus.Logger) {
    logger = logrus.New()
    logger.SetOutput(os.Stdout)
    logger.SetFormatter(&prefixed.TextFormatter{ForceFormatting: true})
    return
}

// Initialize logger (post-config)
func configureLogging(logger *logrus.Logger, cfg config) (log *logrus.Entry, err error) {
    // Log level
    var logLevel logrus.Level
    logLevel, err = logrus.ParseLevel(cfg.LogLevel)
    if err != nil {
        err = errors.Wrapv(err, "invalid value for log-level: ", cfg.LogLevel)
        return
    }
    logger.SetLevel(logLevel)

    log = logrus.NewEntry(logger)

    return
}
