package search

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"

	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
	"github.com/pantheon-systems/search-secrets/pkg/interact/progress"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

const (
	barSuffixMsg    = "searched %d of %d commits"
	barCompletedMsg = "search of %s is complete"
)

type (
	JobBuilder struct {
		// Filters
		repoFilter   manip.Filter
		commitFilter *gitpkg.CommitFilter

		sourceDir string

		// Execution parameters
		workerCount   int
		chunkSize     int
		showBarPerJob bool

		// Services
		git *gitpkg.Git

		interact *interactpkg.Interact
		db       *database.Database
		log      logg.Logg
	}

	repoData struct {
		RepoID       string
		RepoName     string
		Repository   *gitpkg.Repository
		CommitHashes []string
		Oldest       string
		CommitCount  int
	}
)

func NewJobBuilder(
	repoFilter *manip.SliceFilter,
	sourceDir string,
	commitFilter *gitpkg.CommitFilter,
	workerCount int,
	chunkSize int,
	showBarPerJob bool,
	git *gitpkg.Git,
	interact *interactpkg.Interact,
	db *database.Database,
	log logg.Logg,
) (result *JobBuilder) {

	return &JobBuilder{
		repoFilter:    repoFilter,
		sourceDir:     sourceDir,
		commitFilter:  commitFilter,
		workerCount:   workerCount,
		chunkSize:     chunkSize,
		showBarPerJob: showBarPerJob,
		git:           git,
		interact:      interact,
		db:            db,
		log:           log,
	}
}

func (s *JobBuilder) BuildJobs(jobProg *progress.Progress) (result []*Job, err error) {

	// Get repos from database
	var repos []*database.Repo
	if s.repoFilter != nil {
		repos, err = s.db.GetReposFilteredSorted(s.repoFilter)
	} else {
		repos, err = s.db.GetReposSorted()
	}
	if err != nil {
		err = errors.WithMessage(err, "unable to get repos")
		return
	}
	reposLen := len(repos)

	// Local progress bar
	var bar *progress.Bar
	prog := s.interact.NewProgress()
	if prog != nil {
		bar = prog.AddBar("gathering commits", reposLen, "%d of %d repos", "commits gathered")
		bar.Start()
	}

	// Get jobs, grouped by repo name
	var jobByRepo = make(map[string][]*Job, reposLen)
	for _, repo := range repos {
		err = func() (err error) {
			if bar != nil {
				defer bar.Incr()
			}

			// Get repository and commit objects
			var repoDat *repoData
			repoDat, err = s.repoData(repo)
			if err != nil {
				err = errors.WithMessagev(err, "unable to get repositories and commits for repo", repo.Name)
				return
			}

			var repoJobs []*Job
			repoJobs, err = s.buildRepoJobs(repoDat, jobProg)
			if err != nil {
				err = errors.WithMessagev(err, "unable to build repo jobs for repo", repo.Name)
				return
			}

			jobByRepo[repo.Name] = repoJobs

			return
		}()
		if err != nil {
			return
		}
	}

	// Build a flat list of jobs.
	// To avoid too many workers reading from the same clone directory at once,
	// we will spread the jobs out so that each worker gets its own repo to work on.
	// However, to avoid too many repos being worked on at once, we will
	// So if there are 2 workers and repos A to D to work on, the jobs would be ordered like:
	// A, B, A, B, A, B, A, B, A, B finishes, A, C, A, C, A, C, etc
	for {
		var batch []*Job
		batch, jobByRepo, err = s.getBatch(jobByRepo)
		if err != nil {
			err = errors.WithMessage(err, "unable to get batch")
			return
		}
		if batch == nil {
			break
		}
		result = append(result, batch...)
	}

	// Wait for builder's progress bar to end
	if prog != nil {
		prog.Wait()
	}

	return
}

func (s *JobBuilder) getBatch(jobsByRepo map[string][]*Job) (result []*Job, rest map[string][]*Job, err error) {
	if len(jobsByRepo) == 0 {
		return
	}

	max := s.workerCount

	// Get next job for each repo
	for {
		var firstJobs []*Job
		firstJobs, jobsByRepo, err = s.getFirstJobs(jobsByRepo, max)
		if err != nil {
			err = errors.WithMessage(err, "unable to get first job")
			return
		}
		if firstJobs == nil {
			break
		}

		result = append(result, firstJobs...)

		max = s.workerCount - len(result)

		if max == 0 {
			break
		}
	}

	rest = jobsByRepo

	return
}

