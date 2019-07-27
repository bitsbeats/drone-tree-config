package scm_clients

import (
	"context"
	"github.com/google/go-github/github"
)

type ScmClient interface {
	ListFiles(ctx context.Context, number int) (
		[]*github.CommitFile, *github.Response, error)
	CompareCommits(ctx context.Context, base, head string) (
		*github.CommitsComparison, *github.Response, error)
	GetContents(ctx context.Context, path string, afterRef string) (
		fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
}
