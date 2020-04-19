package logwriter

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "io"
    "os"
    "sync"
)

type LogWriter struct {
    writer      io.Writer
    logFile     *os.File
    extraWriter io.Writer
    mutex       *sync.Mutex
}

func New(logFilePath string) (result *LogWriter, err error) {
    var logFile *os.File
    logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        err = errors.WithMessagev(err, "unable to open file", logFilePath)
        return
    }

    result = &LogWriter{
        logFile: logFile,
        mutex:   &sync.Mutex{},
    }

    result.Reset()

    return
}

func (l *LogWriter) Reset() {
    l.mutex.Lock()
    defer l.mutex.Unlock()

    l.writer = io.MultiWriter(l.logFile, os.Stdout)
}

func (l *LogWriter) DisableStdout() {
    l.mutex.Lock()
    defer l.mutex.Unlock()

    l.writer = l.logFile
}

func (l *LogWriter) Write(p []byte) (n int, err error) {
    l.mutex.Lock()
    defer l.mutex.Unlock()

    return l.writer.Write(p)
}
