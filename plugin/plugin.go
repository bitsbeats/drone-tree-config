package plugin

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

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
		alwaysRunAll  bool
		maxDepth      int
		allowListFile string
		considerFile  string
		cacheTTL      time.Duration
		cache         *configCache
	}

	droneConfig struct {
		Name string `yaml:"name"`
		Kind string `yaml:"kind"`
	}

	request struct {
		*config.Request
		UUID         uuid.UUID
		Client       scm_clients.ScmClient
		ConsiderData *ConsiderData
	}
)

// New creates a drone plugin
func New(options ...func(*Plugin)) config.Plugin {
	p := &Plugin{
		cache: &configCache{},
	}
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

	req := request{
		Request: droneRequest,
		UUID:    someUuid,
		Client:  client,
	}

	// make sure this plugin is enabled for the requested repo slug
	if ok := p.allowlisted(&req); !ok {
		// do the default behavior by returning nil, nil
		return nil, nil
	}

	// avoid running for jsonnet or starlark configurations
	if !strings.HasSuffix(droneRequest.Repo.Config, ".yaml") && !strings.HasSuffix(droneRequest.Repo.Config, ".yml") {
		return nil, nil
	}

	// load the considerFile entries, if configured for considerFile
	if req.ConsiderData, err = p.newConsiderDataFromRequest(ctx, &req); err != nil {
		return nil, err
	}

	return p.getConfig(ctx, &req)
}

// getConfig retrieves drone config data. When the cache is enabled, this func will first check entries in
// the cache as well as add new entries.
func (p *Plugin) getConfig(ctx context.Context, req *request) (*drone.Config, error) {
	logrus.WithFields(logrus.Fields{
		"after":   req.Build.After,
		"before":  req.Build.Before,
		"branch":  req.Repo.Branch,
		"ref":     req.Build.Ref,
		"slug":    req.Repo.Slug,
		"trigger": req.Build.Trigger,
	}).Debugf("drone-tree-config environment")

	// check cache first, when enabled
	ck := newCacheKey(req)
	if p.cacheTTL > 0 {
		if cached, exists := p.cache.retrieve(req.UUID, ck); exists {
			if cached != nil {
				return &drone.Config{Data: cached.config}, cached.error
			}
		}
	}

	// fetch the config data. cache it, when enabled
	return p.cacheAndReturn(
		req.UUID, ck,
		newCacheEntry(
			p.getConfigData(ctx, req),
		),
	)
}

// getConfigData retrieves drone config data from the repo
func (p *Plugin) getConfigData(ctx context.Context, req *request) (string, error) {
	// get changed files
	changedFiles, err := p.getScmChanges(ctx, req)
	if err != nil {
		return "", err
	}

	// get drone.yml for changed files or all of them if no changes/cron
	configData := ""

	if p.alwaysRunAll {
		logrus.Warnf("%s always run all enabled, rebuilding all", req.UUID)
		if p.considerFile == "" {
			logrus.Warnf("recursively scanning for config files with max depth %d", p.maxDepth)
		}
		configData, err = p.getConfigForTree(ctx, req, "", 0)
	} else if changedFiles != nil {
		configData, err = p.getConfigForChanges(ctx, req, changedFiles)
	} else if req.Build.Trigger == "@cron" {
		logrus.Warnf("%s @cron, rebuilding all", req.UUID)
		if p.considerFile == "" {
			logrus.Warnf("recursively scanning for config files with max depth %d", p.maxDepth)
		}
		configData, err = p.getConfigForTree(ctx, req, "", 0)
	} else if p.fallback {
		logrus.Warnf("%s no changed files and fallback enabled, rebuilding all", req.UUID)
		if p.considerFile == "" {
			logrus.Warnf("recursively scanning for config files with max depth %d", p.maxDepth)
		}
		configData, err = p.getConfigForTree(ctx, req, "", 0)
	}
	if err != nil {
		return "", err
	}

	// no file found
	if configData == "" {
		return "", errors.New("did not find a .drone.yml")
	}

	// cleanup
	configData = removeDocEndRegex.ReplaceAllString(configData, "")
	configData = string(dedupRegex.ReplaceAll([]byte(configData), []byte("---")))
	return configData, nil
}

var dedupRegex = regexp.MustCompile(`(?ms)(---[\s]*){2,}`)
var removeDocEndRegex = regexp.MustCompile(`(?ms)^(\.\.\.)$`)

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
