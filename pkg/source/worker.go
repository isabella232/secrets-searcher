package source

import (
	"fmt"
	"io"
	"os"

	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	"github.com/pantheon-systems/search-secrets/pkg/interact/progress"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

type cloneWorker struct {
	repoInfo  *RepoInfo
	cloneDir  string
	skipFetch bool
	git       *gitpkg.Git
	prog      *progress.Progress
	db        *database.Database
	log       logg.Logg
}

func newCloneWorker(repoInfo *RepoInfo, cloneDir string, skipFetch bool, git *gitpkg.Git, prog *progress.Progress,
	db *database.Database, log logg.Logg) *cloneWorker {

	return &cloneWorker{
		repoInfo:  repoInfo,
		cloneDir:  cloneDir,
		skipFetch: skipFetch,
		git:       git,
		prog:      prog,
		db:        db,
		log:       log,
	}
}

func (w *cloneWorker) Perform() {
	if err := w.prepareRepo(w.repoInfo); err != nil {
		errors.ErrLog(w.log, err).Error("unable to prepare repo: ", w.repoInfo.Name)
	}
}

func (w *cloneWorker) prepareRepo(ghRepo *RepoInfo) (err error) {
	var bar *progress.Spinner
	if w.prog != nil {
		bar = w.prog.AddSpinner(w.repoInfo.Name)
	}

	defer func() {
		if w.prog != nil {
			bar.Incr()

			// Error handling
			doneMessage := "- %s repo prepared for search"
			if err != nil {
				bar.BustThrough(func() { errors.ErrLog(w.log, err).Error(err.Error()) })
				doneMessage = "- %s repo not prepared for search!"
			}

			// Send a done message to stdout to replace the bar
			w.prog.Add(0, mpb.BarFillerFunc(func(writer io.Writer, width int, st *decor.Statistics) {
				_, _ = fmt.Fprintf(writer, doneMessage, w.repoInfo.Name)
			})).SetTotal(0, true)
		}
	}()

	cloneDirTmp := w.cloneDir + "-CLONING"

	// Remove temporary clone
	if err = os.RemoveAll(cloneDirTmp); err != nil {
		err = errors.Wrapv(err, "unable to remove temporary clone dir repo, skipping", cloneDirTmp)
		return
	}

	// Remove the clone if it is corrupt
	if err = w.removeExistingCorruptClone(); err != nil {
		err = errors.Wrap(err, "unable to remove corrupt repo, skipping")
		return
	}

	// Clone
	if _, statErr := os.Stat(w.cloneDir); os.IsNotExist(statErr) {
		w.log.Debug("cloning repo")

		if err = w.cloneRepo(ghRepo.RemoteURL, cloneDirTmp, w.log); err != nil {
			err = errors.WithMessagev(err, "unable to clone repo, skipping", ghRepo.RemoteURL, cloneDirTmp)
			return
		}
		if err = os.Rename(cloneDirTmp, w.cloneDir); err != nil {
			err = errors.Wrapv(err, "unable to rename temporary clone dir repo, skipping", cloneDirTmp, w.cloneDir)
			return
		}
		// Check for directory again
		if _, statErr := os.Stat(w.cloneDir); os.IsNotExist(statErr) {
			err = errors.Wrapv(err, "repo clone failed, skipping", ghRepo.RemoteURL, cloneDirTmp)
			return
		}
	} else if !w.skipFetch {
		w.log.Debug("fetching repo")

		if err = w.fetchRepo(ghRepo.RemoteURL, w.cloneDir); err != nil {
			err = errors.WithMessagev(err, "unable to fetch repo", ghRepo.RemoteURL, cloneDirTmp)
			return
		}
	}

	err = w.db.WriteRepo(&database.Repo{
		ID:        database.CreateHashID(ghRepo.RemoteURL),
		Name:      ghRepo.Name,
		RemoteURL: ghRepo.RemoteURL,
	})
	if err != nil {
		err = errors.WithMessagev(err, "unable to write repo, skipping", ghRepo.Name)
		return
	}

	w.log.Debug("repo is prepared")
	return
}

func (w *cloneWorker) removeExistingCorruptClone() (err error) {
	if _, statErr := os.Stat(w.cloneDir); os.IsNotExist(statErr) {
		return
	}

	if !w.git.IsCloneValid(w.cloneDir) {
		w.log.Debug("removing corrupt repo")

		if err = os.RemoveAll(w.cloneDir); err != nil {
			err = errors.Wrapv(err, "unable to remove corrupt clone dir repo, skipping", w.cloneDir)
			return
		}
	}

	return
}

func (w *cloneWorker) cloneRepo(url, cloneDir string, log logg.Logg) (err error) {
	if _, err = w.git.Clone(url, cloneDir); err != nil {
		if gitpkg.IsErrEmptyRemoteRepository(err) {
			log.Warn("clone failed because remote repo has no commits, skipping")
		}
		_ = os.RemoveAll(cloneDir)
	}

	return
}

func (w *cloneWorker) fetchRepo(url, cloneDir string) (err error) {
	var repository *gitpkg.Repository
	if repository, err = w.git.OpenRepository(cloneDir); err != nil {
		err = errors.WithMessagev(err, "unable to open git repository", cloneDir)
		return
	}

	if err = repository.FetchAll(url); err != nil {
		err = errors.Wrapv(err, "unable to fetch all from repository", cloneDir)
	}

	return
}
