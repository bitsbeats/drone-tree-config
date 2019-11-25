package plugin

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// whitelisted determines if the plugin is enabled for the repo slug. decisions are made
// by considering the regex patterns in the regexFile.
//
// returns true (match) or false (no match). false means the repo slug should be bypassed
func (p *Plugin) whitelisted(req *request) bool {
	slug := req.Repo.Slug
	noMatchMsg := fmt.Sprintf("%s no match: %s", req.UUID, slug)
	matchMsg := fmt.Sprintf("%s match: %s", req.UUID, slug)

	// requires a regex file
	if p.whitelistFile == "" {
		// match
		logrus.Info(matchMsg)
		return true
	}

	buf, err := ioutil.ReadFile(p.whitelistFile)
	if err != nil {
		// match
		logrus.Warnf("%s regex file read error: %s", req.UUID, err)
		logrus.Info(matchMsg)
		return true
	}

	lines := strings.Split(string(buf), "\n")

	for _, line := range lines {
		// ignore empty line or line starting with "#" (comment)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		r, err := regexp.Compile(line)
		if err != nil {
			// emit a warning and consider the rest of the lines
			logrus.Warnf("%s %s", req.UUID, err)
			continue
		}

		// the repo is enabled for the plugin, when there is a regex match
		if r.MatchString(slug) {
			// match
			logrus.Info(matchMsg)
			return true
		}
	}

	// no match
	logrus.Info(noMatchMsg)
	return false
}
