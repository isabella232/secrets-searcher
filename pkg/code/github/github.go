package github

import (
	"context"
	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)

type API struct {
	client *github.Client
}

func New(githubToken string) (*API, error) {
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken}))
	client := github.NewClient(tc)

	return &API{
		client: client,
	}, nil
}

func (ga *API) GetRepositoriesByOrganization(organization string) ([]*github.Repository, error) {
	ctx := context.Background()
	repos, _, err := ga.client.Repositories.ListByOrg(ctx, organization, nil)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
