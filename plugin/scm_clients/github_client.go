package scm_clients

import (
	"context"
	"fmt"
	"sync"

	"github.com/drone/drone-go/drone"
	"github.com/google/go-github/v33/github"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	delegate *github.Client
	repo     drone.Repo
}

var (
	lock     sync.Mutex
	ghClient *github.Client
)

// NewGitHubClient creates a GithubClient which can be used to send requests to the Github API
func NewGitHubClient(ctx context.Context, uuid uuid.UUID, server string, token string, repo drone.Repo) (ScmClient, error) {
	client, err := getClientDelegate(ctx, server, token)
	if err != nil {
		logrus.Errorf("%s Unable to connect to Github: '%v'", uuid, err)
		return nil, err
	}

	return GithubClient{
		delegate: client,
		repo:     repo,
	}, nil
}

func getClientDelegate(ctx context.Context, server string, token string) (*github.Client, error) {
	lock.Lock()
	defer lock.Unlock()

	// return pre-existing client delegate
	if ghClient != nil {
		return ghClient, nil
	}

	// create a new one
	trans := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	if server == "" {
		ghClient = github.NewClient(trans)
	} else {
		var err error
		ghClient, err = github.NewEnterpriseClient(server, server, trans)
		if err != nil {
			return nil, err
		}
	}

	return ghClient, nil
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
	f, resp, err := s.delegate.PullRequests.ListFiles(ctx, s.repo.Namespace, s.repo.Name, number, opts)
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		logrus.Debugf("PullRequest.ListFiles %d: %s", resp.StatusCode, resp.Request.URL)
	} else {
		logrus.Debugf("PullRequest.ListFiles <nil> response encountered, err: %s", err.Error())
	}
	return f, resp, err
}

func (s GithubClient) compareCommits(ctx context.Context, base, head string) (
	*github.CommitsComparison, *github.Response, error) {
	c, resp, err := s.delegate.Repositories.CompareCommits(ctx, s.repo.Namespace, s.repo.Name, base, head)
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		logrus.Debugf("PullRequest.CompareCommits %d: %s", resp.StatusCode, resp.Request.URL)
	} else {
		logrus.Debugf("PullRequest.CompareCommits <nil> response encountered, err: %s", err.Error())
	}
	return c, resp, err
}

func (s GithubClient) getContents(ctx context.Context, path string, commitRef string) (
	*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	opts := &github.RepositoryContentGetOptions{Ref: commitRef}
	f, d, resp, err := s.delegate.Repositories.GetContents(ctx, s.repo.Namespace, s.repo.Name, path, opts)
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		logrus.Debugf("PullRequest.GetContents %d: %s", resp.StatusCode, resp.Request.URL)
	} else {
		logrus.Debugf("PullRequest.GetContents <nil> response encountered, err: %s", err.Error())
	}
	return f, d, resp, err
}
