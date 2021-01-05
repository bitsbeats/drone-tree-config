package plugin

import (
	"context"
	"path"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// getConfigForChanges scans a repository based on the changed files
func (p *Plugin) getConfigForChanges(ctx context.Context, req *request, changedFiles []string) (configData string, err error) {
	// collect drone.yml files
	configData = ""
	cache := map[string]bool{}
	for _, file := range changedFiles {
		dir := file
		for dir != "." {
			dir = path.Join(dir, "..")
			file := path.Join(dir, req.Repo.Config)

			// check if file has already been checked
			_, ok := cache[file]
			if ok {
				continue
			} else {
				cache[file] = true
			}

			// when enabled, only process drone.yml from p.considerFile
			if p.considerFile != "" && !req.ConsiderData.consider(file) {
				continue
			}

			// download file from git
			fileContent, critical, err := p.getDroneConfig(ctx, req, file)
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

// getConfigForTree searches for all or first 'drone.yml' in the repo
func (p *Plugin) getConfigForTree(ctx context.Context, req *request, dir string, depth int) (configData string, err error) {
	if p.considerFile != "" {
		// treats all 'drone.yml' entries in the consider file as the changedFiles
		return p.getConfigForChanges(ctx, req, req.ConsiderData.listRepresentation)
	}

	ls, err := req.Client.GetFileListing(ctx, dir, req.Build.After)
	if err != nil {
		return "", err
	}

	if depth > p.maxDepth {
		logrus.Infof("%s skipping scan of %s, max depth %d reached.", req.UUID, dir, depth)
		return "", nil
	}
	depth += 1

	// check recursively for drone.yml
	configData = ""
	for _, f := range ls {
		var fileContent string
		if f.Type == "dir" {
			fileContent, err = p.getConfigForTree(ctx, req, f.Path, depth)
			if err != nil {
				return "", err
			}
		} else if f.Type == "file" && f.Name == req.Repo.Config {
			var critical bool
			fileContent, critical, err = p.getDroneConfig(ctx, req, f.Path)
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

// getDroneConfig downloads a drone config and validates it
func (p *Plugin) getDroneConfig(ctx context.Context, req *request, file string) (configData string, critical bool, err error) {
	fileContent, err := p.getScmFile(ctx, req, file)
	if err != nil {
		logrus.Debugf("%s skipping: unable to load file: %s %v", req.UUID, file, err)
		return "", false, err
	}

	// validate fileContent, exit early if an error was found
	dc := droneConfig{}
	err = yaml.Unmarshal([]byte(fileContent), &dc)
	if err != nil {
		logrus.Errorf("%s skipping: unable to parse yml file: %s %v", req.UUID, file, err)
		return "", true, err
	}
	if dc.Name == "" || dc.Kind == "" {
		logrus.Errorf("%s skipping: missing 'kind' or 'name' in %s.", req.UUID, file)
		return "", true, err
	}

	logrus.Infof("%s found %s/%s %s", req.UUID, req.Repo.Namespace, req.Repo.Name, file)
	return fileContent, false, nil
}
