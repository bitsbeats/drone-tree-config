package main

import (
	"net/http"

	"github.com/bitsbeats/drone-tree-config/plugin"

	"github.com/drone/drone-go/plugin/config"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type (
	spec struct {
		Concat  bool   `envconfig:"PLUGIN_CONCAT" default:false`
		Debug   bool   `envconfig:"PLUGIN_DEBUG"`
		Address string `envconfig:"PLUGIN_ADDRESS" default:":3000"`
		Secret  string `envconfig:"PLUGIN_SECRET"`
		Token   string `envconfig:"GITHUB_TOKEN"`
		Server  string `envconfig:"GITHUB_SERVER"`
	}
)

func main() {
	spec := new(spec)
	if err := envconfig.Process("", spec); err != nil {
		logrus.Fatal(err)
	}

	if spec.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if spec.Secret == "" {
		logrus.Fatalln("missing secret key")
	}
	if spec.Token == "" {
		logrus.Warnln("missing github token")
	}
	if spec.Address == "" {
		spec.Address = ":3000"
	}

	handler := config.Handler(
		plugin.New(
			spec.Server,
			spec.Token,
			spec.Concat,
		),
		spec.Secret,
		logrus.StandardLogger(),
	)

	logrus.Infof("server listening on address %s", spec.Address)

	http.Handle("/", handler)
	logrus.Fatal(http.ListenAndServe(spec.Address, nil))
}
