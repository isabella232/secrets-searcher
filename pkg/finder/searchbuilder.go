package finder

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "math"
    "sort"
    "time"
)

type SearchBuilder struct {
    git                 *gitpkg.Git
    repoFilter          *structures.Filter
    commitFilter        *gitpkg.CommitFilter
    fileChangeFilter    *gitpkg.FileChangeFilter
    commitSearchTimeout time.Duration
    processors          []Processor
    workerCount         int
    chunkSize           int
    showWorkersBars     bool
    interact            interactpkg.Interactish
    db                  *database.Database
    log                 logrus.FieldLogger
}

func NewSearchBuilder(
    git *gitpkg.Git,
    repoFilter *structures.Filter,
    commitFilter *gitpkg.CommitFilter,
    fileChangeFilter *gitpkg.FileChangeFilter,
    commitSearchTimeout time.Duration,
    processors []Processor,
    chunkSize int,
    workerCount int,
    showWorkersBars bool,
    interact interactpkg.Interactish,
    db *database.Database,
    log logrus.FieldLogger,
) (result *SearchBuilder) {
    return &SearchBuilder{
        git:                 git,
        repoFilter:          repoFilter,
        commitFilter:        commitFilter,
        fileChangeFilter:    fileChangeFilter,
        commitSearchTimeout: commitSearchTimeout,
        processors:          processors,
        workerCount:         workerCount,
        chunkSize:           chunkSize,
        showWorkersBars:     showWorkersBars,
        interact:            interact,
        db:                  db,
        log:                 log,
    }
}

func (s *SearchBuilder) getJobs(prog *progress.Progress, out chan *searchResult) (result []Search, totalCommitCount int, err error) {

    // Build targets
    var targets []*searchTarget
    var repoCounts map[string]int
    targets, repoCounts, totalCommitCount, err = s.buildJobParams()
    if err != nil {
        err = errors.WithMessage(err, "unable to build search targets")
        return
    }
    targetsLen := len(targets)

    result = make([]Search, targetsLen)
    for i, target := range targets {

        // Name
        workerName := fmt.Sprintf("%s-%d", target.repo.Name, i)

        // Log
        jobLog := s.log.
            WithField("worker", workerName).
            WithField("commitsLen", len(target.commitHashes))
        jobLog.Debugf("adding finder worker for repo")

        // Progress bar
        bar := s.getBar(prog, target, workerName, repoCounts)

        // Create job
        job := newSearch(
            out,
            workerName,
            target,
            s.processors,
            s.fileChangeFilter,
            s.commitSearchTimeout,
            bar,
            jobLog,
        )

        result[i] = job
    }

    return
}

func (s *SearchBuilder) buildJobParams() (result []*searchTarget, repoCounts map[string]int, commitCount int, err error) {
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

    // Get targets, grouped by repo name
    var targetsByRepo = make(map[string][]*searchTarget, reposLen)

    // Also get a total count of commits for each repo
    repoCounts = make(map[string]int, reposLen)

    var prog *progress.Progress
    prog = s.interact.NewProgress()
    var bar *progress.Bar
    if prog != nil {
        bar = prog.AddBar("gathering commits", reposLen, "%d of %d repos", "commits gathered")
        bar.Start()
    }

    for _, repo := range repos {
        func() {
            if bar != nil {
                defer bar.Incr()
            }

            // Get repository and commit objects
            var repository *gitpkg.Repository
            var commitHashes []string
            var oldest string
            repository, commitHashes, oldest, err = s.repositoryAndCommits(repo)
            if err != nil {
                err = errors.WithMessagev(err, "unable to get repositories and commits for repo", repo.Name)
                return
            }

            // Add counts
            repoCounts[repo.Name] += len(commitHashes)
            commitCount += len(commitHashes)

            var repoTargets []*searchTarget
            repoTargets, err = s.buildRepoTargets(repo, repository, commitHashes, oldest)
            if err != nil {
                err = errors.WithMessagev(err, "unable to build repo targets for repo", repo.Name)
                return
            }

            for _, repoTarget := range repoTargets {
                targetsByRepo[repo.Name] = append(targetsByRepo[repo.Name], repoTarget)
            }
        }()
    }

    for {
        var batch []*searchTarget
        batch, targetsByRepo, err = s.getBatch(targetsByRepo)
        if err != nil {
            err = errors.WithMessage(err, "unable to get batch")
            return
        }
        if batch == nil {
            break
        }
        result = append(result, batch...)
    }

    return
}

func (s *SearchBuilder) getBatch(targetsByRepo map[string][]*searchTarget) (result []*searchTarget, rest map[string][]*searchTarget, err error) {
    if len(targetsByRepo) == 0 {
        return
    }

    max := s.workerCount

    // Get next target for each repo
    for {
        var firstTargets []*searchTarget
        firstTargets, targetsByRepo, err = s.getFirstTargets(targetsByRepo, max)
        if err != nil {
            err = errors.WithMessage(err, "unable to get first target")
            return
        }
        if firstTargets == nil {
            break
        }

        result = append(result, firstTargets...)

        max = s.workerCount - len(result)

        if max == 0 {
            break
        }
    }

    rest = targetsByRepo

    return
}

