package plugin

import (
	"context"
	"path"
	"strings"

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

// getConsiderFile returns the 'drone.yml' entries in a consider file as a string slice
func (p *Plugin) getConsiderFile(ctx context.Context, req *request) ([]string, error) {
	toReturn := make([]string, 0)

	// download considerFile from github
	fc, err := p.getScmFile(ctx, req, p.considerFile)
	if err != nil {
		logrus.Errorf("%s skipping: %s is not present: %v", req.UUID, p.considerFile, err)
		return toReturn, err
	}

	// collect drone.yml files
	for _, v := range strings.Split(fc, "\n") {
		// skip empty lines and comments
		if strings.TrimSpace(v) == "" || strings.HasPrefix(v, "#") {
			continue
		}
		// skip lines which do not contain a 'drone.yml' reference
		if !strings.HasSuffix(v, req.Repo.Config) {
			logrus.Warnf("%s skipping invalid reference to %s in %s", req.UUID, v, p.considerFile)
			continue
		}
		toReturn = append(toReturn, v)
	}

	return toReturn, nil
}

// getConfigForChangesUsingConsider loads 'drone.yml' from the consider file based on the changed files.
// Note: this call does not fail if there are invalid entries in a consider file
func (p *Plugin) getConfigForChangesUsingConsider(ctx context.Context, req *request, changedFiles []string) (string, error) {
	configData := ""
	consider := map[string]bool{}
	cache := map[string]bool{}

	considerEntries, err := p.getConsiderFile(ctx, req)
	if err != nil {
		return "", err
	}
	// convert to a map for O(1) lookup
	for _, v := range considerEntries {
		consider[v] = true
	}

	for _, file := range changedFiles {
		dir := file
		for dir != "." {
			dir = path.Join(dir, "..")
			file := path.Join(dir, req.Repo.Config)

			// check if file has already been checked
			if _, ok := cache[file]; ok {
				continue
			}
			cache[file] = true

			// look for file in consider map
			if _, exists := consider[file]; exists {
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
	}
	return configData, nil
}

// getConfigForTree searches for all or first 'drone.yml' in the repo
func (p *Plugin) getConfigForTree(ctx context.Context, req *request, dir string, depth int) (configData string, err error) {
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

// getConfigForTreeUsingConsider loads all 'drone.yml' which are identified in the consider file.
func (p *Plugin) getConfigForTreeUsingConsider(ctx context.Context, req *request) (string, error) {
	configData := ""
	cache := map[string]bool{}

	consider, err := p.getConsiderFile(ctx, req)
	if err != nil {
		return "", err
	}

	// collect drone.yml files
	for _, v := range consider {
		if _, ok := cache[v]; ok {
			continue
		}
		cache[v] = true

		// download file from github
		fc, critical, err := p.getDroneConfig(ctx, req, v)
		if err != nil {
			if critical {
				return "", err
			}
			continue
		}

		// append
		configData = p.droneConfigAppend(configData, fc)
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
