package interact

import (
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
    "github.com/sirupsen/logrus"
)

type (
    Interact struct {
        Enabled   bool
        logWriter *logwriter.LogWriter
        log       logrus.FieldLogger
    }
    Interactish interface {
        NewProgress() *progress.Progress
        SpinWhile(message string, doFunc func())
    }
)

func New(enabled bool, logWriter *logwriter.LogWriter, log logrus.FieldLogger) *Interact {
    return &Interact{
        Enabled:   enabled,
        logWriter: logWriter,
        log:       log,
    }
}

func (i *Interact) NewProgress() *progress.Progress {
    if !i.Enabled {
        return nil
    }
    return progress.New(i.logWriter, i.log)
}

func (i *Interact) SpinWhile(message string, doFunc func()) {
    if !i.Enabled {
        doFunc()
        return
    }

    prog := progress.New(i.logWriter, i.log)
    spinner := prog.AddSpinner(message)

    doFunc()

    spinner.Incr()
    prog.Wait()
}

type Dummy struct{}

func (d *Dummy) NewProgress() *progress.Progress {
    return nil
}
