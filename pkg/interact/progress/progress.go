package progress

import (
    "github.com/sirupsen/logrus"
    "github.com/vbauerster/mpb/v5"
)

type (
    Progress struct {
        uiProgress *mpb.Progress
        logWriter  LogWriter
        started    bool
        log        logrus.FieldLogger
    }
    LogWriter interface {
        Reset()
        DisableStdout()
    }
)

func New(logWriter LogWriter, log logrus.FieldLogger) *Progress {
    uiProgress := mpb.New(mpb.PopCompletedMode())

    return &Progress{
        uiProgress: uiProgress,
        logWriter:  logWriter,
        log:        log,
    }
}

func (p *Progress) AddBar(barName string, total int, appendMsgFormat, completedMsg string) *Bar {
    if p.logWriter != nil {
        p.logWriter.DisableStdout()
    }
    return newBar(p, barName, total, appendMsgFormat, completedMsg, p.log)
}

func (p *Progress) AddSpinner(barName string) *Spinner {
    if p.logWriter != nil {
        p.logWriter.DisableStdout()
    }
    return newSpinner(p, barName)
}

func (p *Progress) Add(total int64, filler mpb.BarFiller, options ...mpb.BarOption) *mpb.Bar {
    return p.uiProgress.Add(total, filler, options...)
}

func (p *Progress) BustThrough(fnc func()) {
    if p.logWriter != nil {
        p.logWriter.Reset()
    }
    fnc()
    if p.logWriter != nil {
        p.logWriter.DisableStdout()
    }
}

func (p *Progress) ResetLogwriter() {
    if p.logWriter != nil {
        p.logWriter.Reset()
    }
}

func (p *Progress) Wait() {
    p.uiProgress.Wait()
    if p.logWriter != nil {
        p.logWriter.Reset()
    }
}
