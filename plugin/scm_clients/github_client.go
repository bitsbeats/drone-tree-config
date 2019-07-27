package scm_clients

import (
	"context"
	"github.com/drone/drone-go/drone"
	"github.com/google/go-github/github"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	delegate *github.Client
	repo     drone.Repo
}

func GitHubClient(uuid uuid.UUID, server string, token string, repo drone.Repo, ctx context.Context) (ScmClient, error) {
	trans := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	var client *github.Client
	if server == "" {
		client = github.NewClient(trans)
	} else {
		var err error
		client, err = github.NewEnterpriseClient(server, server, trans)
		if err != nil {
			logrus.Errorf("%s Unable to connect to Github: '%v'", uuid, err)
			return nil, err
		}
	}
	return GithubClient{
		delegate: client,
		repo:     repo,
	}, nil
}

func (s GithubClient) ListFiles(ctx context.Context, number int) (
	[]*github.CommitFile, *github.Response, error) {
	opts := &github.ListOptions{}
	return s.delegate.PullRequests.ListFiles(ctx, s.repo.Namespace, s.repo.Name, number, opts)
}

func (s GithubClient) CompareCommits(ctx context.Context, base, head string) (
	*github.CommitsComparison, *github.Response, error) {
	return s.delegate.Repositories.CompareCommits(ctx, s.repo.Namespace, s.repo.Name, base, head)
}

func (s GithubClient) GetContents(ctx context.Context, path string, afterRef string) (
	fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error) {
	opts := &github.RepositoryContentGetOptions{Ref: afterRef}
	return s.delegate.Repositories.GetContents(ctx, s.repo.Namespace, s.repo.Name, path, opts)
}
