package source

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pantheon-systems/secrets-searcher/pkg/database"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	gitpkg "github.com/pantheon-systems/secrets-searcher/pkg/git"
	interactpkg "github.com/pantheon-systems/secrets-searcher/pkg/interact"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/sirsean/go-pool"
)

type (
	Source struct {
		sourceDir   string
		skipFetch   bool
		workerCount int
		repoFilter  *manip.SliceFilter
		git         *gitpkg.Git
		provider    ProviderI
		interact    *interactpkg.Interact
		db          *database.Database
		log         logg.Logg
	}
	ProviderI interface {
		manip.Named
		GetRepositories(repoFilter *manip.SliceFilter) (result []*RepoInfo, err error)
		GetRepoURL(repoName string) string
		GetCommitURL(repoName, commitHash string) string
		GetFileURL(repoName, commitHash, filePath string) string
		GetFileLineURL(repoName, commitHash, filePath string, startLineNum, endLineNum int) string
	}
	RepoInfo struct {
		Name           string
		SourceProvider string
		RemoteURL      string
	}
)

func New(
	sourceDir string,
	skipFetch bool,
	workerCount int,
	repoFilter *manip.SliceFilter,
	git *gitpkg.Git,
	provider ProviderI,
	interact *interactpkg.Interact,
	db *database.Database,
	log logg.Logg,
) *Source {
	return &Source{
		sourceDir:   sourceDir,
		skipFetch:   skipFetch,
		workerCount: workerCount,
		repoFilter:  repoFilter,
		git:         git,
		provider:    provider,
		interact:    interact,
		db:          db,
		log:         log,
	}
}

func (s *Source) PrepareSource() (err error) {
	s.log.Info("preparing repos ...")

	var repoInfos []*RepoInfo
	repoInfos, err = s.getRepoInfos()
	if err != nil {
		err = errors.WithMessage(err, "unable to get repo infos")
		return
	}

	s.cloneRepos(repoInfos)

	// Check
	if err = s.checkPrepare(repoInfos); err != nil {
		err = errors.WithMessage(err, "there were problems preparing the source")
		return
	}

	return
}

func (s *Source) cloneRepos(repoInfos []*RepoInfo) {
	p := pool.NewPool(len(repoInfos), s.workerCount)
	p.Start()

	prog := s.interact.NewProgress()

	for _, repoInfo := range repoInfos {
		log := s.log.WithField("repo", repoInfo.Name)
		cloneDir := filepath.Join(s.sourceDir, repoInfo.Name)

		p.Add(newCloneWorker(repoInfo, cloneDir, s.skipFetch, s.git, prog, s.db, log))
	}

	p.Close()
	if prog != nil {
		prog.Wait()
	}
}

func (s *Source) getRepoInfos() (result []*RepoInfo, err error) {

	// Build repo infos from database repos
	result, err = s.getRepoInfosFromDatabase()
	if err != nil {
		err = errors.WithMessage(err, "unable to get repo infos from database")
		return
	}

	// Get repo infos from source provider if we haven't already gotten the repo list from the database
	if result == nil {
		s.interact.SpinWhile("querying source provider for repo info", func() {
			s.log.Debug("querying source provider for repo info")
			result, err = s.provider.GetRepositories(s.repoFilter)
		})

		if result == nil {
			err = errors.WithMessage(err, "failed to get repos from provider")
			return
		}
	}

	// Sort by name
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	// Not we have repo data in memory so delete the repo table. the cloneWorker class will save the correct list again
	s.log.Debug("deleting repo table")
	if err = s.db.DeleteRepoTable(); err != nil {
		err = errors.WithMessage(err, "unable to delete repo table")
		return
	}

	return
}

func (s *Source) getRepoInfosFromDatabase() (result []*RepoInfo, err error) {
	// If the filter matches the existing list of repos in the database, we'll just proceed
	if !s.db.RepoTableExists() || !s.repoFilter.CanProvideExactValues() {
		s.log.Debug("not possible to use the repo info from the database, querying provider")
		return
	}

	// Build repo infos from database repos
	var repoInfos []*RepoInfo
	var repoNames manip.Set
	repoInfos, repoNames, err = s.getFilteredRepoInfosFromDatabase()
	if err != nil {
		err = errors.WithMessage(err, "unable to get repo infos from database")
		return
	}

	// See if the filtered list equals the filter provided exactly
	repoFilterNameStrings := s.repoFilter.ExactValues().StringValues()
	repoNameStrings := repoNames.StringValues()
	if !manip.StringValuesEqualAfterSort(repoFilterNameStrings, repoNameStrings) {
		s.log.Debug("unable to get repo info from database")
		return
	}

	s.log.Debug("using repo information from database since it contains all repos from the filter, " +
		"skipping source provider")
	result = repoInfos

	return
}

func (s *Source) getFilteredRepoInfosFromDatabase() (result []*RepoInfo, repoNames *manip.BasicSet, err error) {
	var dbRepos []*database.Repo
	dbRepos, err = s.db.GetRepos()
	if err != nil {
		err = errors.WithMessage(err, "unable to get filtered repo infos from database")
		return
	}

	repoNames = manip.NewEmptyBasicSet()
	for _, dbRepo := range dbRepos {
		if !s.repoFilter.Includes(dbRepo.Name) {
			continue
		}

		result = append(result, &RepoInfo{
			Name:           dbRepo.Name,
			SourceProvider: s.provider.GetName(),
			RemoteURL:      dbRepo.RemoteURL,
		})

		repoNames.Add(dbRepo.Name)
	}
	return
}

func (s *Source) checkPrepare(repoInfos []*RepoInfo) (err error) {
	if !s.db.RepoTableExists() {
		err = errors.New("no repos are prepared")
		return
	}

	// Build repo infos from database repos
	var repoInfosInDB []*RepoInfo
	repoInfosInDB, _, err = s.getFilteredRepoInfosFromDatabase()
	if err != nil {
		err = errors.WithMessage(err, "unable to get repo infos from database to check them")
		return
	}

	if repoInfosInDB == nil {
		err = errors.New("none of our repos are present in the database")
		return
	}

	return
}
