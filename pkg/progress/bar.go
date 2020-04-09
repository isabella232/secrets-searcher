package progress

import (
    "fmt"
    "github.com/gosuri/uiprogress"
    "github.com/gosuri/uiprogress/util/strutil"
    "time"
)

type Bar struct {
    barName    string
    progress   *Progress
    uiBar      *uiprogress.Bar
    unitPlural string
}

func NewBar(progress *Progress, barName, unitPlural string) (result *Bar) {
    uiBar := progress.uiProgress.AddBar(1)
    uiBar = uiBar.PrependElapsed()
    uiBar.AppendFunc(func(b *uiprogress.Bar) string {
        return fmt.Sprintf("%d/%d %s", b.Current(), b.Total, unitPlural)
    })
    uiBar.PrependFunc(func(b *uiprogress.Bar) string { return strutil.Resize(barName, 22) })

    result = &Bar{barName: barName, progress: progress, uiBar: uiBar}

    return
}

func (b *Bar) Start(total int) {
    b.progress.Start()
    b.uiBar.Total = total
    b.uiBar.TimeStarted = time.Now()
}

func (b *Bar) End(total int) {
    _, err := fmt.Fprintf(b.progress.Bypass(), "%s finished", b.barName)
    if err != nil {
        panic(err)
    }

    b.uiBar.Total = total
    b.uiBar.TimeStarted = time.Now()
}

func (b *Bar) Incr() bool {
    return b.uiBar.Incr()
}
