package scm_clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/drone/drone-go/drone"
	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"strings"
)

type BitBucketServerClient struct {
	delegate *bitbucketv1.APIClient
	token	string
	server  string
	repo 	drone.Repo
}

type bitbucketDiffs struct {
	Diffs []struct {
		Destination struct {
			Name string `json:"toString"`
		} `json:"destination"`
	} `json:"diffs"`
}

type FileListing struct {
	FileNames []string `json:"values"`
}

func NewBitBucketServerClient(uuid uuid.UUID,
	server string, token string, repo drone.Repo) (ScmClient, error){

	// Create correct context for authentication
	var ctx context.Context
	ctx = context.WithValue(context.Background(), bitbucketv1.ContextAccessToken, token)

	configuration := bitbucketv1.NewConfiguration(server)
	client := bitbucketv1.NewAPIClient(ctx, configuration)

	return BitBucketServerClient{
		delegate: client,
		repo:     repo,
		token:	  token,
		server:   server,
	}, nil
}

func (b BitBucketServerClient) ChangedFilesInPullRequest(ctx context.Context, pullRequestID int) ([]string, error){
	var changedFiles []string
	client := b.delegate.DefaultApi

	ff , err := client.GetPullRequestDiff(b.repo.Namespace, b.repo.Name, pullRequestID, map[string]interface{}{})
	if err != nil {
		return nil ,err
	}
	jsonString, err := json.Marshal(ff.Values)
	if err != nil {
		return nil, err
	}
	res := bitbucketDiffs{}
	err = json.Unmarshal(jsonString, &res)
	if err != nil {
		return nil, err
	}

	for _, diff := range res.Diffs {
		changedFiles = append(changedFiles, diff.Destination.Name)
	}
	return changedFiles, nil
}

func (b BitBucketServerClient) ChangedFilesInDiff(ctx context.Context, base string, head string) ([]string, error) {
	var changedFiles []string
	client := b.delegate.DefaultApi
	params := map[string]interface{}{
		"since": base,
	}
	ff, err := client.StreamDiff(b.repo.Namespace, b.repo.Name, head, params)
	if err != nil {
		return nil, err
	}

	jsonString, err := json.Marshal(ff.Values)
	if err != nil {
		return nil, err
	}
	res := bitbucketDiffs{}
	err = json.Unmarshal(jsonString, &res)
	if err != nil {
		return nil, err
	}

	for _, diff := range res.Diffs {
		changedFiles = append(changedFiles, diff.Destination.Name)
	}
	return changedFiles, nil
}


func (b BitBucketServerClient) GetFileContents(ctx context.Context, path string, commitRef string) (fileContent string, err error) {
	client := b.delegate.DefaultApi
	params := map[string]interface{}{
		"at": commitRef,
	}
	contents, err := client.GetContent_11(b.repo.Namespace, b.repo.Name, path, params)
	if err != nil {
		return "", nil
	}
	fileContents, err := unmarshal(contents.Payload)
	if err != nil {
		return "", err
	}
	return fileContents, nil
}


func (b BitBucketServerClient) GetFileListing(ctx context.Context, path string, commitRef string) (fileListing []FileListingEntry, err error) {
	client := b.delegate.DefaultApi
	params := map[string]interface{}{
		"at": commitRef,
	}
	ls, err := client.StreamFiles_42(b.repo.Namespace, b.repo.Name, path, params)
	if err != nil {
		return nil, err
	}

	var result []FileListingEntry

	jsonString, err := json.Marshal(ls.Values)
	if err != nil {
		return nil, err
	}

	files := FileListing{}
	err = json.Unmarshal(jsonString, &files)
	if err != nil {
		return nil, err
	}

	for _, f := range files.FileNames {
		fileType := "dir"
		filePath := fmt.Sprintf("%v", f)
		fileSlice := strings.Split(filePath, "/")
		fileName := fileSlice[len(fileSlice)-1]
		if strings.Contains(fileName, ".") {
			fileType = "file"
		}
		fileListingEntry := FileListingEntry{
			Path: filePath,
			Name: fileName,
			Type: fileType,
		}
		result = append(result, fileListingEntry)
	}
	return result, err
}

func unmarshal(b []byte) (string, error) {
	var stringArray []string
	buf := bytes.NewBuffer(b)
	dec := yaml.NewDecoder(buf)
	var obj map[string]interface{}
	for {
		err := dec.Decode(&obj)
		if err != nil {
			break
		}
		a,_ := yaml.Marshal(obj)
		stringArray = append(stringArray, string(a))
	}
	var yamlString string
	for n, ass := range stringArray {
		if n == 0 {
			yamlString = ass
			continue
		}
		yamlString = yamlString + "\n---\n" + ass
	}
	return yamlString, nil
}
