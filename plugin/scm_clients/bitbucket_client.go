package scm_clients

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/drone/drone-go/drone"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/wbrefvem/go-bitbucket"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

type BitBucketClient struct {
	delegate *bitbucket.APIClient
	repo     drone.Repo
}

type bitBucketCredentials struct {
	accessToken string
}

func NewBitBucketClient(uuid uuid.UUID, server string,
	clientID string, clientSecret string, repo drone.Repo) (ScmClient, error) {
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://bitbucket.org/site/oauth2/access_token",
		Scopes:       []string{},
	}
	oauthClient := config.Client(context.Background())
	response, err := oauthClient.Get("https://www.fakeapiprovider.com/endpoint")
	if err != nil {
		return nil, err
	}
	var creds bitBucketCredentials
	if err = json.NewDecoder(response.Body).Decode(&creds); err != nil {
		return nil, err
	}

	conf := bitbucket.NewConfiguration()
	conf.Host = server
	conf.Scheme = "https"
	conf.AddDefaultHeader("Authorization", "Bearer "+creds.accessToken)

	client := bitbucket.NewAPIClient(conf)
	logrus.Debugf("%s Connected to BitBucket.", uuid)

	return BitBucketClient{
		delegate: client,
		repo:     repo,
	}, nil
}

func (s BitBucketClient) ChangedFilesInPullRequest(ctx context.Context, pullRequestID int) ([]string, error) {
	var changedFiles []string
	response, err := s.delegate.PullrequestsApi.RepositoriesUsernameRepoSlugPullrequestsPullRequestIdDiffstatGet(
		ctx, s.repo.Namespace, string(pullRequestID), s.repo.Name)

	if err != nil {
		return nil, err
	}
	var diffStat bitbucket.PaginatedDiffstats

	if err = json.NewDecoder(response.Body).Decode(&diffStat); err != nil {
		return nil, err
	}
	for _, fileDiff := range diffStat.Values {
		changedFiles = append(changedFiles, fileDiff.New.Path)
	}
	return changedFiles, nil
}

func (s BitBucketClient) ChangedFilesInDiff(ctx context.Context, base string, head string) ([]string, error) {
	var changedFiles []string
	spec := fmt.Sprintf("%s..%s", base, head)
	diffStat, _, err := s.delegate.DefaultApi.RepositoriesUsernameRepoSlugDiffstatSpecGet(
		ctx, s.repo.Namespace, s.repo.Name, spec, make(map[string]interface{}))
	if err != nil {
		return nil, err
	}
	for _, fileDiff := range diffStat.Values {
		changedFiles = append(changedFiles, fileDiff.New.Path)
	}
	return changedFiles, nil
}

func (s BitBucketClient) GetFileContents(ctx context.Context, path string, commitRef string) (content string, err error) {
	_, response, err := s.delegate.RepositoriesApi.RepositoriesUsernameRepoSlugSrcNodePathGet(
		ctx, s.repo.Namespace, s.repo.Name, commitRef, path, make(map[string]interface{}))

	if response == nil {
		return "", fmt.Errorf("failed to get %s: is not a file", path)
	}
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get %s: status code %s", path, response.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	return bodyString, nil
}

func (s BitBucketClient) GetContents(ctx context.Context, path string, commitRef string) (
	fileListing []FileListingEntry, err error) {
	opts := make(map[string]interface{})
	opts["format"] = "meta"
	ls, _, err := s.delegate.RepositoriesApi.RepositoriesUsernameRepoSlugSrcNodePathGet(
		ctx, s.repo.Namespace, s.repo.Name, commitRef, path, opts)

	var result []FileListingEntry

	if err != nil {
		return result, err
	}

	for _, f := range ls.Values {
		var fileType string
		if f.Type_ == "commit_file" {
			fileType = "file"
		} else if f.Type_ == "commit_directory" {
			fileType = "dir"
		} else {
			continue
		}
		fileName := filepath.Base(f.Path)
		fileListingEntry := FileListingEntry{
			Path: &f.Path,
			Name: &fileName,
			Type: &fileType,
		}
		result = append(result, fileListingEntry)
	}
	return result, err
}
