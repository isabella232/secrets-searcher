package provider

import (
    "context"
    "github.com/google/go-github/v29/github"
    "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "golang.org/x/oauth2"
)

const perPage = 100

type GithubProvider struct {
    name         string
    client       *github.Client
    organization string
    repoFilter   *structures.Filter
    excludeForks bool
    log          logrus.FieldLogger
}

func NewGithubProvider(name, apiToken, organization string, repoFilter *structures.Filter, excludeForks bool, log logrus.FieldLogger) *GithubProvider {
    ctx := context.Background()
    tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiToken}))
    client := github.NewClient(tc)

    return &GithubProvider{
        name:         name,
        client:       client,
        organization: organization,
        repoFilter:   repoFilter,
        excludeForks: excludeForks,
        log:          log,
    }
}

func (p *GithubProvider) GetName() (result string) {
    return p.name
}

func (p *GithubProvider) GetRepositories() (result []*code.RepoInfo, err error) {
    var ghRepos []*github.Repository
    ghRepos, err = p.QueryReposByOrg(p.organization)
    if err != nil {
        err = errors.WithMessagev(err, "unable to get repositories", p.organization)
        return
    }

    for _, ghRepo := range ghRepos {
        if !p.repoFilter.IsIncluded(ghRepo.GetName()) {
            p.log.Debug("repo excluded using filter, skipping")
            continue
        }
        if p.excludeForks && ghRepo.GetFork() {
            p.log.Debug("repo is a fork, skipping")
            continue
        }

        result = append(result, &code.RepoInfo{
            Name:     ghRepo.GetName(),
            FullName: ghRepo.GetFullName(),
            Owner:    ghRepo.GetOwner().GetLogin(),
            SSHURL:   ghRepo.GetSSHURL(),
            HTMLURL:  ghRepo.GetHTMLURL(),
        })
    }

    return
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
