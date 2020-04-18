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
        log       *logrus.Entry
    }
    Dummy       struct{}
    Interactish interface {
        NewProgress() *progress.Progress
    }
)

func New(enabled bool, logWriter *logwriter.LogWriter, log *logrus.Entry) *Interact {
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

func (d *Dummy) NewProgress() *progress.Progress {
    return nil
}
