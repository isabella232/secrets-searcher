package reporter

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/sirupsen/logrus"
    "os"
)

type Reporter struct {
    dir string
    db  *database.Database
    log *logrus.Logger
}

func New(dir string, db *database.Database, log *logrus.Logger) *Reporter {
    return &Reporter{
        dir: dir,
        db:  db,
        log: log,
    }
}

func (r *Reporter) PrepareReport() (err error) {
    if _, err := os.Stat(r.dir); os.IsNotExist(err) {
        if err := os.Mkdir(r.dir, 0644); err != nil {
            return errors.Wrapv(err, "unable to create report directory", r.dir)
        }
    }

    // TODO

    return
}
