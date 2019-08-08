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

const mockClientId = "abra"
const mockSecret = "c4d4br4"

func TestBitBucket_GetFileContents(t *testing.T) {
	ts := httptest.NewServer(testMuxBitBucket())
	defer ts.Close()
	client, err := createBitBucketClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	BaseTest_GetFileContents(t, client)
}

func TestBitBucket_ChangedFilesInDiff(t *testing.T) {
	ts := httptest.NewServer(testMuxBitBucket())
	defer ts.Close()
	client, err := createBitBucketClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	BaseTest_ChangedFilesInDiff(t, client)
}

func TestBitBucket_ChangedFilesInPullRequest(t *testing.T) {
	ts := httptest.NewServer(testMuxBitBucket())
	defer ts.Close()
	client, err := createBitBucketClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	BaseTest_ChangedFilesInPullRequest(t, client)
}

func TestBitBucket_GetFileListing(t *testing.T) {
	ts := httptest.NewServer(testMuxBitBucket())
	defer ts.Close()
	client, err := createBitBucketClient(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	BaseTest_GetFileListing(t, client)
}

func createBitBucketClient(server string) (ScmClient, error) {
	repo := drone.Repo{
		Namespace: "foosinn",
		Name:      "dronetest",
		Slug:      "foosinn/dronetest",
	}
	return NewBitBucketClient(uuid.New(), server, server, mockClientId, mockSecret, repo)
}

func testMuxBitBucket() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/site/oauth2/access_token",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket/token.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/2.0/repositories/foosinn/dronetest/diffstat/2897b31ec3a1b59279a08a8ad54dc360686327f7..8ecad91991d5da985a2a8dd97cc19029dc1c2899",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket/compare.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/2.0/repositories/foosinn/dronetest/pullrequests/3/diffstat",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket/pull_3_files.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/2.0/repositories/foosinn/dronetest/src/8ecad91991d5da985a2a8dd97cc19029dc1c2899/afolder/.drone.yml",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket/afolder_.drone.yml")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/2.0/repositories/foosinn/dronetest/src/8ecad91991d5da985a2a8dd97cc19029dc1c2899/afolder/",
		func(w http.ResponseWriter, r *http.Request) {
			f, _ := os.Open("../testdata/bitbucket/afolder.json")
			_, _ = io.Copy(w, f)
		})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Errorf("Url not found: %s", r.URL)
	})
	return mux
}
