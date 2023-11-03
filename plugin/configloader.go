package plugin

import (
	"context"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// KeyOnlyMap is a map with only keys
type KeyOnlyMap map[string]interface{}

// LoadedDroneConfig holds the name and the string content of a `.drone.yml` file
type LoadedDroneConfig struct {
	Name    string
	Content string
}

// DroneConfigCombiner holds multiple LoadedDroneConfigs to combine them
type DroneConfigCombiner struct {
	LoadedConfigs []*LoadedDroneConfig
}

// Append adds a new LoadedDroneConfig
func (dcc *DroneConfigCombiner) Append(ldc *LoadedDroneConfig) {
	dcc.LoadedConfigs = append(dcc.LoadedConfigs, ldc)
}

// Merge merges `left` into current DroneConfigCombiner
func (dcc *DroneConfigCombiner) Merge(left *DroneConfigCombiner) {
	dcc.LoadedConfigs = append(dcc.LoadedConfigs, left.LoadedConfigs...)
}

// ConfigNames loads the names of all Pipelines as map keys
func (dcc *DroneConfigCombiner) ConfigNames(without KeyOnlyMap) []string {
	names := []string{}
	for _, config := range dcc.LoadedConfigs {
		if _, ok := without[config.Name]; !ok {
			names = append(names, config.Name)
		}
	}
	return names
}

// Combine concats all appended configs in to a single string
func (dcc *DroneConfigCombiner) Combine() string {
	combined := ""
	finalize := ""

	// nothing to do
	if len(dcc.LoadedConfigs) == 0 {
		return ""
	}

	// combine all configs except finalize
	for _, ldc := range dcc.LoadedConfigs {
		data := ldc.Content

		if ldc.Name == "finalize" {
			names := dcc.ConfigNames(KeyOnlyMap{"finalize": nil})

			var mdc map[string]interface{}
			_ = yaml.Unmarshal([]byte(data), &mdc)
			logrus.Infof("finalize steps is depending on %+v", names)
			mdc["depends_on"] = names
			dataBytes, _ := yaml.Marshal(mdc)
			data = string(dataBytes)
		}

		data = strings.Trim(data, " \n")
		if data != "" {
			if !strings.HasPrefix(data, "---\n") {
				data = "---\n" + data
			}
			if !strings.HasSuffix(data, "\n") {
				data += "\n"
			}
		}

		// skip finalize as it needs to be added at the end
		if ldc.Name != "finalize" {
			combined += data
		} else {
			finalize = data
		}
	}

	// add finalize at the end
	combined += finalize

	// cleanup
	combined = removeDocEndRegex.ReplaceAllString(combined, "")
	combined = string(dedupRegex.ReplaceAll([]byte(combined), []byte("---")))

	return combined
}

// getConfigForChanges scans a repository for drone configs based on the changed
// files and concats them to a single file.
func (p *Plugin) getConfigForChanges(ctx context.Context, req *request, changedFiles []string) (dcc *DroneConfigCombiner, err error) {
	// collect drone.yml files
	combiner := &DroneConfigCombiner{}
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
			ldc, critical, err := p.getDroneConfig(ctx, req, file)
			if err != nil {
				if critical {
					return nil, err
				}
				continue
			}

			// append
			combiner.Append(ldc)
			if !p.concat {
				logrus.Infof("%s concat is disabled. Using just first .drone.yml.", req.UUID)
				break
			}
		}
	}
	return combiner, nil
}

// getConfigForTree searches for all or first 'drone.yml' in the repo
func (p *Plugin) getConfigForTree(ctx context.Context, req *request, dir string, depth int) (dcc *DroneConfigCombiner, err error) {
	dcc = &DroneConfigCombiner{}

	if p.considerFile != "" {
		// treats all 'drone.yml' entries in the consider file as the changedFiles
		return p.getConfigForChanges(ctx, req, req.ConsiderData.listRepresentation)
	}

	ls, err := req.Client.GetFileListing(ctx, dir, req.Build.After)
	if err != nil {
		return nil, err
	}

	if depth > p.maxDepth {
		logrus.Infof("%s skipping scan of %s, max depth %d reached.", req.UUID, dir, depth)
		return dcc, nil
	}
	depth += 1

	// check recursively for drone.yml
	for _, f := range ls {
		if f.Type == "dir" {
			innerDcc, err := p.getConfigForTree(ctx, req, f.Path, depth)
			if err != nil {
				return nil, err
			}
			dcc.Merge(innerDcc)
		} else if f.Type == "file" && f.Name == req.Repo.Config {
			ldc, critical, err := p.getDroneConfig(ctx, req, f.Path)
			if critical {
				return nil, err
			}
			dcc.Append(ldc)
		}
		if !p.concat {
			logrus.Infof("%s concat is disabled. Using just first .drone.yml.", req.UUID)
			break
		}
	}

	return dcc, nil
}

// getDroneConfig downloads a drone config and validates it
func (p *Plugin) getDroneConfig(
	ctx context.Context, req *request, file string,
) (
	loadedDroneConfig *LoadedDroneConfig, critical bool, err error,
) {
	fileContent, err := p.getScmFile(ctx, req, file)
	if err != nil {
		logrus.Debugf("%s skipping: unable to load file: %s %v", req.UUID, file, err)
		return nil, false, err
	}

	// validate fileContent, exit early if an error was found
	dc := droneConfig{}
	err = yaml.Unmarshal([]byte(fileContent), &dc)
	if err != nil {
		logrus.Errorf("%s skipping: unable to parse yml file: %s %v", req.UUID, file, err)
		return nil, true, err
	}
	if dc.Name == "" || dc.Kind == "" {
		logrus.Errorf("%s skipping: missing 'kind' or 'name' in %s.", req.UUID, file)
		return nil, true, err
	}

	logrus.Infof("%s found %s/%s %s", req.UUID, req.Repo.Namespace, req.Repo.Name, file)
	return &LoadedDroneConfig{
		Name:    dc.Name,
		Content: fileContent,
	}, false, nil
}
