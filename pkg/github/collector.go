package github

import (
	"context"

	"github.com/google/go-github/v57/github"
	"github.com/stenic/go-git-backup/pkg/model"
	"golang.org/x/oauth2"
)

type GithubCollector struct {
	client *github.Client
}

func NewGithubCollector(ctx context.Context, token string) *GithubCollector {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GithubCollector{
		client: github.NewClient(tc),
	}
}

func (c GithubCollector) GetRepositories(ctx context.Context, org string) ([]model.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var allRepos2 []model.Repository

	for _, repo := range allRepos {
		allRepos2 = append(allRepos2, model.Repository{
			Name: *repo.Name,
			Slug: *repo.FullName,
			URL:  *repo.CloneURL,
		})
	}

	return allRepos2, nil
}
