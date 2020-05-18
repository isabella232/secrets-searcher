package search

import (
	"time"

	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/stats"
)

type Search struct {
	jobBuilder *JobBuilder
	jobRunner  *JobRunner
	interact   *interactpkg.Interact
	db         *database.Database
	log        logg.Logg
}

func New(jobBuilder *JobBuilder, jobRunner *JobRunner, interact *interactpkg.Interact,
	db *database.Database, log logg.Logg) *Search {

	return &Search{
		jobBuilder: jobBuilder,
		jobRunner:  jobRunner,
		interact:   interact,
		db:         db,
		log:        log,
	}
}

func (s *Search) Search() (err error) {
	stats.SearchStartTime = time.Now()

	s.log.Info("finding secrets ...")

	// Progress bar for the search jobs
	jobProg := s.interact.NewProgress()

	// Create jobs
	var jobs []*Job
	jobs, err = s.jobBuilder.BuildJobs(jobProg)
	if err != nil {
		err = errors.WithMessage(err, "unable to build jobs")
		return
	}
	commitsLen := countCommits(jobs)

	// Run jobs
	s.jobRunner.runJobs(jobs)

	// End the progress bar and resume logging to stdout
	if jobProg != nil {
		jobProg.Wait()
	}

	s.log.Info("completed search")

	// Stats
	stats.CommitsSearchedCount = int64(commitsLen)
	stats.SecretsFoundCount = int64(s.jobRunner.secretCount)
	stats.SearchEndTime = time.Now()

	return
}

func countCommits(jobs []*Job) (result int) {
	for _, job := range jobs {
		result += len(job.commitHashes)
	}
	return
}
