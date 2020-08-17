package scm_clients

import (
	"context"
	"fmt"

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

func NewGitHubClient(ctx context.Context, uuid uuid.UUID, server string, token string, repo drone.Repo) (ScmClient, error) {
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

func (s GithubClient) ChangedFilesInPullRequest(ctx context.Context, pullRequestID int) ([]string, error) {
	var changedFiles []string
	opts := &github.ListOptions{}

	for {
		files, resp, err := s.listFiles(ctx, pullRequestID, opts)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			changedFiles = append(changedFiles, *file.Filename)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return changedFiles, nil
}

func (s GithubClient) ChangedFilesInDiff(ctx context.Context, base string, head string) ([]string, error) {
	var changedFiles []string
	changes, _, err := s.compareCommits(ctx, base, head)
	if err != nil {
		return nil, err
	}
	for _, file := range changes.Files {
		changedFiles = append(changedFiles, *file.Filename)
	}
	return changedFiles, nil
}

func (s GithubClient) GetFileContents(ctx context.Context, path string, commitRef string) (content string, err error) {
	data, _, _, err := s.getContents(ctx, path, commitRef)
	if data == nil {
		err = fmt.Errorf("failed to get %s: is not a file", path)
	}
	if err != nil {
		return "", err
	}
	return data.GetContent()
}

func (s GithubClient) GetFileListing(ctx context.Context, path string, commitRef string) (
	fileListing []FileListingEntry, err error) {
	_, ls, _, err := s.getContents(ctx, path, commitRef)
	var result []FileListingEntry

	if err != nil {
		return result, err
	}

	for _, f := range ls {
		fileListingEntry := FileListingEntry{
			Path: *f.Path,
			Name: *f.Name,
			Type: *f.Type,
		}
		result = append(result, fileListingEntry)
	}
	return result, err
}

func (s GithubClient) listFiles(ctx context.Context, number int, opts *github.ListOptions) (
	[]*github.CommitFile, *github.Response, error) {
	return s.delegate.PullRequests.ListFiles(ctx, s.repo.Namespace, s.repo.Name, number, opts)
}

func (s GithubClient) compareCommits(ctx context.Context, base, head string) (
	*github.CommitsComparison, *github.Response, error) {
	return s.delegate.Repositories.CompareCommits(ctx, s.repo.Namespace, s.repo.Name, base, head)
}

func (s GithubClient) getContents(ctx context.Context, path string, commitRef string) (
	fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error) {
	opts := &github.RepositoryContentGetOptions{Ref: commitRef}
	return s.delegate.Repositories.GetContents(ctx, s.repo.Namespace, s.repo.Name, path, opts)
}
