package plugin

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/config"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func New(server, token string) config.Plugin {
	return &plugin{
		server: server,
		token:  token,
	}
}

type plugin struct {
	server string
	token  string
}

func (p *plugin) Find(ctx context.Context, req *config.Request) (*drone.Config, error) {
	// log
	logrus.Infof("Handling %s %s: %s to %s", req.Repo.Namespace, req.Repo.Name, req.Build.Before, req.Build.After)

	// github client
	client := &github.Client{}

	// creates a github transport that authenticates
	// http requests using the github access token.
	trans := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: p.token},
	))

	// connect with github
	if p.server == "" {
		client = github.NewClient(trans)
	} else {
		var err error
		client, err = github.NewEnterpriseClient(p.server, p.server, trans)
		if err != nil {
			logrus.Errorf("Unable to connect to Github: '%v', server: '%s'", err, p.server)
			return nil, err
		}
	}

	// get repo changes
	changes, _, err := client.Repositories.CompareCommits(ctx, req.Repo.Namespace, req.Repo.Name, req.Build.Before, req.Build.After)
	if err != nil {
		logrus.Errorf("Unable to fetch diff: '%v', server: '%s'", err, p.server)
		return nil, err
	}

	// collect all directories with changes
	for _, file := range changes.Files {
		dir := *file.Filename
		if !strings.HasPrefix(dir, "/") {
			dir = "/" + dir
		}
		done := false
		for !done {
			done = bool(dir == "/")
			dir = path.Join(dir, "..")
			file := path.Join(dir, req.Repo.Config)

			// check file on github
			content, err := p.getGithubFile(ctx, req, client, file)
			if err != nil {
				logrus.Infof("Unable to load file: %s %v", file, err)
			} else {
				logrus.Infof("Found %s/%s %s", req.Repo.Namespace, req.Repo.Name, file)
				return &drone.Config{Data: content}, nil
			}
		}
	}

	// no file found
	return nil, errors.New("Did not found a .drone.yml")
}

// get the contents of a file on github, if the file is not found throw an error
func (p *plugin) getGithubFile(ctx context.Context, req *config.Request, client *github.Client, file string) (string, error) {
	logrus.Infof("Testing %s/%s %s", req.Repo.Namespace, req.Repo.Name, file)

	ref := github.RepositoryContentGetOptions{Ref: req.Build.After}
	data, _, _, err := client.Repositories.GetContents(ctx, req.Repo.Namespace, req.Repo.Name, file, &ref)
	if err != nil {
		return "", err
	}

	content, err := data.GetContent()
	if err != nil {
		return "", err
	}

	return content, nil
}
