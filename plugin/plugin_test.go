package plugin

import (
	"context"
	"io"
	"os"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/config"
	"github.com/sirupsen/logrus"
)

var noContext = context.Background()

const mockToken = "7535706b694c63526c6e4f5230374243"

// test commit
func TestPlugin(t *testing.T) {
	ts := httptest.NewServer(testMux())
	defer ts.Close()

	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(ts.URL, mockToken, false)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "CORRECT\n\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestConcat(t *testing.T) {
	ts := httptest.NewServer(testMux())
	defer ts.Close()

	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(ts.URL, mockToken, true)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "CORRECT\n\n\n---\nCORRECT2\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestPullRequest(t *testing.T) {
	ts := httptest.NewServer(testMux())
	defer ts.Close()

	req := &config.Request{
		Build: drone.Build{
			Fork: "octocat/dronetest",
			Ref:  "refs/pull/3/head",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(ts.URL, mockToken, true)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "CORRECT2\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func testMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/foosinn/dronetest/compare/2897b31ec3a1b59279a08a8ad54dc360686327f7...8ecad91991d5da985a2a8dd97cc19029dc1c2899",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/compare.json")
			io.Copy(w, f)
		})
	mux.HandleFunc("/repos/foosinn/dronetest/contents/a/b/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/a_b_.drone.yml.json")
			io.Copy(w, f)
		})
	mux.HandleFunc("/repos/foosinn/dronetest/contents/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/.drone.yml.json")
			io.Copy(w, f)
		})
	mux.HandleFunc("/repos/foosinn/dronetest/pulls/3/files",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/pull_3_files.json")
			io.Copy(w, f)
		})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Info(r.URL)
	})
	return mux
}
