package logwriter

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "io"
    "os"
    "sync"
)

type LogWriter struct {
    logFilePath string
    logFile     *os.File
    extraWriter io.Writer
    mutex       *sync.Mutex
    stdEnabled  bool
}

func New(logFilePath string) (result *LogWriter) {
    return &LogWriter{
        logFilePath: logFilePath,
        mutex:       &sync.Mutex{},
        stdEnabled:  true,
    }
}

func (l *LogWriter) Reset() {
    l.stdEnabled = true
}

func (l *LogWriter) DisableStdout() {
    l.stdEnabled = false
}

func (l *LogWriter) Write(p []byte) (n int, err error) {
    l.mutex.Lock()
    defer l.mutex.Unlock()

    file := l.file()

    if l.stdEnabled {
        return io.MultiWriter(file, os.Stdout).Write(p)
    }
    return file.Write(p)
}

func (l *LogWriter) file() (result *os.File) {
    if l.logFile != nil {
        return l.logFile
    }

    var err error
    result, err = os.OpenFile(l.logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        panic(errors.WithMessagev(err, "unable to open file", l.logFilePath).Error())
    }

    l.logFile = result

    return
}
