package scm_clients

import (
	"context"
)

type FileListingEntry struct {
	Type *string
	Name *string
	Path *string
}

type ScmClient interface {
	ChangedFilesInPullRequest(ctx context.Context, pullRequestID int) ([]string, error)
	ChangedFilesInDiff(ctx context.Context, base string, head string) ([]string, error)
	GetFileContents(ctx context.Context, path string, afterRef string) (
		fileContent string, err error)
	GetContents(ctx context.Context, path string, afterRef string) (
		fileListing []FileListingEntry, err error)
}
