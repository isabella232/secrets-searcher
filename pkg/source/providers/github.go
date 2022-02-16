package providers

import (
	"context"
	"fmt"

	"github.com/google/go-github/v29/github"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/pantheon-systems/secrets-searcher/pkg/source"
)

const perPage = 100

type GithubProvider struct {
	name         string
	user         string
	organization string
	client       *github.Client
	skipForks    bool
	log          logg.Logg
}

func NewGithubProvider(name, organization string, client *github.Client, skipForks bool, log logg.Logg) *GithubProvider {
	return &GithubProvider{
		name:         name,
		user:         "", // TODO
		organization: organization,
		client:       client,
		skipForks:    skipForks,
		log:          log,
	}
}

func (p *GithubProvider) GetName() (result string) {
	return p.name
}

func (p *GithubProvider) GetRepositories(repoFilter *manip.SliceFilter) (result []*source.RepoInfo, err error) {
	var ghRepos []*github.Repository
	ghRepos, err = p.QueryReposByOrg(p.organization)
	if err != nil {
		err = errors.WithMessagev(err, "unable to get repositories", p.organization)
		return
	}

	for _, ghRepo := range ghRepos {
		if !repoFilter.Includes(ghRepo.GetName()) {
			continue
		}
		if p.skipForks && ghRepo.GetFork() {
			continue
		}

		result = append(result, &source.RepoInfo{
			Name:      ghRepo.GetName(),
			RemoteURL: ghRepo.GetSSHURL(),
		})
	}

	return
}

func (p *GithubProvider) GetRepoURL(repoName string) (result string) {
	return fmt.Sprintf("https://github.com/%s/%s", p.Owner(), repoName)
}

func (p *GithubProvider) GetCommitURL(repoName, commitHash string) (result string) {
	return fmt.Sprintf("%s/commit/%s", p.GetRepoURL(repoName), commitHash)
}

func (p *GithubProvider) GetFileURL(repoName, commitHash, filePath string) (result string) {
	return fmt.Sprintf("%s/blob/%s/%s", p.GetRepoURL(repoName), commitHash, filePath)
}

func (p *GithubProvider) GetFileLineURL(repoName, commitHash, filePath string, startLineNum, endLineNum int) (result string) {
	var lineSpecifier string
	if startLineNum == endLineNum {
		lineSpecifier = fmt.Sprintf("L%d", startLineNum)
	} else {
		lineSpecifier = fmt.Sprintf("L%d-L%d", startLineNum, endLineNum)
	}
	return fmt.Sprintf("%s/blob/%s/%s#%s", p.GetRepoURL(repoName), commitHash, filePath, lineSpecifier)
}

func (p *GithubProvider) Owner() string {
	if p.organization != "" {
		return p.organization
	}
	return p.user
}

func (p *GithubProvider) QueryReposByOrg(organization string) (result []*github.Repository, err error) {
	ctx := context.Background()
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: perPage},
	}
	for {
		var repos []*github.Repository
		var resp *github.Response
		repos, resp, err = p.client.Repositories.ListByOrg(ctx, organization, opt)
		if err != nil {
			err = errors.WithMessagef(err, "unable to get %d repos from GitHub (page %d)", perPage, opt.Page)
			return
		}
		result = append(result, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return
}

func (p *GithubProvider) GetChange(owner, repo, sha string) (result *github.RepositoryCommit, err error) {
	ctx := context.Background()
	result, _, err = p.client.Repositories.GetCommit(ctx, owner, repo, sha)
	return
}
