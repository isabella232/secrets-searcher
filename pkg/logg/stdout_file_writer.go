package logg

import (
	"io"
	"os"
	"sync"
)

type StdoutFileWriter struct {
	logFilePath string
	logFile     *os.File
	extraWriter io.Writer
	mutex       *sync.Mutex
	stdEnabled  bool
}

func New(logFilePath string) (result *StdoutFileWriter) {
	return &StdoutFileWriter{
		logFilePath: logFilePath,
		mutex:       &sync.Mutex{},
		stdEnabled:  true,
	}
}

func (l *StdoutFileWriter) Reset() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.stdEnabled = true
}

func (l *StdoutFileWriter) DisableStdout() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.stdEnabled = false
}

func (l *StdoutFileWriter) Write(p []byte) (n int, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	file := l.file()

	if l.stdEnabled {
		return io.MultiWriter(file, os.Stdout).Write(p)
	}
	return file.Write(p)
}

func (l *StdoutFileWriter) file() (result *os.File) {
	if l.logFile != nil {
		return l.logFile
	}

	var err error
	result, err = os.OpenFile(l.logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("unable to open file: " + l.logFilePath)
	}

	l.logFile = result

	return
}
