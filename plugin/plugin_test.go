package plugin

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var noContext = context.Background()

const mockToken = "7535706b694c63526c6e4f5230374243"

var ts *httptest.Server

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)

	ts = httptest.NewServer(testMux())
	defer ts.Close()
	os.Exit(m.Run())
}

// test commit
func TestPlugin(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Branch:    "master",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithFallback(true),
		WithMaxDepth(2),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestPluginWithConsider(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Branch:    "master",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithFallback(true),
		WithMaxDepth(2),
		WithConsiderFile(".drone-consider"),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestPluginWithConsiderRepoConfig(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Branch:    "master",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithFallback(true),
		WithMaxDepth(2),
		WithConsiderFile(".no-existent-drone-consider"),
		WithConsiderRepoConfig(true),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestConcat(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Branch:    "master",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithConcat(true),
		WithFallback(true),
		WithMaxDepth(2),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestConcatWithConsider(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Branch:    "master",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithConcat(true),
		WithFallback(true),
		WithMaxDepth(2),
		WithConsiderFile(".drone-consider"),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestPullRequest(t *testing.T) {
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
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithConcat(true),
		WithFallback(true),
		WithMaxDepth(2),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestPullRequestWithConsider(t *testing.T) {
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
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithConcat(true),
		WithFallback(true),
		WithMaxDepth(2),
		WithConsiderFile(".drone-consider"),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestCron(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			After:   "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Trigger: "@cron",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithFallback(true),
		WithMaxDepth(2),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n", droneConfig.Data; want != got {
		t.Errorf("Want %q got %q", want, got)
	}
}

func TestCache(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			After:   "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Trigger: "@cron",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}

	// used to directly retrieve the cacheEntry, for verification
	r := &request{
		Request: req,
		UUID:    uuid.New(),
	}
	ck := newCacheKey(r)

	p := &Plugin{
		server:      ts.URL,
		gitHubToken: mockToken,
		concat:      true,
		maxDepth:    2,
		cacheTTL:    time.Minute * 1,
		cache:       &configCache{},
	}

	// test cache hit
	for i := 0; i < 2; i++ {
		droneConfig, err := p.Find(noContext, req)
		if err != nil {
			t.Error(err)
			return
		}

		if entry, ok := p.cache.retrieve(r.UUID, ck); ok {
			if want := droneConfig.Data; entry.config != want {
				t.Errorf("Want %q got %q", droneConfig.Data, entry.config)
			}
			if want := err; entry.error != want {
				t.Errorf("Want %q got %q", want, entry.error)
			}
		} else {
			t.Error("entry not in cache")
		}
	}

	// test cache expire
	p.cache.expire(ck)
	entry, ok := p.cache.retrieve(r.UUID, ck)
	if entry != nil || ok {
		t.Error("entry still in cache")
	}
}

func TestMatchEnable(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Branch:    "master",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}

	type scenario struct {
		file string
		want *drone.Config
	}

	matchWant := &drone.Config{
		Data: "---\nkind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n",
	}

	scenarios := map[string]scenario{
		// matches all repos for this plugin
		"MatchAll": {
			file: "testdata/regex/matchall",
			want: matchWant,
		},
		// matches no repos for this plugin
		"MatchNone": {
			file: "testdata/regex/matchnone",
			want: nil,
		},
		// matches all repos for this plugin. specified file with match rules does not exist
		"FileError": {
			file: "no_such_file",
			want: matchWant,
		},
		// matches all repos for this plugin. there is no file with match rules
		"NoFileSpecified": {
			file: "",
			want: matchWant,
		},
	}

	for name, s := range scenarios {
		t.Run(name, func(t *testing.T) {
			plugin := New(
				WithServer(ts.URL),
				WithGithubToken(mockToken),
				WithFallback(true),
				WithMaxDepth(2),
				WithAllowListFile(s.file),
			)
			droneConfig, err := plugin.Find(noContext, req)
			if err != nil {
				t.Error(err)
				return
			}

			if droneConfig != nil {
				if got := droneConfig.Data; s.want.Data != got {
					t.Errorf("Want %q got %q", s.want, got)
				}
			} else {
				if got := droneConfig; s.want != got {
					t.Errorf("Want %q got %q", s.want, got)
				}
			}
		})
	}
}

func TestCronWithConsider(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			After:   "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Trigger: "@cron",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithConcat(true),
		WithFallback(true),
		WithMaxDepth(2),
		WithConsiderFile(".drone-consider"),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n---\nkind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n", droneConfig.Data; want != got {
		t.Errorf("Want\n  %q\ngot\n  %q", want, got)
	}
}
func TestCronConcat(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			After:   "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Trigger: "@cron",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithConcat(true),
		WithFallback(true),
		WithMaxDepth(2),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n---\nkind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n", droneConfig.Data; want != got {
		t.Errorf("Want\n  %q\ngot\n  %q", want, got)
	}
}
func TestAlwaysRunAll(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithAlwaysRunAll(true),
		WithMaxDepth(2),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n", droneConfig.Data; want != got {
		t.Errorf("Want\n  %q\ngot\n  %q", want, got)
	}
}
func TestAlwaysRunAllConcat(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.yml",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithConcat(true),
		WithAlwaysRunAll(true),
		WithMaxDepth(2),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "---\nkind: pipeline\nname: default\n\nsteps:\n- name: frontend\n  image: node\n  commands:\n  - npm install\n  - npm test\n\n- name: backend\n  image: golang\n  commands:\n  - go build\n  - go test\n---\nkind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n", droneConfig.Data; want != got {
		t.Errorf("Want\n  %q\ngot\n  %q", want, got)
	}
}

func TestStarlark(t *testing.T) {
	req := &config.Request{
		Build: drone.Build{
			Before: "2897b31ec3a1b59279a08a8ad54dc360686327f7",
			After:  "8ecad91991d5da985a2a8dd97cc19029dc1c2899",
			Source: "master",
		},
		Repo: drone.Repo{
			Namespace: "foosinn",
			Name:      "dronetest",
			Branch:    "master",
			Slug:      "foosinn/dronetest",
			Config:    ".drone.star",
		},
	}
	plugin := New(
		WithServer(ts.URL),
		WithGithubToken(mockToken),
		WithFallback(false),
		WithMaxDepth(2),
	)
	droneConfig, err := plugin.Find(noContext, req)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := (*drone.Config)(nil), droneConfig; want != got {
		t.Errorf("Want\n  %q\ngot\n  %q", want, got)
	}
}

func testMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/root.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/compare/2897b31ec3a1b59279a08a8ad54dc360686327f7...8ecad91991d5da985a2a8dd97cc19029dc1c2899",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/compare.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/a/b/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/a_b_.drone.yml.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/.drone.yml.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/.drone-consider",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/.drone-consider.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/pulls/3/files",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/pull_3_files.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/afolder/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/afolder_.drone.yml.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/afolder",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/afolder.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/afolder/abfolder",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("testdata/github/afolder_abfolder.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Errorf("Url not found: %s", r.URL)
	})
	return mux
}
