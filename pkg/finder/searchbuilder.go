package finder

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "math"
    "sort"
)

type (
    searchBuilder struct {
        git          *gitpkg.Git
        repoFilter   *structures.Filter
        commitFilter *gitpkg.CommitFilter

        // How many commits for each payload
        chunkSize int

        // Queue the next payload from each of the first n repos, then repeat until there are no more payloads to queue
        workerCount int

        interact interactpkg.Interactish
        db       *database.Database
        log      logrus.FieldLogger
    }
    searchParameters struct {
        repo         *database.Repo
        repository   *gitpkg.Repository
        commits      []*gitpkg.Commit
        commitHashes []string
        commitsLen   int
    }
)

func newPayloadSet(git *gitpkg.Git, repoFilter *structures.Filter, commitFilter *gitpkg.CommitFilter, workerCount, chunkSize int, interact interactpkg.Interactish, db *database.Database, log logrus.FieldLogger) (result *searchBuilder) {
    return &searchBuilder{
        git:          git,
        repoFilter:   repoFilter,
        commitFilter: commitFilter,
        chunkSize:    chunkSize,
        workerCount:  workerCount,
        interact: interact,
        db:           db,
        log:          log,
    }
}

func (s *searchBuilder) buildPayloads() (result []*searchParameters, repoCounts map[string]int, commitCount int, err error) {
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

    // Get payloads, grouped by repo name
    var payloadsByRepo = make(map[string][]*searchParameters, reposLen)

    // Also get a total count of commits for each repo
    repoCounts = make(map[string]int, reposLen)

    var prog *progress.Progress
    prog = s.interact.NewProgress()
    var bar *progress.Bar
    if prog != nil {
        bar = prog.AddBar("gathering commits", reposLen, "%d of %d repos", "search of %s completed")
        bar.Start()
    }

    for _, repo := range repos {
        func() {
            if bar != nil {
                defer bar.Incr()
            }

            // Get repository and commit objects
            var repository *gitpkg.Repository
            var commits []*gitpkg.Commit
            repository, commits, err = s.repositoryAndCommits(repo)
            if err != nil {
                err = errors.WithMessagev(err, "unable to get repositories and commits for repo", repo.Name)
                return
            }

            // Add count
            repoCounts[repo.Name] += len(commits)

            commitCount += len(commits)

            var repoPayloads []*searchParameters
            repoPayloads, err = s.buildRepoPayloads(repo, repository, commits)
            if err != nil {
                err = errors.WithMessagev(err, "unable to build repo payloads for repo", repo.Name)
                return
            }

            for _, repoPayload := range repoPayloads {
                payloadsByRepo[repo.Name] = append(payloadsByRepo[repo.Name], repoPayload)
            }
        }()
    }

    for {
        var batch []*searchParameters
        batch, payloadsByRepo, err = s.getBatch(payloadsByRepo)
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

func (s *searchBuilder) getBatch(payloadsByRepo map[string][]*searchParameters) (result []*searchParameters, rest map[string][]*searchParameters, err error) {
    if len(payloadsByRepo) == 0 {
        return
    }

    max := s.workerCount

    // Get next payload for each repo
    for {
        var firstPayloads []*searchParameters
        firstPayloads, payloadsByRepo, err = s.getFirstPayloads(payloadsByRepo, max)
        if err != nil {
            err = errors.WithMessage(err, "unable to get first payloads")
            return
        }
        if firstPayloads == nil {
            break
        }

        result = append(result, firstPayloads...)

        max = s.workerCount - len(result)

        if max == 0 {
            break
        }
    }

    rest = payloadsByRepo

    return
}

func (s *searchBuilder) getFirstPayloads(payloadsByRepo map[string][]*searchParameters, max int) (result []*searchParameters, rest map[string][]*searchParameters, err error) {
    if len(payloadsByRepo) == 0 {
        return
    }

    // Get first n repo names
    repoNames := make([]string, len(payloadsByRepo))
    i := 0
    for k := range payloadsByRepo {
        repoNames[i] = k
        i++
    }
    sort.Strings(repoNames)

    // Get next payload for each repo
    collected := 0
    for _, repoName := range repoNames {
        // Get next payload from repo payloads and delete it from the source
        result = append(result, payloadsByRepo[repoName][0])
        payloadsByRepo[repoName] = payloadsByRepo[repoName][1:]

        // Delete repo from map if its slice is empty now
        if len(payloadsByRepo[repoName]) == 0 {
            delete(payloadsByRepo, repoName)
        }

        collected += 1

        if collected == max {
            break
        }
    }

    rest = payloadsByRepo

    return
}

func (s *searchBuilder) buildRepoPayloads(repo *database.Repo, repository *gitpkg.Repository, commits []*gitpkg.Commit) (result []*searchParameters, err error) {
    // Get chunks of commits
    commitChunks := s.chunkCommits(commits)
    commitChunksLen := len(commitChunks)

    result = make([]*searchParameters, commitChunksLen)

    for i := 0; i < commitChunksLen; i++ {
        chunkCommits := commitChunks[i]
        chunkCommitsLen := len(chunkCommits)

        // Set first payload using the original repository and commits
        if i == 0 {
            result[i] = newPayload(repo, repository, chunkCommits, nil)
            continue
        }

        // Set subsequent payloads, each with a fresh repository object,
        // to avoid race conditions (https://github.com/src-d/go-git/issues/702).
        // To save memory, the payloads will have commit hashes only.
        var spawnedRepository *gitpkg.Repository
        spawnedRepository, err = repository.Spawn()
        if err != nil {
            err = errors.WithMessage(err, "unable to spawn repository")
            return
        }

        // Get hashes of commits
        chunkCommitHashes := make([]string, chunkCommitsLen)
        for i, chunkCommit := range chunkCommits {
            chunkCommitHashes[i] = chunkCommit.Hash
        }

        result[i] = newPayload(repo, spawnedRepository, nil, chunkCommitHashes)
    }

    return
}

func (s *searchBuilder) repositoryAndCommits(repo *database.Repo) (repository *gitpkg.Repository, commits []*gitpkg.Commit, err error) {
    repository, err = s.git.NewRepository(repo.CloneDir)
    if err != nil {
        err = errors.Wrapv(err, "unable to open git repository", repo.CloneDir)
        return
    }

    commits, err = repository.Log(s.commitFilter)
    if err != nil {
        err = errors.Wrap(err, "unable to run git log")
        return
    }
    commitsLen := len(commits)

    s.log.WithField("repo", repo.Name).Debugf("%d commits found for repo", commitsLen)

    if len(commits) == 0 {
        err = errors.New("no commits found in repo")
        return
    }

    return
}

func (s *searchBuilder) chunkCommits(items []*gitpkg.Commit) (result [][]*gitpkg.Commit) {
    itemsLen := len(items)
    chunksLen := int(math.Ceil(float64(itemsLen) / float64(s.chunkSize)))
    result = make([][]*gitpkg.Commit, chunksLen)

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

func newPayload(repo *database.Repo, repository *gitpkg.Repository, commits []*gitpkg.Commit, commitHashes []string) *searchParameters {
    var commitsLen int
    if commits != nil {
        commitsLen = len(commits)
    }
    if commitHashes != nil {
        commitsLen = len(commitHashes)
    }

    return &searchParameters{
        repo:         repo,
        repository:   repository,
        commits:      commits,
        commitHashes: commitHashes,
        commitsLen:   commitsLen,
    }
}

func (p *searchParameters) getCommits() (result []*gitpkg.Commit, err error) {
    if p.commits != nil {
        result = p.commits
        return
    }

    result = make([]*gitpkg.Commit, len(p.commitHashes))
    for i, commitHash := range p.commitHashes {
        var commit *gitpkg.Commit
        commit, err = p.repository.Commit(commitHash)
        if err != nil {
            err = errors.WithMessagev(err, "unable to get commit", commitHash)
            return
        }
        result[i] = commit
    }

    p.commits = result

    return
}
