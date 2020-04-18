package code

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/sirupsen/logrus"
    "github.com/vbauerster/mpb/v5"
    "github.com/vbauerster/mpb/v5/decor"
    "io"
    "os"
)

type worker struct {
    repoInfo *RepoInfo
    cloneDir string
    git      *gitpkg.Git
    prog     *progress.Progress
    db       *database.Database
    log      *logrus.Entry
}

func (w worker) Perform() {
    if err := w.prepareRepo(w.repoInfo); err != nil {
        errors.ErrorLogForEntry(w.log, errors.WithMessagev(err, "unable to prepare repo", w.repoInfo.Name))
    }
}

func (w worker) prepareRepo(ghRepo *RepoInfo) (err error) {
    var bar *progress.Spinner
    if w.prog != nil {
        bar = w.prog.AddSpinner(w.repoInfo.Name)
    }

    defer func() {
        if w.prog != nil {
            bar.Incr()
            w.prog.Add(0, mpb.BarFillerFunc(func(writer io.Writer, width int, st *decor.Statistics) {
                fmt.Fprintf(writer, "- source of %s is prepared ", w.repoInfo.Name)
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
        err = errors.Wrapv(err, "unable to remove corrupt repo, skipping")
        return
    }

    // Clone
    if _, statErr := os.Stat(w.cloneDir); os.IsNotExist(statErr) {
        w.log.Debug("cloning repo")

        if err = w.cloneRepo(ghRepo.SSHURL, cloneDirTmp, w.log); err != nil {
            err = errors.Wrapv(err, "unable to clone repo, skipping", ghRepo.SSHURL, cloneDirTmp)
            return
        }
        if err = os.Rename(cloneDirTmp, w.cloneDir); err != nil {
            err = errors.Wrapv(err, "unable to rename temporary clone dir repo, skipping", cloneDirTmp, w.cloneDir)
            return
        }
        if _, statErr := os.Stat(w.cloneDir); os.IsNotExist(statErr) {
            w.log.Debug("cloning repo")
            err = errors.Wrapv(err, "repo clone failed, skipping", ghRepo.SSHURL, cloneDirTmp)
            return
        }
    }

    err = w.db.WriteRepo(&database.Repo{
        ID:       database.CreateHashID(ghRepo.FullName),
        Name:     ghRepo.Name,
        FullName: ghRepo.FullName,
        Owner:    ghRepo.Owner,
        SSHURL:   ghRepo.SSHURL,
        HTMLURL:  ghRepo.HTMLURL,
        CloneDir: w.cloneDir,
    })
    if err != nil {
        err = errors.WithMessagev(err, "unable to write repo, skipping", ghRepo.Name)
        return
    }

    w.log.Debug("repo is prepared")
    return
}

func (w worker) removeExistingCorruptClone() (err error) {
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

func (w worker) cloneRepo(url, cloneDir string, log *logrus.Entry) (err error) {
    if _, err = w.git.Clone(url, cloneDir); err != nil {
        if gitpkg.IsErrEmptyRemoteRepository(err) {
            log.Warn("clone failed because remote repo has no commits, skipping")
        }
        _ = os.RemoveAll(cloneDir)
    }

    return
}
