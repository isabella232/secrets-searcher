package finder

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/stats"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "runtime"
    "sync"
    "time"
)

type (
    Finder struct {
        chunkSize           int
        workerCount         int
        commitSearchTimeout time.Duration
        processors          []Processor
        fileChangeFilter    *gitpkg.FileChangeFilter
        interact            interact.Interactish
        writer              *Writer
        payloadSet          *searchBuilder
        db                  *database.Database
        log               logrus.FieldLogger
    }
    job interface {
        Perform()
    }
)

func New(repoFilter *structures.Filter, commitFilter *gitpkg.CommitFilter, fileChangeFilter *gitpkg.FileChangeFilter, chunkSize, workerCount int, commitSearchTimeout time.Duration, processors []Processor, whitelistSecretIDSet structures.Set, interact interact.Interactish, db *database.Database, log logrus.FieldLogger) *Finder {
    git := gitpkg.New(log)
    payloadSet := newPayloadSet(git, repoFilter, commitFilter, workerCount, chunkSize, interact, db, log)
    writer := newWriter(whitelistSecretIDSet, db, log)

    return &Finder{
        chunkSize:           chunkSize,
        workerCount:         workerCount,
        fileChangeFilter:    fileChangeFilter,
        processors:          processors,
        interact:            interact,
        writer:              writer,
        payloadSet:          payloadSet,
        commitSearchTimeout: commitSearchTimeout,
        db:                  db,
        log:                 log,
    }
}

func (f *Finder) Search() (err error) {
    stats.SearchStartTime = time.Now()

    f.log.Info("finding secrets ... ")

    for _, tableName := range []string{database.CommitTable, database.FindingTable, database.SecretTable} {
        if f.db.TableExists(tableName) {
            err = errors.Errorv("one or more finder-specific tables already exist, cannot prepare findings", tableName)
            return
        }
    }

    var payloads []*searchParameters
    var repoCounts map[string]int
    var commitCount int
    payloads, repoCounts, commitCount, err = f.payloadSet.buildPayloads()
    if err != nil {
        err = errors.WithMessage(err, "unable to build search payloads")
        return
    }
    payloadsLen := len(payloads)

    var prog = f.interact.NewProgress()
    bars := f.progressBars(prog, repoCounts)

    numCPU := runtime.NumCPU()
    runtime.GOMAXPROCS(numCPU)

    jobQueue := make(chan job, payloadsLen)
    out := make(chan *searchResult)

    // Wait group ends when all jobs are done
    var jobsWG sync.WaitGroup
    jobsWG.Add(payloadsLen)

    // Wait group ends when all processing is done
    var processWG sync.WaitGroup
    processWG.Add(1)

    // Set up workers
    for i := 0; i < f.workerCount; i++ {
        go func() {
            // Pull from job queue
            for job := range jobQueue {
                job.Perform()
                jobsWG.Done()
            }
        }()
    }

    // Add jobs to queue
    for i, payload := range payloads {
        var bar *progress.Bar
        if bars != nil {
            bar = bars[payload.repo.Name]
        }

        workerName := fmt.Sprintf("%s-%d", payload.repo.Name, i)

        workerLog := f.log.
            WithField("worker", workerName).
            WithField("commitsLen", payload.commitsLen)

        workerLog.Debugf("adding finder worker for repo")

        jobQueue <- NewSearch(out, workerName, payload, f.processors, f.fileChangeFilter, f.commitSearchTimeout, bar, f.db, workerLog)
    }

    // Close job queue to further writes
    close(jobQueue)

    // Process findings from channel
    go func() {
        defer processWG.Done()

        for result := range out {

            //log := f.log.WithField("repo", result.RepoID)

            if err = f.writer.persistResult(result); err != nil {
                errors.ErrorLogger(f.log, errors.WithMessage(err, "error processing commit"))
                continue
            }
        }
    }()

    // Wait for all jobs to finish
    jobsWG.Wait()
    // Wait for progress bar to end
    if prog != nil {
        prog.Wait()
    }
    // Close result queue
    close(out)
    // Wait for processing loop to stop
    processWG.Wait()

    // Stats
    stats.CommitsSearchedCount += int64(commitCount)
    stats.SecretsFoundCount = int64(f.writer.secretTracker.Len())

    f.log.Infof("completed search")

    stats.SearchEndTime = time.Now()

    return
}

func (f *Finder) progressBars(prog *progress.Progress, repoCounts map[string]int) (result map[string]*progress.Bar) {
    if prog == nil {
        return
    }

    result = make(map[string]*progress.Bar, len(repoCounts))
    for repoName, commitCount := range repoCounts {
        barName := fmt.Sprintf("%s", repoName)
        result[repoName] = prog.AddBar(barName, commitCount, "searched %d of %d commits", "search of %s is complete")
    }
    return
}
