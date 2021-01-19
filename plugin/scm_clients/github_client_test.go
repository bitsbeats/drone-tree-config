package scm_clients

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/drone/drone-go/drone"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const mockGithubToken = "7535706b694c63526c6e4f5230374243"

var ts *httptest.Server

func TestMain(m *testing.M) {
	ts = httptest.NewServer(testMuxGithub())
	defer ts.Close()
	os.Exit(m.Run())
}

func TestGithubClient_GetFileContents(t *testing.T) {
	client, err := createGithubClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	BaseTest_GetFileContents(t, client)
}

func TestGithubClient_ChangedFilesInDiff(t *testing.T) {
	client, err := createGithubClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	BaseTest_ChangedFilesInDiff(t, client)
}

func TestGithubClient_ChangedFilesInPullRequest(t *testing.T) {
	client, err := createGithubClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	BaseTest_ChangedFilesInPullRequest(t, client)
}

func TestGithubClient_ChangedFilesInPullRequest_Paginated(t *testing.T) {
	ts := httptest.NewServer(testMuxGithub())
	defer ts.Close()
	client, err := createGithubClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	actualFiles, err := client.ChangedFilesInPullRequest(noContext, 4)
	if err != nil {
		t.Error(err)
		return
	}

	expectedFiles := []string{
		"e/f/g/h/.drone.yml",
		"e/f/g/h/.drone.yml",
	}

	if want, got := expectedFiles, actualFiles; !reflect.DeepEqual(want, got) {
		t.Errorf("Test failed:\n  want %q\n   got %q", want, got)
	}
}

func TestGithubClient_GetFileListing(t *testing.T) {
	client, err := createGithubClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	BaseTest_GetFileListing(t, client)
}

func createGithubClient(server string) (ScmClient, error) {
	someUuid := uuid.New()
	repo := drone.Repo{
		Namespace: "foosinn",
		Name:      "dronetest",
		Slug:      "foosinn/dronetest",
	}
	return NewGitHubClient(noContext, someUuid, server, mockGithubToken, repo)
}

func testMuxGithub() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/github/root.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/compare/2897b31ec3a1b59279a08a8ad54dc360686327f7...8ecad91991d5da985a2a8dd97cc19029dc1c2899",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/github/compare.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/a/b/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/github/a_b_.drone.yml.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/github/.drone.yml.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/pulls/3/files",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/github/pull_3_files.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/pulls/4/files",
		func(w http.ResponseWriter, r *http.Request) {
			// simulate a paginated response
			if r.FormValue("page") == "" {
				next := fmt.Sprintf("<%s?page=2>; rel=\"next\"", r.URL.String())
				last := fmt.Sprintf("<%s?page=2>; rel=\"last\"", r.URL.String())
				w.Header().Add("Link", fmt.Sprintf("%s, %s", next, last))
			}
			f, _ := os.Open("../testdata/github/pull_3_files.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/afolder/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/github/afolder_.drone.yml.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/v3/repos/foosinn/dronetest/contents/afolder",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/github/afolder.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Errorf("Url not found: %s", r.URL)
	})
	return mux
}
