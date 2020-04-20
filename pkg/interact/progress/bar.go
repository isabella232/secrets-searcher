package progress

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/vbauerster/mpb/v5"
    "github.com/vbauerster/mpb/v5/decor"
    "io"
    "strings"
    "sync"
)

type Bar struct {
    barName          string
    progress         *Progress
    uiBar            *mpb.Bar
    unitPlural       string
    total            int
    runningTotal     int
    appendMsgFormat  string
    completedMessage string
    mutex            *sync.Mutex
    log              logrus.FieldLogger

    // FIXME Not the place for this
    SecretTracker structures.Set
}

func newBar(progress *Progress, barName string, total int, appendMsgFormat, completedMessage string, log logrus.FieldLogger) (result *Bar) {
    result = &Bar{
        barName:          barName,
        progress:         progress,
        total:            total,
        runningTotal:     total,
        appendMsgFormat:  appendMsgFormat,
        completedMessage: completedMessage,
        mutex:            &sync.Mutex{},
        log:              log,
        SecretTracker:    structures.NewSet(nil),
    }
    return
}

func (b *Bar) Start() {
    b.mutex.Lock()
    defer b.mutex.Unlock()

    if b.uiBar != nil {
        return
    }

    b.SecretTracker = structures.NewSet(nil)

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

    b.uiBar.Increment()

    b.runningTotal -= 1

    if b.runningTotal == 0 {

        // FIXME Not the place for this
        secretsFound := b.SecretTracker.Len()
        message := fmt.Sprintf("%d commits searched", b.total)
        if secretsFound > 0 {
            message += fmt.Sprintf(", %d SECRETS FOUND", secretsFound)
        }

        b.Finished(message)
    }
}

func (b *Bar) Finished(perensMessage string) {
    if b.completedMessage != "" {
        b.progress.Add(0, mpb.BarFillerFunc(func(writer io.Writer, width int, st *decor.Statistics) {
            message := fmt.Sprintf(b.completedMessage, b.barName)
            if strings.Contains(message, "%!(EXTRA") {
                message = b.completedMessage
            }
            fmt.Fprintf(writer, "- %s", message)
            if perensMessage != "" {
                fmt.Fprintf(writer, " (%s)", perensMessage)
            }
        })).SetTotal(0, true)
    }
}

func (b *Bar) BustThrough(fnc func()) {
    b.progress.BustThrough(fnc)
}

// FIXME Not the place for this
func (b *Bar) AddSecret(secretIdent string) {
    b.SecretTracker.Add(secretIdent)
}
