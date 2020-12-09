package plugin

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

// ConsiderData holds the considerFile information in both list and map representations
type ConsiderData struct {
	mapRepresentation  map[string]bool
	listRepresentation []string
}

// consider returns true if the path provided matches an entry in the considerFile
func (c *ConsiderData) consider(path string) bool {
	_, exists := c.mapRepresentation[path]
	return exists
}

// newConsiderDataFromRequest returns the ConsiderData which is loaded from the considerFile
func (p *Plugin) newConsiderDataFromRequest(ctx context.Context, req *request) (*ConsiderData, error) {
	cd := new(ConsiderData)
	cd.mapRepresentation = make(map[string]bool)
	cd.listRepresentation = make([]string, 0)

	// bail early without calling the scm provider when there is no considerFile configured
	if p.considerFile == "" {
		return cd, nil
	}

	// download considerFile from github
	fc, err := p.getScmFile(ctx, req, p.considerFile)
	if err != nil {
		logrus.Errorf("%s skipping: %s is not present: %v", req.UUID, p.considerFile, err)
		return cd, err
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
		cd.listRepresentation = append(cd.listRepresentation, v)
		cd.mapRepresentation[v] = true
	}

	return cd, nil
}
