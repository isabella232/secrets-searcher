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
	if err := w.prepareRepo(); err != nil {
		errors.ErrLog(w.log, err).Error("unable to prepare repo, skipping", w.repoInfo.Name)
	}
}

func (w *cloneWorker) prepareRepo() (err error) {
	var bar *progress.Spinner
	if w.prog != nil {
		bar = w.prog.AddSpinner(w.repoInfo.Name)
	}

	defer func() { w.finish(bar, err) }()

	// Remove the clone if it is corrupt
	if err = w.removeExistingCorruptClone(); err != nil {
		err = errors.Wrap(err, "unable to remove corrupt repo")
		return
	}

	// Clone
	if _, statErr := os.Stat(w.cloneDir); os.IsNotExist(statErr) {
		w.log.Debug("cloning repo")

		if err = w.cloneRepo(); err != nil {
			err = errors.WithMessage(err, "unable to clone repo")
			return
		}
	} else if !w.skipFetch {
		w.log.Debug("fetching repo")

		if err = w.fetchRepo(); err != nil {
			err = errors.WithMessage(err, "unable to fetch repo")
			return
		}
	}

	err = w.db.WriteRepo(&database.Repo{
		ID:        database.CreateHashID(w.repoInfo.RemoteURL),
		Name:      w.repoInfo.Name,
		RemoteURL: w.repoInfo.RemoteURL,
	})
	if err != nil {
		err = errors.WithMessagev(err, "unable to write repo", w.repoInfo.Name)
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

func (w *cloneWorker) cloneRepo() (err error) {
	cloneDirTmp := w.cloneDir + "-CLONING"

	// Remove clones
	if err = os.RemoveAll(w.cloneDir); err != nil {
		err = errors.Wrapv(err, "unable to remove temporary clone dir", w.cloneDir)
		return
	}
	if err = os.RemoveAll(cloneDirTmp); err != nil {
		err = errors.Wrapv(err, "unable to remove temporary clone dir", cloneDirTmp)
		return
	}

	// Clone to tmp dir
	if _, err = w.git.Clone(w.repoInfo.RemoteURL, cloneDirTmp); err != nil {
		err = errors.Wrapv(err, "unable to clone", cloneDirTmp)
		_ = os.RemoveAll(cloneDirTmp)
		return
	}

	// Move into position
	if err = os.Rename(cloneDirTmp, w.cloneDir); err != nil {
		err = errors.Wrapv(err, "unable to rename temporary clone dir", cloneDirTmp, w.cloneDir)
		_ = os.RemoveAll(cloneDirTmp)
		_ = os.RemoveAll(w.cloneDir)
		return
	}

	return
}

func (w *cloneWorker) fetchRepo() (err error) {
	var repository *gitpkg.Repository
	if repository, err = w.git.OpenRepository(w.cloneDir); err != nil {
		err = errors.WithMessagev(err, "unable to open git repository", w.cloneDir)
		return
	}

	if err = repository.FetchAll(w.repoInfo.RemoteURL); err != nil {
		err = errors.Wrapv(err, "unable to fetch all from repository", w.cloneDir)
	}

	return
}

func (w *cloneWorker) finish(bar *progress.Spinner, err error) {
	if gitpkg.IsErrEmptyRemoteRepository(err) {
		w.finishBar(bar, "- %s repo empty, skipping search")
		return
	}

	if err != nil {
		errors.ErrLog(w.log, err).Error(err.Error())
		w.finishBar(bar, "- %s repo prep error, skipping search: "+err.Error())
		return
	}

	w.finishBar(bar, "- %s repo prepared for search")
}

func (w *cloneWorker) finishBar(bar *progress.Spinner, doneMessage string) {
	if bar == nil {
		return
	}

	bar.Incr()
	w.prog.Add(0, mpb.BarFillerFunc(func(writer io.Writer, width int, st *decor.Statistics) {
		_, _ = fmt.Fprintf(writer, doneMessage, w.repoInfo.Name)
	})).SetTotal(0, true)
}
