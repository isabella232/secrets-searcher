package progress

import (
    "github.com/vbauerster/mpb/v5"
    "github.com/vbauerster/mpb/v5/decor"
)

type Spinner struct {
    barName    string
    progress   *Progress
    uiBar      *mpb.Bar
    unitPlural string
}

func NewSpinner(progress *Progress, barName string) (result *Spinner) {
    uiBar := progress.uiProgress.AddSpinner(int64(1), mpb.SpinnerOnMiddle,
        mpb.BarNoPop(),
        mpb.BarRemoveOnComplete(),
        mpb.PrependDecorators(
            decor.Name(barName, decor.WC{W: 30, C: decor.DidentRight}),
        ),
    )

    result = &Spinner{barName: barName, progress: progress, uiBar: uiBar}

    return
}

func (b *Spinner) Incr() {
    b.uiBar.Increment()
}
