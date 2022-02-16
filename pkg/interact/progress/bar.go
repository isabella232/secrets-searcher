package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

type Bar struct {
	barName          string
	progress         *Progress
	uiBar            *mpb.Bar
	unitPlural       string
	total            int
	appendMsgFormat  string
	completedMessage string
	mutex            *sync.Mutex
	log              logg.Logg
}

func newBar(progress *Progress, barName string, total int, appendMsgFormat, completedMessage string, log logg.Logg) *Bar {
	return &Bar{
		barName:          barName,
		progress:         progress,
		total:            total,
		appendMsgFormat:  appendMsgFormat,
		completedMessage: completedMessage,
		mutex:            &sync.Mutex{},
		log:              log,
	}
}

func (b *Bar) Start() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.uiBar != nil {
		return
	}

	b.uiBar = b.progress.uiProgress.AddBar(int64(b.total),
		mpb.BarNoPop(),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(b.barName, decor.WC{W: 50, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.CountersNoUnit(b.appendMsgFormat),
		),
	)
}

func (b *Bar) Incr() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.progress.DisableStdout()

	b.uiBar.Increment()
}

func (b *Bar) Finished(perensMessage string) {
	if b.completedMessage != "" {
		b.progress.Add(0, mpb.BarFillerFunc(func(writer io.Writer, width int, st *decor.Statistics) {
			message := fmt.Sprintf(b.completedMessage, b.barName)
			if strings.Contains(message, "%!(EXTRA") {
				message = b.completedMessage
			}
			_, _ = fmt.Fprintf(writer, "- %s", message)
			if perensMessage != "" {
				_, _ = fmt.Fprintf(writer, " (%s)", perensMessage)
			}
		})).SetTotal(0, true)
	}
}

func (b *Bar) BustThrough(fnc func()) {
	b.progress.BustThrough(fnc)
}
