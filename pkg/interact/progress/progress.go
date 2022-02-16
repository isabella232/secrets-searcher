package progress

import (
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/vbauerster/mpb/v5"
)

type (
	Progress struct {
		uiProgress *mpb.Progress
		logWriter  *logg.StdoutFileWriter
		started    bool
		bars       map[string]*Bar
		log        logg.Logg
	}
)

func New(logWriter *logg.StdoutFileWriter, log logg.Logg) *Progress {
	uiProgress := mpb.New(mpb.PopCompletedMode())

	return &Progress{
		uiProgress: uiProgress,
		logWriter:  logWriter,
		bars:       map[string]*Bar{},
		log:        log,
	}
}

func (p *Progress) AddBar(barName string, total int, appendMsgFormat, completedMsg string) (result *Bar) {
	if p.logWriter != nil {
		p.logWriter.DisableStdout()
	}

	var ok bool
	result, ok = p.bars[barName]
	if ok {
		return
	}

	p.bars[barName] = newBar(p, barName, total, appendMsgFormat, completedMsg, p.log)

	result = p.bars[barName]

	return
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

func (p *Progress) DisableStdout() {
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