func (s *SearchBuilder) getFirstTargets(targetsByRepo map[string][]*searchTarget, max int) (result []*searchTarget, rest map[string][]*searchTarget, err error) {
    if len(targetsByRepo) == 0 {
        return
    }

    // Get first n repo names
    repoNames := make([]string, len(targetsByRepo))
    i := 0
    for k := range targetsByRepo {
        repoNames[i] = k
        i++
    }
    sort.Strings(repoNames)

    // Get next target for each repo
    collected := 0
    for _, repoName := range repoNames {
        // Get next target from repo targets and delete it from the source
        result = append(result, targetsByRepo[repoName][0])
        targetsByRepo[repoName] = targetsByRepo[repoName][1:]

        // Delete repo from map if its slice is empty now
        if len(targetsByRepo[repoName]) == 0 {
            delete(targetsByRepo, repoName)
        }

        collected += 1

        if collected == max {
            break
        }
    }

    rest = targetsByRepo

    return
}

func (s *SearchBuilder) buildRepoTargets(repo *database.Repo, repository *gitpkg.Repository, commitHashes []string, oldest string) (result []*searchTarget, err error) {
    // Get chunks of commits
    commitHashChunks := s.chunkCommits(commitHashes)
    commitHashChunksLen := len(commitHashChunks)

    result = make([]*searchTarget, commitHashChunksLen)

    for i := 0; i < commitHashChunksLen; i++ {
        chunkCommitHashes := commitHashChunks[i]

        var spawnedRepository *gitpkg.Repository
        spawnedRepository, err = repository.Spawn()
        if err != nil {
            err = errors.WithMessage(err, "unable to spawn repository")
            return
        }

        result[i] = &searchTarget{
            repo:         repo,
            repository:   spawnedRepository,
            commitHashes: chunkCommitHashes,
            oldest:       oldest,
        }
    }

    return
}

func (s *SearchBuilder) repositoryAndCommits(repo *database.Repo) (repository *gitpkg.Repository, commitHashes []string, oldest string, err error) {
    repository, err = s.git.NewRepository(repo.CloneDir)
    if err != nil {
        err = errors.Wrapv(err, "unable to open git repository", repo.CloneDir)
        return
    }

    commitHashes, oldest, err = s.commitHashes(repo, repository)
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

    return
}

func (s *SearchBuilder) commitHashes(repo *database.Repo, repository *gitpkg.Repository) (result []string, oldest string, err error) {
    // From cache
    if result, oldest = s.commitHashesFromCache(repo); result != nil {
        s.log.WithField("repo", repo.Name).Debugf("commits cache found for repo")
        return
    }

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

    s.log.WithField("repo", repo.Name).Debugf("%d commits found for repo", commitsLen)

    // Write cache
    cacheErr := s.db.WriteRepoCommitsCache(&database.RepoCommitsCache{
        RepoName:   repo.Name,
        Hashes:     result,
        OldestHash: result[len(result)-1],
    })
    if cacheErr != nil {
        errors.ErrLog(s.log, cacheErr).Error("unable to cache get repo commits")
    }

    return
}

func (s *SearchBuilder) commitHashesFromLog(repository *gitpkg.Repository) (result []string, oldest string, err error) {
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

func (s *SearchBuilder) commitHashesFromCache(repo *database.Repo) (result []string, oldest string) {
    if !s.commitFilter.IncludesAll() {
        return
    }

    repoCommitsCache, cacheErr := s.db.GetRepoCommitsCache(repo.Name)
    if cacheErr != nil {
        errors.ErrLog(s.log, cacheErr).Error("unable to get repo commits cache")
        return
    }

    if repoCommitsCache == nil {
        return
    }

    s.log.WithField("repo", repo.Name).Debugf("commits cache found for repo")
    result = repoCommitsCache.Hashes
    oldest = repoCommitsCache.OldestHash

    return
}

func (s *SearchBuilder) chunkCommits(items []string) (result [][]string) {
    itemsLen := len(items)
    chunksLen := int(math.Ceil(float64(itemsLen) / float64(s.chunkSize)))
    result = make([][]string, chunksLen)

    for i, _ := range result {
        start := i * s.chunkSize
        end := start + s.chunkSize
        if end > itemsLen {
            end = itemsLen
        }
        result[i] = items[start:end]
    }

    return
}

func (s *SearchBuilder) getBar(prog *progress.Progress, target *searchTarget, workerName string, repoCounts map[string]int) (result *progress.Bar) {
    if prog == nil {
        return
    }

    var barName string
    var totalCommitCount int
    if s.showWorkersBars {
        barName = workerName
        totalCommitCount = len(target.commitHashes)
    } else {
        barName = target.repo.Name
        totalCommitCount = repoCounts[target.repo.Name]
    }

    return prog.AddBar(barName, totalCommitCount, "searched %d of %d commits", "search of %s is complete")
}
