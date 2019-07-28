package scm_clients

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/drone/drone-go/drone"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/wbrefvem/go-bitbucket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type BitBucketClient struct {
	delegate      *bitbucket.APIClient
	basePath      string
	authorization string
	repo          drone.Repo
}

type BitBucketCredentials struct {
	AccessToken string `json:"access_token"`
}

func NewBitBucketClient(someUuid uuid.UUID, authServer string, server string,
	clientID string, clientSecret string, repo drone.Repo) (ScmClient, error) {

	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	req, err := http.NewRequest("POST", authServer+"/site/oauth2/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(clientID, clientSecret))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	var creds BitBucketCredentials
	if err = json.NewDecoder(response.Body).Decode(&creds); err != nil && creds.AccessToken != "" {
		return nil, err
	}
	logrus.Infof("%s Authenticated with BitBucket: '%v'", someUuid, authServer)

	authorization := "Bearer " + creds.AccessToken
	conf := bitbucket.NewConfiguration()
	conf.Host = server
	conf.Scheme = "https"
	conf.AddDefaultHeader("Authorization", authorization)

	basePath := server + "/2.0"
	client := bitbucket.NewAPIClient(conf)
	client.ChangeBasePath(basePath)
	logrus.Infof("%s Created BitBucket API client: '%v'", someUuid, server)

	return BitBucketClient{
		delegate:      client,
		basePath:      basePath,
		authorization: authorization,
		repo:          repo,
	}, nil
}

func (s BitBucketClient) ChangedFilesInPullRequest(ctx context.Context, pullRequestID int) ([]string, error) {
	var changedFiles []string
	// Custom implementation because the BitBucket client does not specify the right type
	requestUrl := fmt.Sprintf("%v/repositories/%v/%v/pullrequests/%v/diffstat",
		s.basePath, s.repo.Namespace, s.repo.Name, pullRequestID)
	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return []string{}, fmt.Errorf("failed to construct request for pull request %v", pullRequestID)
	}
	request.Header.Add("Authorization", s.authorization)
	response, err := http.DefaultClient.Do(request)

	if response == nil || err != nil {
		return []string{}, fmt.Errorf("failed to get %v: is not a pull request", pullRequestID)
	}
	var diffStat bitbucket.PaginatedDiffstats
	if err = json.NewDecoder(response.Body).Decode(&diffStat); err != nil {
		return nil, err
	}
	for _, fileDiff := range diffStat.Values {
		if fileDiff.Status == "removed" || fileDiff.Status == "renamed" {
			changedFiles = append(changedFiles, fileDiff.Old.Path)
		}
		if fileDiff.Status == "modified" || fileDiff.Status == "added" || fileDiff.Status == "renamed" {
			changedFiles = append(changedFiles, fileDiff.New.Path)
		}
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
	if diffStat.Values == nil {
		return nil, err
	}
	for _, fileDiff := range diffStat.Values {
		if fileDiff.Status == "removed" || fileDiff.Status == "renamed" {
			changedFiles = append(changedFiles, fileDiff.Old.Path)
		}
		if fileDiff.Status == "modified" || fileDiff.Status == "added" || fileDiff.Status == "renamed" {
			changedFiles = append(changedFiles, fileDiff.New.Path)
		}
	}
	return changedFiles, nil
}

func (s BitBucketClient) GetFileContents(ctx context.Context, path string, commitRef string) (content string, err error) {
	// Custom implementation because the BitBucket client always tries to deserialize the file as JSON
	requestUrl := fmt.Sprintf("%v/repositories/%v/%v/src/%v/%v",
		s.basePath, s.repo.Namespace, s.repo.Name, commitRef, path)
	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to construct request for %s", path)
	}
	request.Header.Add("Authorization", s.authorization)
	response, err := http.DefaultClient.Do(request)

	if response == nil || err != nil {
		return "", fmt.Errorf("failed to get %s: is not a file", path)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get %s: status code %v", path, response.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	return bodyString, nil
}

func (s BitBucketClient) GetFileListing(ctx context.Context, path string, commitRef string) (
	fileListing []FileListingEntry, err error) {
	opts := make(map[string]interface{})
	opts["format"] = "meta"
	ls, _, err := s.delegate.RepositoriesApi.RepositoriesUsernameRepoSlugSrcNodePathGet(
		ctx, s.repo.Namespace, commitRef, path+"/", s.repo.Name, opts)

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
			Path: f.Path,
			Name: fileName,
			Type: fileType,
		}
		result = append(result, fileListingEntry)
	}
	return result, err
}

func basicAuth(username string, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
