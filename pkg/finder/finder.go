package finder

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/stats"
    "github.com/sirupsen/logrus"
    "sync"
    "time"
)

type Finder struct {
    *Writer
    *SearchBuilder
    workerCount int
    interact    interact.Interactish
    log         logrus.FieldLogger
}

func New(writer *Writer,
    searchBuilder *SearchBuilder,
    workerCount int,
    interact interact.Interactish,
    log logrus.FieldLogger,
) *Finder {
    return &Finder{
        Writer:        writer,
        SearchBuilder: searchBuilder,
        workerCount:   workerCount,
        interact:      interact,
        log:           log,
    }
}

func (f *Finder) Search() (err error) {
    stats.SearchStartTime = time.Now()

    f.log.Info("finding secrets ... ")

    if err = f.Writer.prepareFilesystem(); err != nil {
        err = errors.WithMessage(err, "unable to prepare filesystem")
        return
    }

    // Results queue
    out := make(chan *searchResult)

    // Progress bar
    prog := f.interact.NewProgress()

    // Create job queue
    var jobs []Search
    var totalCommitCount int
    jobs, totalCommitCount, err = f.SearchBuilder.getJobs(prog, out)
    if err != nil {
        err = errors.WithMessage(err, "unable to build jobs")
        return
    }
    jobLen := len(jobs)

    // Wait group ends when all jobs are done
    var jobsWG sync.WaitGroup
    jobsWG.Add(jobLen)

    // Wait group ends when all processing is done
    var processWG sync.WaitGroup
    processWG.Add(1)

    // Job queue
    jobQueue := make(chan Search, jobLen)

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
    for _, job := range jobs {
        jobQueue <- job
    }
    close(jobQueue)

    // Process findings from results channel
    go func() {
        defer processWG.Done()

        for result := range out {

            //log := f.log.WithField("repo", result.RepoID)

            if err = f.Writer.persistResult(result); err != nil {
                errors.ErrLog(f.log, errors.WithMessage(err, "error processing commit"))
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
    stats.CommitsSearchedCount += int64(totalCommitCount)
    stats.SecretsFoundCount = int64(f.Writer.secretTracker.Len())

    f.log.Info("completed search")

    stats.SearchEndTime = time.Now()

    return
}
