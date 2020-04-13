package interact

import (
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
)

type (
    Interact struct {
        Enabled   bool
        logWriter *logwriter.LogWriter
    }
    Dummy       struct{}
    Interactish interface {
        NewProgress() *progress.Progress
    }
)

func New(enabled bool, logWriter *logwriter.LogWriter) *Interact {
    return &Interact{
        Enabled:   enabled,
        logWriter: logWriter,
    }
}

func (f *Interact) NewProgress() *progress.Progress {
    if !f.Enabled {
        return nil
    }
    return progress.New(f.logWriter)
}

func (d *Dummy) NewProgress() *progress.Progress {
    return nil
}
