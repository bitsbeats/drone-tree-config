package plugin

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitsbeats/drone-tree-config/plugin/scm_clients"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/config"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// New creates a drone plugin
func New(server, gitHubToken string, bitBucketClient string, bitBucketSecret string,
	concat bool, fallback bool, maxDepth int) config.Plugin {
	return &Plugin{
		server:          server,
		gitHubToken:     gitHubToken,
		bitBucketClient: bitBucketClient,
		bitBucketSecret: bitBucketSecret,
		concat:          concat,
		fallback:        fallback,
		maxDepth:        maxDepth,
	}
}

type (
	Plugin struct {
		server          string
		gitHubToken     string
		bitBucketClient string
		bitBucketSecret string
		concat          bool
		fallback        bool
		maxDepth        int
	}

	droneConfig struct {
		Name string `yaml:"name"`
		Kind string `yaml:"kind"`
	}

	request struct {
		*config.Request
		UUID   uuid.UUID
		Client scm_clients.ScmClient
	}
)

var dedupRegex = regexp.MustCompile(`(?ms)(---[\s]*){2,}`)

func (p *Plugin) NewScmClient(uuid uuid.UUID, repo drone.Repo, ctx context.Context) scm_clients.ScmClient {
	var scmClient scm_clients.ScmClient
	var err error
	if p.gitHubToken != "" {
		scmClient, err = scm_clients.NewGitHubClient(uuid, p.server, p.gitHubToken, repo, ctx)
	} else if p.bitBucketClient != "" {
		scmClient, err = scm_clients.NewBitBucketClient(uuid, p.server, p.bitBucketClient, p.bitBucketSecret, repo)
	} else {
		err = fmt.Errorf("no SCM credentials specified")
	}
	if err != nil {
		logrus.Errorf("Unable to connect to SCM server.")
	}
	return scmClient
}

// Find is called by drone
func (p *Plugin) Find(ctx context.Context, droneRequest *config.Request) (*drone.Config, error) {
	someUuid := uuid.New()
	logrus.Infof("%s %s/%s started", someUuid, droneRequest.Repo.Namespace, droneRequest.Repo.Name)
	defer logrus.Infof("%s finished", someUuid)

	// connect to github
	client := p.NewScmClient(someUuid, droneRequest.Repo, ctx)

	req := request{droneRequest, someUuid, client}

	// get changed files
	changedFiles, err := p.getGithubChanges(ctx, &req)
	if err != nil {
		return nil, err
	}

	// get drone.yml for changed files or all of them if no changes/cron
	configData := ""
	if changedFiles != nil {
		configData, err = p.getGithubConfigData(ctx, &req, changedFiles)
	} else if req.Build.Trigger == "@cron" {
		logrus.Warnf("%s @cron, rebuilding all", req.UUID)
		configData, err = p.getAllConfigData(ctx, &req, "/", 0)
	} else if p.fallback {
		logrus.Warnf("%s no changed files and fallback enabled, rebuilding all", req.UUID)
		configData, err = p.getAllConfigData(ctx, &req, "/", 0)
	}
	if err != nil {
		return nil, err
	}

	// no file found
	if configData == "" {
		return nil, errors.New("did not find a .drone.yml")
	}

	// cleanup
	configData = strings.ReplaceAll(configData, "...", "")
	configData = string(dedupRegex.ReplaceAll([]byte(configData), []byte("---")))

	return &drone.Config{Data: configData}, nil
}

// getGithubChanges tries to get a list of changed files from github
func (p *Plugin) getGithubChanges(ctx context.Context, req *request) ([]string, error) {
	var changedFiles []string

	if req.Build.Trigger == "@cron" {
		// cron jobs trigger a full build
		changedFiles = []string{}
	} else if strings.HasPrefix(req.Build.Ref, "refs/pull/") {
		// use pullrequests api to get changed files
		pullRequestID, err := strconv.Atoi(strings.Split(req.Build.Ref, "/")[2])
		if err != nil {
			logrus.Errorf("%s unable to get pull request id %v", req.UUID, err)
			return nil, err
		}
		changedFiles, err = req.Client.ChangedFilesInPullRequest(ctx, pullRequestID)
		if err != nil {
			logrus.Errorf("%s unable to fetch diff for Pull request %v", req.UUID, err)
		}
	} else {
		// use diff to get changed files
		before := req.Build.Before
		if before == "0000000000000000000000000000000000000000" || before == "" {
			before = fmt.Sprintf("%s~1", req.Build.After)
		}
		var err error
		changedFiles, err = req.Client.ChangedFilesInDiff(ctx, before, req.Build.After)
		if err != nil {
			logrus.Errorf("%s unable to fetch diff: '%v'", req.UUID, err)
			return nil, err
		}
	}

	if len(changedFiles) > 0 {
		changedList := strings.Join(changedFiles, "\n  ")
		logrus.Debugf("%s changed files: \n  %s", req.UUID, changedList)
	} else {
		return nil, nil
	}
	return changedFiles, nil
}

