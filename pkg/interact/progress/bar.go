package progress

import (
    "github.com/vbauerster/mpb/v5"
    "github.com/vbauerster/mpb/v5/decor"
    "time"
)

type Bar struct {
    barName    string
    progress   *Progress
    uiBar      *mpb.Bar
    unitPlural string
}

func NewBar(progress *Progress, barName string, total int) (result *Bar) {
    uiBar := progress.uiProgress.AddBar(int64(total),
        mpb.BarNoPop(),
        mpb.BarRemoveOnComplete(),
        mpb.PrependDecorators(
            decor.Name(barName, decor.WC{W: 30, C: decor.DidentRight}),
        ),
        mpb.AppendDecorators(
            decor.OnComplete(
                decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
            ),
        ),
    )

    result = &Bar{barName: barName, progress: progress, uiBar: uiBar}

    return
}

func (b *Bar) DecoratorEwmaUpdate(dur time.Duration) {
    b.uiBar.DecoratorEwmaUpdate(dur)
}

func (b *Bar) Incr() {
    b.uiBar.Increment()
}
