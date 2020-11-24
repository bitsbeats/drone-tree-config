package scm_clients

import (
	"github.com/drone/drone-go/drone"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var mockBitBucketToken = "bar"

func TestBitBucketServer_GetFileContents(t *testing.T) {
	ts := httptest.NewServer(testMuxBitBucketServer())
	defer ts.Close()
	client, err := createBitBucketServerClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	BaseTest_GetFileContents(t, client)
}

func TestBitBucketServer_ChangedFilesInDiff(t *testing.T) {
	ts := httptest.NewServer(testMuxBitBucketServer())
	defer ts.Close()
	client, err := createBitBucketServerClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	BaseTest_ChangedFilesInDiff(t, client)
}

func TestBitBucketServer_ChangedFilesInPullRequest(t *testing.T) {
	ts := httptest.NewServer(testMuxBitBucketServer())
	defer ts.Close()
	client, err := createBitBucketServerClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	BaseTest_ChangedFilesInPullRequest(t, client)
}

func TestBitBucketServer_GetFileListing(t *testing.T) {
	ts := httptest.NewServer(testMuxBitBucketServer())
	defer ts.Close()
	client, err := createBitBucketServerClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	BaseTest_GetFileListing(t, client)
}

func createBitBucketServerClient(server string) (ScmClient, error) {
	repo := drone.Repo{
		Namespace: "foosinn",
		Name:      "dronetest",
		Slug:      "foosinn/dronetest",
	}
	return NewBitBucketServerClient(uuid.New(), server, mockBitBucketToken, repo)
}

func testMuxBitBucketServer() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/1.0/projects/foosinn/repos/dronetest/commits/8ecad91991d5da985a2a8dd97cc19029dc1c2899/diff",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket_server/compare.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/1.0/projects/foosinn/repos/dronetest/pull-requests/3/diff",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket_server/compare_pr.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/1.0/projects/foosinn/repos/dronetest/raw/afolder/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket_server/drone-test.yml")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/api/1.0/projects/foosinn/repos/dronetest/files/afolder",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket_server/afolder.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Errorf("Url not found: %s", r.URL)
	})
	return mux
}
