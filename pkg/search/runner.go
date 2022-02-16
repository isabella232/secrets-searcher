package search

import (
	"sync"

	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/contract"
)

type (
	JobRunner struct {
		workers        []*Worker
		dbResultWriter *dbResultWriter
		log            logg.Logg
		*jobRunnerState
	}
	jobRunnerState struct {
		secretCount int
		jobQueue    chan *Job
		resultsChan chan *contract.JobResult
		jobsWG      *sync.WaitGroup
		processWG   *sync.WaitGroup
	}
)

func NewJobRunner(workers []*Worker, dbResultWriter *dbResultWriter, log logg.Logg) *JobRunner {
	return &JobRunner{
		workers:        workers,
		dbResultWriter: dbResultWriter,
		log:            log,
	}
}

func (r *JobRunner) runJobs(jobs []*Job) {
	r.resetState(jobs)

	// Start workers hanging on job queue
	for _, worker := range r.workers {
		go r.putToWork(worker)
	}

	// Queue the jobs and start the work
	for _, job := range jobs {
		r.jobQueue <- job
	}
	close(r.jobQueue)

	// Pull results from the results channel and write to the database
	go r.processJobResults()

	// Wait for all jobs to finish
	r.jobsWG.Wait()
	// Close result queue to further writes
	close(r.resultsChan)
	// Wait for the last result to be processed
	r.processWG.Wait()
}

// This is run once per worker. The workers all continuously pull from
// the job queue until it there are no more jobs, then they stop.
func (r *JobRunner) putToWork(worker *Worker) {
	for job := range r.jobQueue {

		// Worker performs the job
		worker.Do(job)

		// Get job results
		jobResults := job.GetJobResults()

		// Write results to results channel for processing
		for _, jobResult := range jobResults {
			r.resultsChan <- jobResult
		}

		// One down, more to go
		r.jobsWG.Done()
	}
}

// This is run in a goroutine to process job results as they come in and write them
// to the database.
func (r *JobRunner) processJobResults() {
	for result := range r.resultsChan {
		if err := r.dbResultWriter.WriteResult(result); err != nil {
			errors.ErrLog(r.log, err).Error("error writing search result")
			continue
		}
	}

	r.secretCount = r.dbResultWriter.secretTracker.Len()

	r.processWG.Done()
}

func (r *JobRunner) resetState(jobs []*Job) {
	jobCount := len(jobs)

	jobsWG := &sync.WaitGroup{}
	jobsWG.Add(jobCount)

	processWG := &sync.WaitGroup{}
	processWG.Add(1)

	r.jobRunnerState = &jobRunnerState{
		secretCount: 0,
		jobQueue:    make(chan *Job, jobCount),
		resultsChan: make(chan *contract.JobResult),
		jobsWG:      jobsWG,
		processWG:   processWG,
	}
}
