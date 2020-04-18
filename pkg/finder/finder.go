package finder

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirsean/go-pool"
    "github.com/sirupsen/logrus"
    "runtime"
)

type (
    Finder struct {
        Stats            *Stats
        chunkSize        int
        workerCount      int
        processors       []Processor
        fileChangeFilter *gitpkg.FileChangeFilter
        interact         interact.Interactish
        writer           *Writer
        payloadSet       *payloadSet
        db               *database.Database
        log              *logrus.Entry
    }
    Stats struct {
        CommitsSearchedCount int64
        SecretsFoundCount    int64
    }
)

func New(repoFilter *structures.Filter, commitFilter *gitpkg.CommitFilter, fileChangeFilter *gitpkg.FileChangeFilter, chunkSize, workerCount int, processors []Processor, whitelistSecretIDSet structures.Set, interact interact.Interactish, db *database.Database, log *logrus.Entry) *Finder {
    git := gitpkg.New(log)
    payloadSet := newPayloadSet(git, repoFilter, commitFilter, workerCount, chunkSize, db, log)
    writer := newWriter(whitelistSecretIDSet, db, log)

    return &Finder{
        Stats:            &Stats{},
        chunkSize:        chunkSize,
        workerCount:      workerCount,
        fileChangeFilter: fileChangeFilter,
        processors:       processors,
        interact:         interact,
        writer:           writer,
        payloadSet:       payloadSet,
        db:               db,
        log:              log,
    }
}

func (f *Finder) Search() (err error) {
    f.Stats = &Stats{}

    for _, tableName := range []string{database.CommitTable, database.FindingTable, database.SecretTable} {
        if f.db.TableExists(tableName) {
            err = errors.Errorv("one or more finder-specific tables already exist, cannot prepare findings", tableName)
            return
        }
    }

    var payloads []*payload
    var repoCounts map[string]int
    payloads, repoCounts, err = f.payloadSet.buildPayloads()
    if err != nil {
        return
    }

    var prog = f.interact.NewProgress()
    bars := f.progressBars(prog, repoCounts)

    numCPU := runtime.NumCPU()
    runtime.GOMAXPROCS(numCPU)

    out := make(chan *workerResult)

    pl := pool.NewPool(len(payloads), f.workerCount)
    pl.Start()

    go func() {
        for i, payload := range payloads {
            var bar *progress.Bar
            if bars != nil {
                bar = bars[payload.repo.Name]
            }

            workerName := fmt.Sprintf("%s-%d", payload.repo.Name, i)

            workerLog := f.log.
                WithField("worker", workerName).
                WithField("commits", payload.commitsLen)

            workerLog.Debugf("adding finder worker for repo")

            worker := NewWorker(
                workerName,
                payload,
                f.processors,
                f.fileChangeFilter,
                bar,
                out,
                f.db,
                workerLog,
            )

            pl.Add(worker)
        }

        pl.Close()
        if prog != nil {
            prog.Wait()
        }
        close(out)
    }()

    // Process findings from channel
    for dr := range out {
        log := f.log.WithField("repo", dr.RepoID)
        log.Debug("received finding from channel")

        if err = f.writer.persistResult(dr); err != nil {
            errors.ErrorLogForEntry(f.log, errors.WithMessage(err, "error processing commit"))
            continue
        }
    }

    // Stats
    for _, commitsLen := range repoCounts {
        f.Stats.CommitsSearchedCount += int64(commitsLen)
    }
    f.Stats.SecretsFoundCount = int64(f.writer.secretTracker.Len())

    f.log.Infof("completed search")

    return
}

func (f *Finder) progressBars(prog *progress.Progress, repoCounts map[string]int) (result map[string]*progress.Bar) {
    if prog == nil {
        return
    }

    result = make(map[string]*progress.Bar, len(repoCounts))
    for repoName, commitCount := range repoCounts {
        barName := fmt.Sprintf("%s", repoName)
        result[repoName] = prog.AddBar(barName, commitCount, "search of %s is complete")
    }
    return
}