// getGithubFile downloads a file from github
func (p *Plugin) getGithubFile(ctx context.Context, req *request, file string) (content string, err error) {
	logrus.Debugf("%s checking %s/%s %s", req.UUID, req.Repo.Namespace, req.Repo.Name, file)
	return req.Client.GetFileContents(ctx, file, req.Build.After)
}

// getGithubDroneConfig downloads a drone config and validates it
func (p *Plugin) getGithubDroneConfig(ctx context.Context, req *request, file string) (configData string, critical bool, err error) {
	fileContent, err := p.getGithubFile(ctx, req, file)
	if err != nil {
		logrus.Debugf("%s skipping: unable to load file: %s %v", req.UUID, file, err)
		return "", false, err
	}

	// validate fileContent, exit early if an error was found
	dc := droneConfig{}
	err = yaml.Unmarshal([]byte(fileContent), &dc)
	if err != nil {
		logrus.Errorf("%s skipping: unable do parse yml file: %s %v", req.UUID, file, err)
		return "", true, err
	}
	if dc.Name == "" || dc.Kind == "" {
		logrus.Errorf("%s skipping: missing 'kind' or 'name' in %s.", req.UUID, file)
		return "", true, err
	}

	logrus.Infof("%s found %s/%s %s", req.UUID, req.Repo.Namespace, req.Repo.Name, file)
	return fileContent, false, nil
}

// getGithubConfigData scans a repository based on the changed files
func (p *Plugin) getGithubConfigData(ctx context.Context, req *request, changedFiles []string) (configData string, err error) {
	// collect drone.yml files
	configData = ""
	cache := map[string]bool{}
	for _, file := range changedFiles {
		if !strings.HasPrefix(file, "/") {
			file = "/" + file
		}

		done := false
		dir := file
		for !done {
			done = bool(dir == "/")
			dir = path.Join(dir, "..")
			file := path.Join(dir, req.Repo.Config)

			// check if file has already been checked
			_, ok := cache[file]
			if ok {
				continue
			} else {
				cache[file] = true
			}

			// download file from git
			fileContent, critical, err := p.getGithubDroneConfig(ctx, req, file)
			if err != nil {
				if critical {
					return "", err
				}
				continue
			}

			// append
			configData = p.droneConfigAppend(configData, fileContent)
			if !p.concat {
				logrus.Infof("%s concat is disabled. Using just first .drone.yml.", req.UUID)
				break
			}
		}
	}
	return configData, nil
}

// getAllConfigData searches for all or fist 'drone.yml' in the repo
func (p *Plugin) getAllConfigData(ctx context.Context, req *request, dir string, depth int) (configData string, err error) {
	ls, err := req.Client.GetContents(ctx, dir, req.Build.After)
	if err != nil {
		return "", err
	}

	if depth > p.maxDepth {
		logrus.Infof("%s skipping scan of %s, max depth %d reached.", req.UUID, dir, depth)
		return "", nil
	}
	depth += 1

	// check recursivly for drone.yml
	configData = ""
	for _, f := range ls {
		var fileContent string
		if *f.Type == "dir" {
			fileContent, _ = p.getAllConfigData(ctx, req, *f.Path, depth)
		} else if *f.Type == "file" && *f.Name == req.Repo.Config {
			var critical bool
			fileContent, critical, err = p.getGithubDroneConfig(ctx, req, *f.Path)
			if critical {
				return "", err
			}
		}
		// append
		configData = p.droneConfigAppend(configData, fileContent)
		if !p.concat {
			logrus.Infof("%s concat is disabled. Using just first .drone.yml.", req.UUID)
			break
		}
	}

	return configData, nil
}

// droneConfigAppend concats multiple 'drone.yml's to a multi-machine pipeline
// see https://docs.drone.io/user-guide/pipeline/multi-machine/
func (p *Plugin) droneConfigAppend(droneConfig string, appends ...string) string {
	for _, a := range appends {
		a = strings.Trim(a, " \n")
		if a != "" {
			if !strings.HasPrefix(a, "---\n") {
				a = "---\n" + a
			}
			droneConfig += a
			if !strings.HasSuffix(droneConfig, "\n") {
				droneConfig += "\n"
			}
		}
	}
	return droneConfig
}
