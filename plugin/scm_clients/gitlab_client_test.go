package scm_clients

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/drone/drone-go/drone"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const mockGitlabToken = "7535706b694c63526c6e4f5230374243"

func TestGitlabClient_GetFileContents(t *testing.T) {
	ts := httptest.NewServer(testMuxGitlab())
	defer ts.Close()
	client, err := createGitlabClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	BaseTest_GetFileContents(t, client)
}

func TestGitlabClient_ChangedFilesInDiff(t *testing.T) {
	ts := httptest.NewServer(testMuxGitlab())
	defer ts.Close()
	client, err := createGitlabClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	BaseTest_ChangedFilesInDiff(t, client)
}

func TestGitlabClient_ChangedFilesInPullRequest(t *testing.T) {
	ts := httptest.NewServer(testMuxGitlab())
	defer ts.Close()
	client, err := createGitlabClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	BaseTest_ChangedFilesInPullRequest(t, client)
}

func TestGitlabClient_GetFileListing(t *testing.T) {
	ts := httptest.NewServer(testMuxGitlab())
	defer ts.Close()
	client, err := createGitlabClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	BaseTest_GetFileListing(t, client)
}

func createGitlabClient(server string) (ScmClient, error) {
	someUUID := uuid.New()
	repo := drone.Repo{
		UID:       "1234",
		Namespace: "foosinn",
		Name:      "dronetest",
		Slug:      "foosinn/dronetest",
	}
	return NewGitLabClient(noContext, someUUID, server, mockGitlabToken, repo)
}

func testMuxGitlab() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/1234/repository/tree",
		func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("path") == "afolder" && r.FormValue("ref") == "8ecad91991d5da985a2a8dd97cc19029dc1c2899" {
				f, _ := os.Open("../testdata/gitlab/afolder.json")
				_, _ = io.Copy(w, f)
			}
		})
	mux.HandleFunc("/api/v4/projects/1234/repository/compare",
		func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("from") == "2897b31ec3a1b59279a08a8ad54dc360686327f7" && r.FormValue("to") == "8ecad91991d5da985a2a8dd97cc19029dc1c2899" {
				f, _ := os.Open("../testdata/gitlab/compare.json")
				_, _ = io.Copy(w, f)
			}
		})
	mux.HandleFunc("/api/v4/projects/1234/merge_requests/3/changes",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/gitlab/pull_3_files.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v4/projects/1234/repository/files/",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.RawPath == "/api/v4/projects/1234/repository/files/.drone.yml" {
				f, _ := os.Open("../testdata/gitlab/.drone.yml.json")
				_, _ = io.Copy(w, f)
			}
			if r.URL.RawPath == "/api/v4/projects/1234/repository/files/a%2Fb%2F.drone.yml" {
				f, _ := os.Open("../testdata/gitlab/a_b_.drone.yml.json")
				_, _ = io.Copy(w, f)
			}
			if r.URL.RawPath == "/api/v4/projects/1234/repository/files/afolder%2F.drone.yml" && r.FormValue("ref") == "8ecad91991d5da985a2a8dd97cc19029dc1c2899" {
				f, _ := os.Open("../testdata/gitlab/afolder_.drone.yml.json")
				_, _ = io.Copy(w, f)
			}
		})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Errorf("Url not found: %s", r.URL)
	})
	return mux
}
