package progress

import (
    "github.com/vbauerster/mpb/v5"
)

type (
    Progress struct {
        uiProgress *mpb.Progress
        logWriter  LogWriter
        started    bool
    }
    LogWriter interface {
        Reset()
        DisableStdout()
    }
)

func New(logWriter LogWriter) *Progress {
    uiProgress := mpb.New(mpb.PopCompletedMode())

    return &Progress{
        uiProgress: uiProgress,
        logWriter:  logWriter,
    }
}

func (p *Progress) AddBar(barName string, total int) *Bar {
    if p.logWriter != nil {
        p.logWriter.DisableStdout()
    }
    return NewBar(p, barName, total)
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

func (p *Progress) AddSpinner(barName string) *Spinner {
    if p.logWriter != nil {
        p.logWriter.DisableStdout()
    }
    return NewSpinner(p, barName)
}
