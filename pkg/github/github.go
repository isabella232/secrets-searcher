package github

import (
    "context"
    "github.com/google/go-github/v29/github"
    "golang.org/x/oauth2"
)

type API struct {
    client *github.Client
}

func NewAPI(githubToken string) *API {
    ctx := context.Background()
    tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken}))
    client := github.NewClient(tc)

    return &API{
        client: client,
    }
}

func (ga *API) GetRepositoriesByOrganization(organization string) (result []*github.Repository, err error) {
    ctx := context.Background()
    opt := &github.RepositoryListByOrgOptions{
        ListOptions: github.ListOptions{PerPage: 100},
    }
    for {
        var repos []*github.Repository
        var resp *github.Response
        repos, resp, err = ga.client.Repositories.ListByOrg(ctx, organization, opt)
        if err != nil {
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

func (ga *API) GetChange(owner, repo, sha string) (result *github.RepositoryCommit, err error) {
    ctx := context.Background()
    result, _, err = ga.client.Repositories.GetCommit(ctx, owner, repo, sha)
    return
}
