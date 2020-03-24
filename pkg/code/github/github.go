package github

import (
    "context"
    "github.com/google/go-github/v29/github"
    "golang.org/x/oauth2"
)

type API struct {
    client *github.Client
}

func New(githubToken string) *API {
    ctx := context.Background()
    tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken}))
    client := github.NewClient(tc)

    return &API{
        client: client,
    }
}

func (ga *API) GetRepositoriesByOrganization(organization string) (repos []*github.Repository, err error) {
    ctx := context.Background()
    repos, _, err = ga.client.Repositories.ListByOrg(ctx, organization, nil)
    return
}
