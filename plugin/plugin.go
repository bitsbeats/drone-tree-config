package plugin

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/bitsbeats/drone-tree-config/plugin/scm_clients"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/config"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type (
	Plugin struct {
		server              string
		gitHubToken         string
		gitLabToken         string
		gitLabServer        string
		bitBucketAuthServer string
		bitBucketClient     string
		bitBucketSecret     string

		concat        bool
		fallback      bool
		maxDepth      int
		whitelistFile string
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

// New creates a drone plugin
func New(options ...func(*Plugin)) config.Plugin {
	p := &Plugin{}
	for _, opt := range options {
		opt(p)
	}

	return p
}

// Find is called by drone
func (p *Plugin) Find(ctx context.Context, droneRequest *config.Request) (*drone.Config, error) {
	someUuid := uuid.New()
	logrus.Infof("%s %s/%s started", someUuid, droneRequest.Repo.Namespace, droneRequest.Repo.Name)
	defer logrus.Infof("%s finished", someUuid)

	// connect to scm
	client, err := p.NewScmClient(ctx, someUuid, droneRequest.Repo)
	if err != nil {
		return nil, err
	}

	req := request{droneRequest, someUuid, client}

	// make sure this plugin is enabled for the requested repo slug
	if ok := p.whitelisted(&req); !ok {
		// do the default behavior by returning nil, nil
		return nil, nil
	}

	// get changed files
	changedFiles, err := p.getScmChanges(ctx, &req)
	if err != nil {
		return nil, err
	}

	// get drone.yml for changed files or all of them if no changes/cron
	configData := ""
	if changedFiles != nil {
		configData, err = p.getConfigForChanges(ctx, &req, changedFiles)
	} else if req.Build.Trigger == "@cron" {
		logrus.Warnf("%s @cron, rebuilding all", req.UUID)
		configData, err = p.getConfigForTree(ctx, &req, "", 0)
	} else if p.fallback {
		logrus.Warnf("%s no changed files and fallback enabled, rebuilding all", req.UUID)
		configData, err = p.getConfigForTree(ctx, &req, "", 0)
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

var dedupRegex = regexp.MustCompile(`(?ms)(---[\s]*){2,}`)

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
