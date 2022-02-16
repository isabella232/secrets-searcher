package providers

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	gitpkg "github.com/pantheon-systems/secrets-searcher/pkg/git"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/pantheon-systems/secrets-searcher/pkg/source"
)

type LocalProvider struct {
	name             string
	dir              string
	git              *gitpkg.Git
	metadataProvider source.ProviderI
	log              logg.Logg
}

func NewLocalProvider(name, dir string, git *gitpkg.Git, metadataProvider source.ProviderI, log logg.Logg) *LocalProvider {
	return &LocalProvider{
		name:             name,
		dir:              dir,
		git:              git,
		metadataProvider: metadataProvider,
		log:              log,
	}
}

func (p *LocalProvider) GetName() (result string) {
	return p.name
}

func (p *LocalProvider) GetRepositories(repoFilter *manip.SliceFilter) (result []*source.RepoInfo, err error) {
	var repoDirs []os.FileInfo
	repoDirs, err = ioutil.ReadDir(p.dir)
	if err != nil {
		err = errors.WithMessagev(err, "unable to get repositories", p.dir)
		return
	}

	for _, dir := range repoDirs {
		if !dir.IsDir() {
			continue
		}
		if dir.Name() == "libcloud" {
			print("found it")
		}
		if !repoFilter.Includes(dir.Name()) {
			continue
		}

		cloneDir := filepath.Join(p.dir, dir.Name())
		dotGitDir := filepath.Join(cloneDir, ".git")
		repoName := dir.Name()

		if err = p.git.ValidateClone(cloneDir); err != nil {
			err = errors.WithMessagev(err, "invalid repo", cloneDir)
			return
		}

		result = append(result, &source.RepoInfo{
			Name:      repoName,
			RemoteURL: dotGitDir,
		})
	}

	return
}

func (p *LocalProvider) GetRepoURL(repoName string) (result string) {
	if p.metadataProvider == nil {
		return
	}
	return p.metadataProvider.GetRepoURL(repoName)
}

func (p *LocalProvider) GetCommitURL(repoName, commitHash string) (result string) {
	if p.metadataProvider == nil {
		return
	}
	return p.metadataProvider.GetCommitURL(repoName, commitHash)
}

func (p *LocalProvider) GetFileURL(repoName, commitHash string, filePath string) (result string) {
	if p.metadataProvider == nil {
		return
	}
	return p.metadataProvider.GetFileURL(repoName, commitHash, filePath)
}

func (p *LocalProvider) GetFileLineURL(repoName, commitHash string, filePath string, startLineNum, endLineNum int) (result string) {
	if p.metadataProvider == nil {
		return
	}
	return p.metadataProvider.GetFileLineURL(repoName, commitHash, filePath, startLineNum, endLineNum)
}