func (s *JobBuilder) getFirstJobs(jobsByRepo map[string][]*Job, max int) (result []*Job, rest map[string][]*Job, err error) {
	if len(jobsByRepo) == 0 {
		return
	}

	// Get first n repo names
	repoNames := make([]string, len(jobsByRepo))
	i := 0
	for k := range jobsByRepo {
		repoNames[i] = k
		i++
	}
	sort.Strings(repoNames)

	// Get next job for each repo
	collected := 0
	for _, repoName := range repoNames {
		// Get next job from repo jobs and delete it from the source
		result = append(result, jobsByRepo[repoName][0])
		jobsByRepo[repoName] = jobsByRepo[repoName][1:]

		// Delete repo from map if its slice is empty now
		if len(jobsByRepo[repoName]) == 0 {
			delete(jobsByRepo, repoName)
		}

		collected += 1

		if collected == max {
			break
		}
	}

	rest = jobsByRepo

	return
}

func (s *JobBuilder) buildRepoJobs(repoDat *repoData, jobProg *progress.Progress) (result []*Job, err error) {
	// Get chunks of commits
	commitHashChunks := s.chunkCommits(repoDat.CommitHashes)
	commitHashChunksLen := len(commitHashChunks)

	result = make([]*Job, commitHashChunksLen)

	for i := 0; i < commitHashChunksLen; i++ {
		jobCommitHashes := commitHashChunks[i]
		jobCommitCount := len(jobCommitHashes)

		var spawnedRepository *gitpkg.Repository
		spawnedRepository, err = repoDat.Repository.Spawn()
		if err != nil {
			err = errors.WithMessage(err, "unable to spawn repository")
			return
		}

		jobName := fmt.Sprintf("%s-%d", repoDat.RepoName, i)

		// Log
		jobLog := s.log.WithFields(logg.Fields{
			"prefix":     "search/job",
			"repo":       repoDat.RepoName,
			"job":        jobName,
			"commitsLen": jobCommitCount,
		})

		// Progress bar
		var jobBar *progress.Bar
		if jobProg != nil {
			var barName string
			var commitCount int
			if s.showBarPerJob {
				barName = jobName
				commitCount = jobCommitCount
			} else {
				barName = repoDat.RepoName
				commitCount = repoDat.CommitCount
			}
			jobBar = jobProg.AddBar(barName, commitCount, barSuffixMsg, barCompletedMsg)
		}

		result[i] = NewJob(
			jobName,
			repoDat.RepoID,
			repoDat.RepoName,
			spawnedRepository,
			jobCommitHashes,
			repoDat.Oldest,
			jobBar,
			jobLog,
		)
	}

	return
}

func (s *JobBuilder) repoData(repo *database.Repo) (result *repoData, err error) {
	var repository *gitpkg.Repository
	var commitHashes []string
	var oldest string

	cloneDir := filepath.Join(s.sourceDir, repo.Name)
	repository, err = s.git.OpenRepository(cloneDir)
	if err != nil {
		err = errors.Wrapv(err, "unable to open git repository", cloneDir)
		return
	}

	s.log.Tracef("getting commits for %s", repo)
	commitHashes, oldest, err = s.commitHashesForRepo(repo, repository)
	if err != nil {
		err = errors.WithMessage(err, "unable to get commit hashes")
		return
	}

	commitsLen := len(commitHashes)

	s.log.WithField("repo", repo.Name).Debugf("%d commits found for repo", commitsLen)

	if len(commitHashes) == 0 {
		err = errors.New("no commits found in repo")
		return
	}

	result = &repoData{
		RepoID:       repo.ID,
		RepoName:     repo.Name,
		Repository:   repository,
		CommitHashes: commitHashes,
		CommitCount:  len(commitHashes),
		Oldest:       oldest,
	}

	return
}

func (s *JobBuilder) commitHashesForRepo(repo *database.Repo, repository *gitpkg.Repository) (result []string, oldest string, err error) {

	// From git log
	result, oldest, err = s.commitHashesFromLog(repository)
	if err != nil {
		err = errors.WithMessage(err, "unable to get commit hashes from git log")
		return
	}
	commitsLen := len(result)

	if commitsLen == 0 {
		err = errors.Errorv("no commits found in repo", repo.Name)
		return
	}

	return
}

func (s *JobBuilder) commitHashesFromLog(repository *gitpkg.Repository) (result []string, oldest string, err error) {
	var commits []*gitpkg.Commit
	commits, err = repository.Log(s.commitFilter)
	if err != nil {
		err = errors.WithMessage(err, "unable to run git log")
		return
	}
	for _, commit := range commits {
		result = append(result, commit.Hash)
		if commit.Oldest {
			oldest = commit.Hash
		}
	}

	return
}

func (s *JobBuilder) chunkCommits(items []string) (result [][]string) {
	itemsLen := len(items)
	chunksLen := int(math.Ceil(float64(itemsLen) / float64(s.chunkSize)))
	result = make([][]string, chunksLen)

	for i := range result {
		start := i * s.chunkSize
		end := start + s.chunkSize
		if end > itemsLen {
			end = itemsLen
		}
		result[i] = items[start:end]
	}

	return
}
