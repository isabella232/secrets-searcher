package progress

import (
    "github.com/gosuri/uiprogress"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
    "io"
)

type Progress struct {
    uiProgress *uiprogress.Progress
    logWriter  *logwriter.LogWriter
    oldLogOut  io.Writer
    started    bool
}

func New(logWriter *logwriter.LogWriter) *Progress {
    uiProgress := uiprogress.New()
    uiProgress.Width = 80

    return &Progress{
        uiProgress: uiProgress,
        logWriter:  logWriter,
    }
}

func (p *Progress) Start() {
    if p.started {
        return
    }
    p.uiProgress.Start()
    if p.logWriter != nil {
        p.logWriter.DisableStdout()
    }

    p.started = true
}

func (p *Progress) Stop() {
    p.uiProgress.Stop()
    if p.logWriter != nil {
        p.logWriter.Reset()
    }
}

func (p *Progress) AddBar(barName, unitPlural string) (result *Bar) {
    return NewBar(p, barName, unitPlural)
}

func (p *Progress) Bypass() io.Writer {
    return p.uiProgress.Bypass()
}
